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
	return nil
}
