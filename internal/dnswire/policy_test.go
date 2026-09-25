package dnswire

import "testing"

func TestDefaultPolicyAllowsA(t *testing.T) {
	q := query([]byte{1, 'a', 0})
	if err := ValidateQueryPolicy(q, DefaultPolicy()); err != nil {
		t.Fatal(err)
	}
}

func TestDefaultPolicyDeniesAXFR(t *testing.T) {
	q := query([]byte{1, 'a', 0})
	q[len(q)-4] = 0
	q[len(q)-3] = 252
	if err := ValidateQueryPolicy(q, DefaultPolicy()); err == nil {
		t.Fatal("AXFR must be denied")
	}
}

func TestEDNS1232Allowed(t *testing.T) {
	q := query([]byte{1, 'a', 0})
	q[11] = 1
	q = append(q, 0, 0, 41, 0x04, 0xd0, 0, 0, 0, 0, 0, 0)
	if err := ValidateQueryPolicy(q, DefaultPolicy()); err != nil {
		t.Fatal(err)
	}
}

func TestOversizedEDNSRejected(t *testing.T) {
	q := query([]byte{1, 'a', 0})
	q[11] = 1
	q = append(q, 0, 0, 41, 0x10, 0x00, 0, 0, 0, 0, 0, 0)
	if err := ValidateQueryPolicy(q, DefaultPolicy()); err == nil {
		t.Fatal("oversized EDNS accepted")
	}
}

func TestNonOPTAdditionalRejected(t *testing.T) {
	q := query([]byte{1, 'a', 0})
	q[11] = 1
	q = append(q, 0, 0, 1, 0x02, 0x00, 0, 0, 0, 0, 0, 0)
	if err := ValidateQueryPolicy(q, DefaultPolicy()); err == nil {
		t.Fatal("non-OPT additional accepted")
	}
}

func TestAnswerSectionRejectedInQuery(t *testing.T) {
	q := query([]byte{1, 'a', 0})
	q[7] = 1
	if err := ValidateQueryPolicy(q, DefaultPolicy()); err == nil {
		t.Fatal("answer section accepted in query")
	}
}

func TestAuthoritySectionRejectedInQuery(t *testing.T) {
	q := query([]byte{1, 'a', 0})
	q[9] = 1
	if err := ValidateQueryPolicy(q, DefaultPolicy()); err == nil {
		t.Fatal("authority section accepted in query")
	}
}

func TestTrailingDataRejectedWithoutAdditional(t *testing.T) {
	q := append(query([]byte{1, 'a', 0}), 0)
	if err := ValidateQueryPolicy(q, DefaultPolicy()); err == nil {
		t.Fatal("trailing data accepted")
	}
}

func TestEDNSVersionOneRejected(t *testing.T) {
	q := query([]byte{1, 'a', 0})
	q[11] = 1
	q = append(q, 0, 0, 41, 0x04, 0xd0, 0, 1, 0, 0, 0, 0)
	if err := ValidateQueryPolicy(q, DefaultPolicy()); err == nil {
		t.Fatal("EDNS version 1 accepted")
	}
}

func TestEDNSReservedFlagsRejected(t *testing.T) {
	q := query([]byte{1, 'a', 0})
	q[11] = 1
	q = append(q, 0, 0, 41, 0x04, 0xd0, 0, 0, 0, 1, 0, 0)
	if err := ValidateQueryPolicy(q, DefaultPolicy()); err == nil {
		t.Fatal("reserved EDNS flags accepted")
	}
}

func TestEDNSDOFlagAllowed(t *testing.T) {
	q := query([]byte{1, 'a', 0})
	q[11] = 1
	q = append(q, 0, 0, 41, 0x04, 0xd0, 0, 0, 0x80, 0, 0, 0)
	if err := ValidateQueryPolicy(q, DefaultPolicy()); err != nil {
		t.Fatal(err)
	}
}


func TestDefaultPolicyAllowsModernSafeTypes(t *testing.T) {
	for _, typ := range []uint16{35, 43, 44, 48, 52, 64, 65, 257} {
		q := query([]byte{1, 'a', 0})
		q[len(q)-4] = byte(typ >> 8)
		q[len(q)-3] = byte(typ)
		if err := ValidateQueryPolicy(q, DefaultPolicy()); err != nil {
			t.Fatalf("DNS type %d rejected: %v", typ, err)
		}
	}
}

func TestDefaultPolicyStillDeniesTransferAndAnyTypes(t *testing.T) {
	for _, typ := range []uint16{251, 252, 255} {
		q := query([]byte{1, 'a', 0})
		q[len(q)-4] = byte(typ >> 8)
		q[len(q)-3] = byte(typ)
		if err := ValidateQueryPolicy(q, DefaultPolicy()); err == nil {
			t.Fatalf("DNS type %d must be denied", typ)
		}
	}
}
