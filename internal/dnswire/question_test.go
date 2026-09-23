package dnswire

import "testing"

func query(name []byte) []byte {
	m := make([]byte, 12)
	m[5] = 1
	m = append(m, name...)
	m = append(m, 0, 1, 0, 1)
	return m
}

func TestParseQuestion(t *testing.T) {
	q, err := ParseQuestion(query([]byte{3, 'w', 'w', 'w', 7, 'e', 'x', 'a', 'm', 'p', 'l', 'e', 3, 'c', 'o', 'm', 0}))
	if err != nil {
		t.Fatal(err)
	}
	if q.Name != "www.example.com" || q.Type != 1 || q.Class != 1 {
		t.Fatalf("unexpected question: %+v", q)
	}
}

func TestCompressionLoopRejected(t *testing.T) {
	m := make([]byte, 12)
	m[5] = 1
	m = append(m, 0xc0, 0x0c, 0, 1, 0, 1)
	if _, err := ParseQuestion(m); err == nil {
		t.Fatal("compression loop accepted")
	}
}

func TestPointerOutOfBoundsRejected(t *testing.T) {
	m := make([]byte, 12)
	m[5] = 1
	m = append(m, 0xc0, 0xff, 0, 1, 0, 1)
	if _, err := ParseQuestion(m); err == nil {
		t.Fatal("out-of-bounds pointer accepted")
	}
}

func TestTruncatedLabelRejected(t *testing.T) {
	m := make([]byte, 12)
	m[5] = 1
	m = append(m, 5, 'a', 'b')
	if _, err := ParseQuestion(m); err == nil {
		t.Fatal("truncated label accepted")
	}
}

func FuzzParseQuestion(f *testing.F) {
	f.Add(query([]byte{1, 'a', 0}))
	f.Fuzz(func(t *testing.T, m []byte) { _, _ = ParseQuestion(m) })
}
