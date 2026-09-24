#!/bin/sh
set -eu

TMP=/tmp/pag-dns-shutdown
SOCK="$TMP/pag-dns.sock"
rm -rf "$TMP"; mkdir -p "$TMP"
trap 'kill ${CLIENT_PID:-} ${STALL_PID:-} ${PID:-} 2>/dev/null || true; rm -rf "$TMP"' EXIT INT TERM

go build -o "$TMP/pag-dns" ./cmd/pag-dns
"$TMP/pag-dns" -listen 127.0.0.1:55355 -upstream 127.0.0.1:9 -control "$SOCK" 2>"$TMP/log" &
PID=$!

i=0
while [ ! -S "$SOCK" ]; do
  i=$((i+1))
  [ "$i" -lt 50 ] || { echo "FAIL: control socket not created" >&2; exit 1; }
  sleep .05
done

# Hold an accepted TCP DNS client open so shutdown must drain a live worker.
python3 - <<'PY' &
import socket, time
s = socket.create_connection(("127.0.0.1", 55355))
time.sleep(10)
s.close()
PY
CLIENT_PID=$!
sleep .1

kill -TERM "$PID"
i=0
while kill -0 "$PID" 2>/dev/null; do
  i=$((i+1))
  [ "$i" -lt 160 ] || { echo "FAIL: pag-dns did not exit after bounded worker drain" >&2; exit 1; }
  sleep .05
done
wait "$PID"
PID=
kill "$CLIENT_PID" 2>/dev/null || true
wait "$CLIENT_PID" 2>/dev/null || true
CLIENT_PID=

grep -q "shutdown drain timed out with active DNS workers" "$TMP/log" || {
  echo "FAIL: active worker did not exercise bounded drain timeout" >&2
  cat "$TMP/log" >&2
  exit 1
}

[ ! -e "$SOCK" ] || { echo "FAIL: control socket remains after shutdown" >&2; exit 1; }
echo "PASS: pag-dns bounds active-worker drain and removes control socket"
