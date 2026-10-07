---
title: Check Methods
description: Check methods options and examples
---

## Check Methods

Xray Checker supports three methods for verifying proxy functionality.

:::note[Latency Measurement]
All check methods measure latency using **TTFB (Time To First Byte)** - the time from request start until the first byte of the response is received. This provides consistent and accurate latency measurements across all methods.
:::

### IP Check Method (Default)

```bash
--proxy-check-method=ip
```

This method:

1. Gets current IP without proxy
2. Connects through proxy
3. Gets IP through proxy
4. Compares IPs to verify proxy is working

Benefits:

- More reliable verification
- Confirms actual proxy functionality
- Detects transparent proxies

Configuration:

```bash
PROXY_CHECK_METHOD=ip
PROXY_IP_CHECK_URL=https://api.ipify.org?format=text
PROXY_TIMEOUT=30
```

### Status Check Method

```bash
--proxy-check-method=status
```

This method:

1. Connects through proxy
2. Requests specified URL
3. Verifies response status code

Benefits:

- Faster verification
- Lower bandwidth usage
- Works with restrictive firewalls

Configuration:

```bash
PROXY_CHECK_METHOD=status
PROXY_STATUS_CHECK_URL=http://cp.cloudflare.com/generate_204
PROXY_TIMEOUT=30
```

### Download Check Method

```bash
--proxy-check-method=download
```

This method:

1. Connects through proxy
2. Downloads a file from specified URL
3. Verifies minimum bytes received

Benefits:

- Tests actual data transfer capability
- Detects proxies with connection issues
- Useful for verifying streaming/download capability

Configuration:

```bash
PROXY_CHECK_METHOD=download
PROXY_DOWNLOAD_URL=https://proof.ovh.net/files/1Mb.dat
PROXY_DOWNLOAD_TIMEOUT=60
PROXY_DOWNLOAD_MIN_SIZE=51200
```

:::tip[Choosing the Right Method]
- Use **ip** for most reliable verification (default)
- Use **status** for fast checks with minimal bandwidth
- Use **download** when you need to verify data transfer capability
:::

## Latency measurement mode

Set `PROXY_PING_MODE=keepalive` (CLI: `--proxy-ping-mode=keepalive`) to exclude connection setup from latency, following [Happ's keepalive mode](https://docs.happ.info/main/dev-docs/app-management#proxy-ping-mode). The default, `default`, keeps the existing single-request measurement including connection setup.

For each proxy check, the checker sends a warm-up GET, reads its entire response, then measures time to the first response byte of a second GET using the same HTTP transport and established connection (including TLS for HTTPS URLs). This applies to `ip`, `status`, and `download`; the second response determines the check result. Idle connections are closed after each check.

The endpoint must support persistent connections. If the measured request opens a new connection, the check fails instead of reporting a latency that includes a new handshake. Each request has its own configured timeout. Warm-up downloads read the entire file, increasing traffic and potentially requiring a larger download timeout; prefer a small response endpoint for latency checks.

Example: `--proxy-check-method=status --proxy-ping-mode=keepalive --proxy-status-check-url=https://cp.cloudflare.com/generate_204`.