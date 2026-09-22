#!/bin/sh
set -eu

TMP=/tmp/pag-integration
SOCK=/tmp/pag-test.sock
rm -rf "$TMP" "$SOCK"
mkdir -p "$TMP"

go build -o "$TMP/pag" ./cmd/pag
go build -o "$TMP/pag-test-gateway" ./cmd/pag-test-gateway

"$TMP/pag-test-gateway" -socket "$SOCK" -id pag-test >"$TMP/gateway.log" 2>&1 &
PID=$!
trap 'kill "$PID" 2>/dev/null || true; rm -f "$SOCK"' EXIT INT TERM

i=0
while [ ! -S "$SOCK" ]; do
  i=$((i+1))
  [ "$i" -lt 50 ] || { echo "FAIL: test gateway socket not created" >&2; exit 1; }
  sleep 0.1
done

OUT=$("$TMP/pag" -config configs/pag.integration.json)
printf '%s\n' "$OUT"
printf '%s\n' "$OUT" | grep -q '1 gateway(s) healthy'
printf '%s\n' "$OUT" | grep -q 'gateway: pag-test'

kill "$PID"
wait "$PID" 2>/dev/null || true
PID=

if "$TMP/pag" -config configs/pag.integration.json >"$TMP/closed.out" 2>&1; then
  echo "FAIL: PAG started with required gateway absent" >&2
  exit 1
fi
grep -q 'fail closed' "$TMP/closed.out"

echo "PASS: external gateway discovered; gateway removal fails closed"
