package dnswire

import "testing"

func TestValidateQuery(t *testing.T) {
	q := query([]byte{1, 'a', 0})
	if err := ValidateQuery(q); err != nil {
		t.Fatal(err)
	}
}
func TestRejectShort(t *testing.T) {
	if err := ValidateQuery([]byte{1, 2}); err == nil {
		t.Fatal("short DNS message accepted")
	}
}
func TestRejectResponse(t *testing.T) {
	q := query([]byte{1, 'a', 0})
	q[2] = 0x80
	if err := ValidateQuery(q); err == nil {
		t.Fatal("response accepted as query")
	}
}
func TestRejectMultipleQuestions(t *testing.T) {
	q := query([]byte{1, 'a', 0})
	q[5] = 2
	if err := ValidateQuery(q); err == nil {
		t.Fatal("multiple questions accepted")
	}
}
