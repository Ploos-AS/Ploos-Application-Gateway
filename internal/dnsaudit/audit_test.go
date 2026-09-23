package dnsaudit

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestAuditIsPrivacyMinimal(t *testing.T) {
	var b bytes.Buffer
	l := New(&b)
	l.now = func() time.Time { return time.Unix(0, 0) }
	l.Log("deny", "udp", 252, "policy")
	s := b.String()
	if !strings.Contains(s, "deny") || !strings.Contains(s, "252") { t.Fatal(s) }
	if strings.Contains(s, "client") || strings.Contains(s, "qname") { t.Fatal("privacy-sensitive field present") }
}
