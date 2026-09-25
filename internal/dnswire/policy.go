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
			1:  true, // A
			28: true, // AAAA
			5:  true, // CNAME
			15: true, // MX
			16: true, // TXT
			2:  true, // NS
			6:  true, // SOA
			12: true, // PTR
			33:  true, // SRV
			35:  true, // NAPTR
			43:  true, // DS
			44:  true, // SSHFP
			48:  true, // DNSKEY
			52:  true, // TLSA
			64:  true, // SVCB
			65:  true, // HTTPS
			257: true, // CAA
		},
		MaxUDPSize: 1232,
	}
}

func ValidateQueryPolicy(m []byte, p Policy) error {
	if err := ValidateQuery(m); err != nil {
		return err
	}
	q, err := ParseQuestion(m)
	if err != nil {
		return err
	}
	if !p.AllowedTypes[q.Type] {
		return fmt.Errorf("DNS type %d denied by policy", q.Type)
	}

	if binary.BigEndian.Uint16(m[6:8]) != 0 {
		return fmt.Errorf("answer records are not allowed in queries")
	}
	if binary.BigEndian.Uint16(m[8:10]) != 0 {
		return fmt.Errorf("authority records are not allowed in queries")
	}
	ar := binary.BigEndian.Uint16(m[10:12])
	if ar == 0 {
		if q.End != len(m) {
			return fmt.Errorf("trailing data after DNS question")
		}
		return nil
	}
	if ar != 1 {
		return fmt.Errorf("only one additional EDNS record is allowed")
	}

	off := q.End
	if off >= len(m) || m[off] != 0 {
		return fmt.Errorf("additional record must be root-named EDNS OPT")
	}
	off++
	if off+10 > len(m) {
		return fmt.Errorf("truncated EDNS OPT record")
	}
	typ := binary.BigEndian.Uint16(m[off : off+2])
	if typ != 41 {
		return fmt.Errorf("additional record type %d is not EDNS OPT", typ)
	}
	udpSize := binary.BigEndian.Uint16(m[off+2 : off+4])
	ttl := binary.BigEndian.Uint32(m[off+4 : off+8])
	version := uint8((ttl >> 16) & 0xff)
	if version != 0 {
		return fmt.Errorf("unsupported EDNS version %d", version)
	}
	flags := uint16(ttl & 0xffff)
	if flags&0x7fff != 0 {
		return fmt.Errorf("reserved EDNS flags are set")
	}
	if udpSize < 512 {
		return fmt.Errorf("EDNS UDP size %d below DNS minimum", udpSize)
	}
	if p.MaxUDPSize != 0 && udpSize > p.MaxUDPSize {
		return fmt.Errorf("EDNS UDP size %d exceeds policy maximum %d", udpSize, p.MaxUDPSize)
	}
	rdlen := int(binary.BigEndian.Uint16(m[off+8 : off+10]))
	if off+10+rdlen != len(m) {
		return fmt.Errorf("invalid EDNS OPT length")
	}
	return nil
}
