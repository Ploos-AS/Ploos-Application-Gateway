#!/bin/sh
set -eu

C=pag-dns-client
R=pag-dns-router
U=pag-dns-upstream
TMP=/tmp/pag-dns-integration
SOCK=/tmp/pag-dns-control.sock

cleanup() {
  ip netns del "$C" 2>/dev/null || true
  ip netns del "$R" 2>/dev/null || true
  ip netns del "$U" 2>/dev/null || true
  rm -rf "$TMP" "$SOCK"
}
trap cleanup EXIT INT TERM
cleanup
mkdir -p "$TMP"

go build -o "$TMP/pag-dns" ./cmd/pag-dns
go build -o "$TMP/test-dns" ./cmd/pag-test-dns

ip netns add "$C"; ip netns add "$R"; ip netns add "$U"
ip link add dns-c type veth peer name dns-r0
ip link add dns-r1 type veth peer name dns-u
ip link set dns-c netns "$C"; ip link set dns-r0 netns "$R"
ip link set dns-r1 netns "$R"; ip link set dns-u netns "$U"
for ns in "$C" "$R" "$U"; do ip -n "$ns" link set lo up; done
ip -n "$C" link set dns-c up; ip -n "$R" link set dns-r0 up
ip -n "$R" link set dns-r1 up; ip -n "$U" link set dns-u up

ip -n "$C" addr add 192.0.2.2/24 dev dns-c
ip -n "$R" addr add 192.0.2.1/24 dev dns-r0
ip -n "$R" addr add 198.51.100.1/24 dev dns-r1
ip -n "$U" addr add 198.51.100.2/24 dev dns-u
ip -n "$C" route add 198.51.100.0/24 via 192.0.2.1
ip -n "$U" route add 192.0.2.0/24 via 198.51.100.1
ip netns exec "$R" sysctl -q -w net.ipv4.ip_forward=1

ip netns exec "$U" "$TMP/test-dns" -listen 198.51.100.2:5353 -truncate-udp &
UPID=$!
sleep .2

# Direct routing is blocked by PAG.
ip netns exec "$R" nft -f - <<'EOF'
table inet pag {
 chain forward {
  type filter hook forward priority 0; policy drop;
 }
}
EOF

# Start the real protocol gateway. It terminates on the LAN-side router address
# and creates its own upstream DNS session on the WAN side.
ip netns exec "$R" "$TMP/pag-dns"   -listen 192.0.2.1:5353   -upstream 198.51.100.2:5353   -control "$SOCK" &
DNSPID=$!
sleep .3

# Build one deterministic A/IN query for example.com and send it via UDP.
python3 - <<'PY' >"$TMP/query.bin"
import sys
q=bytearray(b'\x12\x34\x01\x00\x00\x01\x00\x00\x00\x00\x00\x00')
for label in b'example.com'.split(b'.'):
    q.append(len(label)); q.extend(label)
q.extend(b'\x00\x00\x01\x00\x01')
sys.stdout.buffer.write(q)
PY

cat "$TMP/query.bin" | ip netns exec "$C" nc -u -w 2 192.0.2.1 5353 >"$TMP/response.bin"
python3 - "$TMP/response.bin" <<'PY'
import sys
r=open(sys.argv[1],'rb').read()
assert len(r)>=12, "short DNS response"
assert r[:2]==b'\x12\x34', "transaction mismatch"
assert r[2]&0x80, "response bit missing"
print("PASS: pag-dns UDP query succeeded through TC-to-TCP fallback")
PY

# Native DNS-over-TCP path: send the same query with RFC 1035 length framing.
python3 - "$TMP/query.bin" <<'PY' >"$TMP/tcp-query.bin"
import struct,sys
q=open(sys.argv[1],'rb').read()
sys.stdout.buffer.write(struct.pack("!H",len(q))+q)
PY
cat "$TMP/tcp-query.bin" | ip netns exec "$C" nc -w 2 192.0.2.1 5353 >"$TMP/tcp-response.bin"
python3 - "$TMP/tcp-response.bin" <<'PY'
import struct,sys
r=open(sys.argv[1],'rb').read()
assert len(r)>=14, "short framed TCP DNS response"
n=struct.unpack("!H",r[:2])[0]
assert n==len(r)-2, "TCP DNS length mismatch"
dns=r[2:]
assert dns[:2]==b'\x12\x34', "TCP transaction mismatch"
assert dns[2]&0x80, "TCP response bit missing"
print("PASS: pag-dns native TCP path")
PY

# Direct client access to the upstream must still fail.
if cat "$TMP/query.bin" | ip netns exec "$C" nc -u -w 1 198.51.100.2 5353 >"$TMP/direct.bin" 2>/dev/null && [ -s "$TMP/direct.bin" ]; then
 echo "FAIL: client bypassed pag-dns" >&2
 exit 1
fi

kill "$DNSPID" "$UPID" 2>/dev/null || true
echo "PASS: direct DNS blocked; pag-dns path succeeds"
