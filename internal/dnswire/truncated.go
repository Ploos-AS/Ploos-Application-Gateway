package dnswire

// IsTruncated reports whether a DNS response has the TC bit set.
func IsTruncated(m []byte) bool {
	return len(m) >= 4 && m[2]&0x02 != 0
}
