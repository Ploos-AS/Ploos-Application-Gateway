#!/bin/sh
set -eu

TMP=/tmp/pag-dns-shutdown
SOCK="$TMP/pag-dns.sock"
rm -rf "$TMP"; mkdir -p "$TMP"
trap 'kill ${PID:-} 2>/dev/null || true; rm -rf "$TMP"' EXIT INT TERM

go build -o "$TMP/pag-dns" ./cmd/pag-dns
"$TMP/pag-dns" -listen 127.0.0.1:55355 -upstream 127.0.0.1:9 -control "$SOCK" 2>"$TMP/log" &
PID=$!

i=0
while [ ! -S "$SOCK" ]; do
  i=$((i+1))
  [ "$i" -lt 50 ] || { echo "FAIL: control socket not created" >&2; exit 1; }
  sleep .05
done

kill -TERM "$PID"
i=0
while kill -0 "$PID" 2>/dev/null; do
  i=$((i+1))
  [ "$i" -lt 50 ] || { echo "FAIL: pag-dns did not exit after SIGTERM" >&2; exit 1; }
  sleep .05
done
wait "$PID"
PID=

[ ! -e "$SOCK" ] || { echo "FAIL: control socket remains after shutdown" >&2; exit 1; }
echo "PASS: pag-dns SIGTERM shutdown removes control socket"
