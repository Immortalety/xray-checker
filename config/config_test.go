package config

import (
	"os"
	"testing"

	"github.com/alecthomas/kong"
)

func TestPingMode(t *testing.T) {
	for _, tc := range []struct {
		name      string
		args      []string
		env       string
		want      string
		wantError bool
	}{
		{name: "default", want: "default"},
		{name: "flag", args: []string{"--proxy-ping-mode=keepalive"}, want: "keepalive"},
		{name: "environment", env: "keepalive", want: "keepalive"},
		{name: "invalid", args: []string{"--proxy-ping-mode=invalid"}, wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("PROXY_PING_MODE", tc.env)
			if tc.env == "" {
				if err := os.Unsetenv("PROXY_PING_MODE"); err != nil {
					t.Fatal(err)
				}
			}
			var cli CLI
			parser, err := kong.New(&cli)
			if err != nil {
				t.Fatal(err)
			}
			args := append([]string{"--subscription-url=https://example.com/sub"}, tc.args...)
			_, err = parser.Parse(args)
			if (err != nil) != tc.wantError {
				t.Fatalf("parse error=%v wantError=%v", err, tc.wantError)
			}
			if !tc.wantError && cli.Proxy.PingMode != tc.want {
				t.Fatalf("mode=%q want=%q", cli.Proxy.PingMode, tc.want)
			}
		})
	}
}
