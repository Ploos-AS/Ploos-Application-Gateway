#!/bin/sh
set -eu

need() { command -v "$1" >/dev/null 2>&1 || { echo "missing: $1" >&2; exit 77; }; }
need ip
need nft
need nc
need go

C=pag-dp-client
R=pag-dp-router
U=pag-dp-upstream
TMP=/tmp/pag-dataplane

cleanup() {
  ip netns del "$C" 2>/dev/null || true
  ip netns del "$R" 2>/dev/null || true
  ip netns del "$U" 2>/dev/null || true
  rm -rf "$TMP"
}
trap cleanup EXIT INT TERM
cleanup
mkdir -p "$TMP"

go build -o "$TMP/proxy" ./cmd/pag-test-proxy

ip netns add "$C"; ip netns add "$R"; ip netns add "$U"
ip link add dp-c type veth peer name dp-r0
ip link add dp-r1 type veth peer name dp-u
ip link set dp-c netns "$C"; ip link set dp-r0 netns "$R"
ip link set dp-r1 netns "$R"; ip link set dp-u netns "$U"
for ns in "$C" "$R" "$U"; do ip -n "$ns" link set lo up; done
ip -n "$C" link set dp-c up; ip -n "$R" link set dp-r0 up
ip -n "$R" link set dp-r1 up; ip -n "$U" link set dp-u up

ip -n "$C" addr add 192.0.2.2/24 dev dp-c
ip -n "$R" addr add 192.0.2.1/24 dev dp-r0
ip -n "$R" addr add 198.51.100.1/24 dev dp-r1
ip -n "$U" addr add 198.51.100.2/24 dev dp-u
ip -n "$C" route add 198.51.100.0/24 via 192.0.2.1
ip -n "$U" route add 192.0.2.0/24 via 198.51.100.1
ip netns exec "$R" sysctl -q -w net.ipv4.ip_forward=1

# Upstream echo service.
ip netns exec "$U" sh -c 'while true; do printf "UPSTREAM-OK\n" | nc -l -q 1 198.51.100.2 18080; done' &
UPID=$!

# Prove direct forwarding works before PAG policy is installed.
sleep 0.2
printf x | ip netns exec "$C" nc -w 2 198.51.100.2 18080 | grep -q UPSTREAM-OK

# PAG invariant: no forwarding through the router.
ip netns exec "$R" nft -f - <<'EOF'
table inet pag {
  chain forward {
    type filter hook forward priority 0; policy drop;
  }
}
EOF

if printf x | ip netns exec "$C" nc -w 1 198.51.100.2 18080 >/dev/null 2>&1; then
  echo "FAIL: direct client-to-upstream dataplane bypass succeeded" >&2
  exit 1
fi

# The gateway terminates the client session locally and creates a separate
# upstream session. It is intentionally a separate process in the router ns.
ip netns exec "$R" "$TMP/proxy" -listen 192.0.2.1:18081 -upstream 198.51.100.2:18080 &
PROXYPID=$!
sleep 0.2

OUT=$(printf hello | ip netns exec "$C" nc -w 2 192.0.2.1 18081)
printf '%s\n' "$OUT" | grep -q UPSTREAM-OK

kill "$PROXYPID" "$UPID" 2>/dev/null || true
echo "PASS: direct forwarding blocked; separately terminated proxy path succeeds"
