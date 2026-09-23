package dnswire

import (
	"encoding/binary"
	"fmt"
	"strings"
)

type Question struct {
	Name  string
	Type  uint16
	Class uint16
	End   int
}

func ParseQuestion(m []byte) (Question, error) {
	if len(m) < 12 {
		return Question{}, fmt.Errorf("DNS message shorter than header")
	}
	name, next, err := parseName(m, 12, map[int]bool{}, 0)
	if err != nil {
		return Question{}, err
	}
	if next+4 > len(m) {
		return Question{}, fmt.Errorf("truncated DNS question")
	}
	return Question{
		Name:  name,
		Type:  binary.BigEndian.Uint16(m[next : next+2]),
		Class: binary.BigEndian.Uint16(m[next+2 : next+4]),
		End:   next + 4,
	}, nil
}

func parseName(m []byte, off int, seen map[int]bool, depth int) (string, int, error) {
	if depth > 16 {
		return "", 0, fmt.Errorf("DNS compression depth exceeded")
	}
	if off < 0 || off >= len(m) {
		return "", 0, fmt.Errorf("DNS name offset out of bounds")
	}

	labels := make([]string, 0, 8)
	next := -1
	total := 1
	for {
		if off >= len(m) {
			return "", 0, fmt.Errorf("truncated DNS name")
		}
		n := int(m[off])
		if n == 0 {
			if next < 0 {
				next = off + 1
			}
			break
		}
		if n&0xc0 == 0xc0 {
			if off+1 >= len(m) {
				return "", 0, fmt.Errorf("truncated DNS compression pointer")
			}
			ptr := int(m[off]&0x3f)<<8 | int(m[off+1])
			if ptr >= len(m) {
				return "", 0, fmt.Errorf("DNS compression pointer out of bounds")
			}
			if seen[ptr] {
				return "", 0, fmt.Errorf("DNS compression loop")
			}
			seen[ptr] = true
			suffix, _, err := parseName(m, ptr, seen, depth+1)
			if err != nil {
				return "", 0, err
			}
			if suffix != "" {
				labels = append(labels, strings.Split(suffix, ".")...)
			}
			if next < 0 {
				next = off + 2
			}
			break
		}
		if n&0xc0 != 0 {
			return "", 0, fmt.Errorf("reserved DNS label encoding")
		}
		if n > 63 {
			return "", 0, fmt.Errorf("DNS label too long")
		}
		off++
		if off+n > len(m) {
			return "", 0, fmt.Errorf("truncated DNS label")
		}
		total += n + 1
		if total > 255 {
			return "", 0, fmt.Errorf("DNS name too long")
		}
		labels = append(labels, string(m[off:off+n]))
		off += n
	}
	return strings.Join(labels, "."), next, nil
}
