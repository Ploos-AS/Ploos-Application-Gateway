package dnswire

import (
	"encoding/binary"
	"fmt"
)

type Policy struct {
	AllowedTypes map[uint16]bool
	MaxUDPSize   uint16
}

func DefaultPolicy() Policy {
	return Policy{
		AllowedTypes: map[uint16]bool{
			1:true,   // A
			28:true,  // AAAA
			5:true,   // CNAME
			15:true,  // MX
			16:true,  // TXT
			2:true,   // NS
			6:true,   // SOA
			12:true,  // PTR
			33:true,  // SRV
			65:true,  // HTTPS
		},
		MaxUDPSize: 1232,
	}
}

func ValidateQueryPolicy(m []byte, p Policy) error {
	if err := ValidateQuery(m); err != nil { return err }
	q, err := ParseQuestion(m)
	if err != nil { return err }
	if !p.AllowedTypes[q.Type] { return fmt.Errorf("DNS type %d denied by policy", q.Type) }

	ar := binary.BigEndian.Uint16(m[10:12])
	if ar == 0 { return nil }
	if ar != 1 { return fmt.Errorf("only one additional EDNS record is allowed") }

	off := q.End
	if off >= len(m) || m[off] != 0 { return fmt.Errorf("additional record must be root-named EDNS OPT") }
	off++
	if off+10 > len(m) { return fmt.Errorf("truncated EDNS OPT record") }
	typ := binary.BigEndian.Uint16(m[off:off+2])
	if typ != 41 { return fmt.Errorf("additional record type %d is not EDNS OPT", typ) }
	udpSize := binary.BigEndian.Uint16(m[off+2:off+4])
	if udpSize < 512 { return fmt.Errorf("EDNS UDP size %d below DNS minimum", udpSize) }
	if p.MaxUDPSize != 0 && udpSize > p.MaxUDPSize { return fmt.Errorf("EDNS UDP size %d exceeds policy maximum %d", udpSize, p.MaxUDPSize) }
	rdlen := int(binary.BigEndian.Uint16(m[off+8:off+10]))
	if off+10+rdlen != len(m) { return fmt.Errorf("invalid EDNS OPT length") }
	return nil
}
