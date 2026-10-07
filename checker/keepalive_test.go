package checker

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestKeepaliveChecks(t *testing.T) {
	for _, method := range []string{"ip", "status", "download"} {
		for _, mode := range []string{"default", "keepalive"} {
			t.Run(method+"/"+mode, func(t *testing.T) {
				var requests, connections atomic.Int32
				server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if requests.Add(1) == 1 {
						time.Sleep(150 * time.Millisecond)
					}
					fmt.Fprint(w, "203.0.113.1")
				}))
				server.Config.ConnState = func(_ net.Conn, state http.ConnState) {
					if state == http.StateNew {
						connections.Add(1)
					}
				}
				server.StartTLS()
				defer server.Close()
				client := server.Client()
				client.Timeout = 2 * time.Second
				defer client.CloseIdleConnections()
				pc := &ProxyChecker{
					pingMode: mode, ipCheck: server.URL, genMethodURL: server.URL,
					downloadURL: server.URL, downloadTimeout: 2, downloadMinSize: 1,
				}
				check := pc.checkByIP
				switch method {
				case "status":
					check = pc.checkByGen
				case "download":
					check = pc.checkByDownload
				}
				start := time.Now()
				ok, message, latency, err := check(client)
				elapsed := time.Since(start)
				if err != nil || !ok {
					t.Fatalf("check: ok=%v message=%s err=%v", ok, message, err)
				}
				wantRequests := int32(1)
				if mode == "keepalive" {
					wantRequests = 2
					if elapsed-latency < 150*time.Millisecond {
						t.Fatalf("warm-up included in latency: elapsed=%s latency=%s", elapsed, latency)
					}
				} else if latency < 150*time.Millisecond {
					t.Fatalf("default latency excludes first request delay: %s", latency)
				}
				if requests.Load() != wantRequests || connections.Load() != 1 {
					t.Fatalf("requests=%d connections=%d", requests.Load(), connections.Load())
				}
			})
		}
	}
}

func TestKeepaliveRejectsNewConnection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Connection", "close")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	client := server.Client()
	defer client.CloseIdleConnections()
	pc := &ProxyChecker{pingMode: "keepalive", genMethodURL: server.URL}
	ok, _, latency, err := pc.checkByGen(client)
	if ok || latency != 0 || err == nil || !strings.Contains(err.Error(), "reused connection") {
		t.Fatalf("ok=%v latency=%s err=%v", ok, latency, err)
	}
}

func TestKeepaliveWarmupReadFailure(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.Header().Set("Content-Length", "100")
		fmt.Fprint(w, "short")
	}))
	defer server.Close()
	client := server.Client()
	defer client.CloseIdleConnections()
	pc := &ProxyChecker{pingMode: "keepalive", genMethodURL: server.URL}
	ok, _, _, err := pc.checkByGen(client)
	if ok || err == nil || !strings.Contains(err.Error(), "warm-up body") || requests.Load() != 1 {
		t.Fatalf("ok=%v requests=%d err=%v", ok, requests.Load(), err)
	}
}

func TestKeepaliveUsesMeasuredResponseStatus(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if requests.Add(1) == 1 {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()
	client := server.Client()
	defer client.CloseIdleConnections()
	pc := &ProxyChecker{pingMode: "keepalive", genMethodURL: server.URL}
	ok, message, _, err := pc.checkByGen(client)
	if ok || err != nil || message != "Status: 503" || requests.Load() != 2 {
		t.Fatalf("ok=%v message=%s requests=%d err=%v", ok, message, requests.Load(), err)
	}
}
