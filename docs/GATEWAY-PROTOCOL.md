# PAG gateway control protocol — M1 prototype

This protocol is deliberately minimal and **not yet a stable ABI**.

## Transport

Each separately installed gateway exposes a Unix domain socket configured in PAG. Core never dynamically loads protocol parser code.

## Hello / health

Core sends one newline-delimited JSON request:

```json
{"op":"hello"}
```

The gateway responds:

```json
{"id":"pag-dns","version":"0.1.0","healthy":true}
```

Core verifies that the reported identity matches policy/configuration and that the gateway reports healthy before registration.

Connection failure, malformed responses, identity mismatch, or unhealthy state fail closed.

## Evolution

M1 will use this small protocol to prove process separation, discovery, health checks and lifecycle behavior. A richer versioned protocol can replace it after the architecture has been exercised without putting protocol implementations into `pag-core`.
