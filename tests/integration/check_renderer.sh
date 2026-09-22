#!/bin/sh
set -eu

go build -o /tmp/pag ./cmd/pag
/tmp/pag -config configs/pag.example.json -nft-dry-run > /tmp/pag.nft

grep -q 'table inet pag' /tmp/pag.nft
grep -q 'policy drop;' /tmp/pag.nft

if grep -Eq '(^|[[:space:]])accept([[:space:];]|$)' /tmp/pag.nft; then
  echo "FAIL: renderer emitted an accept rule" >&2
  exit 1
fi

nft -c -f /tmp/pag.nft

echo "PASS: generated nftables ruleset is syntactically valid and default-deny"
