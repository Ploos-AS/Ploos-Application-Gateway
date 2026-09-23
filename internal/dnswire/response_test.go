package dnswire

import "testing"

func responseFor(q []byte) []byte {
	r := append([]byte(nil), q...)
	r[2] |= 0x80
	return r
}

func TestValidateResponse(t *testing.T) {
	q := query([]byte{7, 'e', 'x', 'a', 'm', 'p', 'l', 'e', 3, 'c', 'o', 'm', 0})
	q[0] = 0x12
	q[1] = 0x34
	if err := ValidateResponse(q, responseFor(q)); err != nil {
		t.Fatal(err)
	}
}

func TestResponseTransactionMismatch(t *testing.T) {
	q := query([]byte{1, 'a', 0})
	r := responseFor(q)
	r[1] ^= 1
	if err := ValidateResponse(q, r); err == nil {
		t.Fatal("transaction mismatch accepted")
	}
}

func TestResponseQuestionMismatch(t *testing.T) {
	q := query([]byte{1, 'a', 0})
	r := responseFor(query([]byte{1, 'b', 0}))
	if err := ValidateResponse(q, r); err == nil {
		t.Fatal("question mismatch accepted")
	}
}

func TestResponseWithoutQRRejected(t *testing.T) {
	q := query([]byte{1, 'a', 0})
	r := append([]byte(nil), q...)
	if err := ValidateResponse(q, r); err == nil {
		t.Fatal("query accepted as response")
	}
}

func FuzzValidateResponse(f *testing.F) {
	q := query([]byte{1, 'a', 0})
	q[0] = 1
	f.Add(responseFor(q))
	f.Fuzz(func(t *testing.T, r []byte) { _ = ValidateResponse(q, r) })
}
