package dnswire

import "testing"

func TestIsTruncated(t *testing.T) {
	m:=make([]byte,12)
	m[2]=0x82
	if !IsTruncated(m) { t.Fatal("TC bit not detected") }
	m[2]=0x80
	if IsTruncated(m) { t.Fatal("false truncation") }
	if IsTruncated([]byte{1,2,3}) { t.Fatal("short message reported truncated") }
}
