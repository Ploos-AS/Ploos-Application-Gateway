#!/bin/sh
set -eu

TMP=/tmp/pag-dns-limits
SOCK=/tmp/pag-dns-limits.sock
rm -rf "$TMP" "$SOCK"; mkdir -p "$TMP"
trap 'kill ${PID:-} 2>/dev/null || true; rm -rf "$TMP" "$SOCK"' EXIT INT TERM

go build -o "$TMP/pag-dns" ./cmd/pag-dns

# No upstream is needed: the rate limiter acts before upstream I/O.
"$TMP/pag-dns" -listen 127.0.0.1:55353 -upstream 127.0.0.1:9 \
  -control "$SOCK" -rate 0.0001 -burst 1 -max-tcp-per-client 1 \
  2>"$TMP/audit.log" &
PID=$!
sleep .3

python3 - <<'PY' >"$TMP/q.bin"
import sys
q=bytearray(b'\x31\x01\x01\x00\x00\x01\x00\x00\x00\x00\x00\x00')
q.extend(b'\x01a\x00\x00\x01\x00\x01')
sys.stdout.buffer.write(q)
PY

# First UDP query consumes the only burst token; the second must be denied.
cat "$TMP/q.bin" | nc -u -w 1 127.0.0.1 55353 >/dev/null 2>&1 || true
cat "$TMP/q.bin" | nc -u -w 1 127.0.0.1 55353 >/dev/null 2>&1 || true
sleep .1
grep -q '"transport":"udp".*"reason":"rate_limit"' "$TMP/audit.log"
echo "PASS: UDP rate limit enforced and audited"

# Restart so TCP gets a fresh limiter. Hold one TCP connection open, then the
# second connection must be rejected by the per-client concurrency limit.
kill "$PID"; wait "$PID" 2>/dev/null || true
: >"$TMP/audit.log"
"$TMP/pag-dns" -listen 127.0.0.1:55353 -upstream 127.0.0.1:9 \
  -control "$SOCK" -rate 100 -burst 100 -max-tcp-per-client 1 \
  2>"$TMP/audit.log" &
PID=$!
sleep .3

python3 - <<'PY' &
import socket,time
s=socket.create_connection(("127.0.0.1",55353))
time.sleep(2)
s.close()
PY
HOLD=$!
sleep .2
python3 - <<'PY'
import socket
s=socket.create_connection(("127.0.0.1",55353))
try:
    s.sendall(b"\x00\x0c"+b"\x00"*12)
except OSError:
    pass
s.close()
PY
sleep .1
grep -q '"transport":"tcp".*"reason":"concurrency_limit"' "$TMP/audit.log"
wait "$HOLD"
echo "PASS: TCP concurrency limit enforced and audited"


# Global TCP cap must protect the gateway even when per-client capacity is higher.
kill "$PID"; wait "$PID" 2>/dev/null || true
: >"$TMP/audit.log"
"$TMP/pag-dns" -listen 127.0.0.1:55353 -upstream 127.0.0.1:9 \
  -control "$SOCK" -rate 100 -burst 100 -max-tcp-per-client 16 -max-tcp-global 1 \
  2>"$TMP/audit.log" &
PID=$!
sleep .3
python3 - <<'PY' &
import socket,time
s=socket.create_connection(("127.0.0.1",55353))
time.sleep(2)
s.close()
PY
HOLD=$!
sleep .2
python3 - <<'PY'
import socket
s=socket.create_connection(("127.0.0.1",55353))
try: s.sendall(b"\x00\x0c"+b"\x00"*12)
except OSError: pass
s.close()
PY
sleep .1
grep -q '"transport":"tcp".*"reason":"global_concurrency_limit"' "$TMP/audit.log"
wait "$HOLD"
echo "PASS: global TCP concurrency limit enforced and audited"
