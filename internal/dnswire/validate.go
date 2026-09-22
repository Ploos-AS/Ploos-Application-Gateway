package dnswire

import (
	"encoding/binary"
	"fmt"
)

func ValidateQuery(m []byte) error {
	if len(m) < 12 { return fmt.Errorf("DNS message shorter than header") }
	if binary.BigEndian.Uint16(m[4:6]) != 1 { return fmt.Errorf("DNS query must contain exactly one question") }
	flags := binary.BigEndian.Uint16(m[2:4])
	if flags&0x8000 != 0 { return fmt.Errorf("DNS query has response bit set") }
	opcode := (flags >> 11) & 0x0f
	if opcode != 0 { return fmt.Errorf("unsupported DNS opcode %d", opcode) }
	q, err := ParseQuestion(m)
	if err != nil { return err }
	if q.Name == "" { return fmt.Errorf("empty DNS question name") }
	if q.Class != 1 { return fmt.Errorf("unsupported DNS class %d", q.Class) }
	return nil
}
