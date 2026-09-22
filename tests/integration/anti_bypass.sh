#!/bin/sh
set -eu

need() { command -v "$1" >/dev/null 2>&1 || { echo "missing: $1" >&2; exit 77; }; }
need ip
need nft
need ping

CLIENT=pag-ci-client
ROUTER=pag-ci-router
UPSTREAM=pag-ci-upstream

cleanup() {
  ip netns del "$CLIENT" 2>/dev/null || true
  ip netns del "$ROUTER" 2>/dev/null || true
  ip netns del "$UPSTREAM" 2>/dev/null || true
}
trap cleanup EXIT INT TERM
cleanup

ip netns add "$CLIENT"
ip netns add "$ROUTER"
ip netns add "$UPSTREAM"

ip link add pag-c type veth peer name pag-r0
ip link add pag-r1 type veth peer name pag-u

ip link set pag-c netns "$CLIENT"
ip link set pag-r0 netns "$ROUTER"
ip link set pag-r1 netns "$ROUTER"
ip link set pag-u netns "$UPSTREAM"

for ns in "$CLIENT" "$ROUTER" "$UPSTREAM"; do ip -n "$ns" link set lo up; done
for pair in "$CLIENT pag-c" "$ROUTER pag-r0" "$ROUTER pag-r1" "$UPSTREAM pag-u"; do
  set -- $pair; ip -n "$1" link set "$2" up
done

ip -n "$CLIENT" addr add 192.0.2.2/24 dev pag-c
ip -n "$ROUTER" addr add 192.0.2.1/24 dev pag-r0
ip -n "$ROUTER" addr add 198.51.100.1/24 dev pag-r1
ip -n "$UPSTREAM" addr add 198.51.100.2/24 dev pag-u
ip -n "$CLIENT" route add 198.51.100.0/24 via 192.0.2.1
ip -n "$UPSTREAM" route add 192.0.2.0/24 via 198.51.100.1

ip -n "$CLIENT" -6 addr add 2001:db8:1::2/64 dev pag-c
ip -n "$ROUTER" -6 addr add 2001:db8:1::1/64 dev pag-r0
ip -n "$ROUTER" -6 addr add 2001:db8:2::1/64 dev pag-r1
ip -n "$UPSTREAM" -6 addr add 2001:db8:2::2/64 dev pag-u
ip -n "$CLIENT" -6 route add 2001:db8:2::/64 via 2001:db8:1::1
ip -n "$UPSTREAM" -6 route add 2001:db8:1::/64 via 2001:db8:2::1

ip netns exec "$ROUTER" sysctl -q -w net.ipv4.ip_forward=1
ip netns exec "$ROUTER" sysctl -q -w net.ipv6.conf.all.forwarding=1

# Baseline proves the namespace topology itself can route.
ip netns exec "$CLIENT" ping -c 1 -W 1 198.51.100.2 >/dev/null
ip netns exec "$CLIENT" ping -6 -c 1 -W 1 2001:db8:2::2 >/dev/null

ip netns exec "$ROUTER" nft -f - <<'EOF'
table inet pag {
  chain forward {
    type filter hook forward priority 0; policy drop;
  }
}
EOF

if ip netns exec "$CLIENT" ping -c 1 -W 1 198.51.100.2 >/dev/null 2>&1; then
  echo "FAIL: IPv4 bypassed PAG default deny" >&2
  exit 1
fi
if ip netns exec "$CLIENT" ping -6 -c 1 -W 1 2001:db8:2::2 >/dev/null 2>&1; then
  echo "FAIL: IPv6 bypassed PAG default deny" >&2
  exit 1
fi

echo "PASS: PAG anti-bypass blocks direct IPv4 and IPv6 forwarding"
