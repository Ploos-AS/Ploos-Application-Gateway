package dnswire

import (
	"encoding/binary"
	"fmt"
	"strings"
)

func ValidateResponse(query, response []byte) error {
	if err := ValidateQuery(query); err != nil { return fmt.Errorf("invalid original query: %w", err) }
	if len(response) < 12 { return fmt.Errorf("DNS response shorter than header") }

	qflags := binary.BigEndian.Uint16(query[2:4])
	rflags := binary.BigEndian.Uint16(response[2:4])
	if rflags&0x8000 == 0 { return fmt.Errorf("DNS response bit not set") }
	if (qflags>>11)&0x0f != (rflags>>11)&0x0f { return fmt.Errorf("DNS opcode mismatch") }
	if query[0] != response[0] || query[1] != response[1] { return fmt.Errorf("DNS transaction ID mismatch") }
	if binary.BigEndian.Uint16(response[4:6]) != 1 { return fmt.Errorf("DNS response must contain exactly one question") }

	qq, err := ParseQuestion(query)
	if err != nil { return err }
	rq, err := ParseQuestion(response)
	if err != nil { return fmt.Errorf("invalid response question: %w", err) }

	if !strings.EqualFold(qq.Name, rq.Name) { return fmt.Errorf("DNS question name mismatch") }
	if qq.Type != rq.Type { return fmt.Errorf("DNS question type mismatch") }
	if qq.Class != rq.Class { return fmt.Errorf("DNS question class mismatch") }
	return nil
}
