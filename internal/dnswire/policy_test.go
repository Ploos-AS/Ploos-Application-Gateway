package dnswire

import "testing"

func TestDefaultPolicyAllowsA(t *testing.T) {
	q:=query([]byte{1,'a',0})
	if err:=ValidateQueryPolicy(q,DefaultPolicy()); err!=nil { t.Fatal(err) }
}

func TestDefaultPolicyDeniesAXFR(t *testing.T) {
	q:=query([]byte{1,'a',0})
	q[len(q)-4]=0; q[len(q)-3]=252
	if err:=ValidateQueryPolicy(q,DefaultPolicy()); err==nil { t.Fatal("AXFR must be denied") }
}

func TestEDNS1232Allowed(t *testing.T) {
	q:=query([]byte{1,'a',0})
	q[11]=1
	q=append(q,0,0,41,0x04,0xd0,0,0,0,0,0,0)
	if err:=ValidateQueryPolicy(q,DefaultPolicy()); err!=nil { t.Fatal(err) }
}

func TestOversizedEDNSRejected(t *testing.T) {
	q:=query([]byte{1,'a',0})
	q[11]=1
	q=append(q,0,0,41,0x10,0x00,0,0,0,0,0,0)
	if err:=ValidateQueryPolicy(q,DefaultPolicy()); err==nil { t.Fatal("oversized EDNS accepted") }
}

func TestNonOPTAdditionalRejected(t *testing.T) {
	q:=query([]byte{1,'a',0})
	q[11]=1
	q=append(q,0,0,1,0x02,0x00,0,0,0,0,0,0)
	if err:=ValidateQueryPolicy(q,DefaultPolicy()); err==nil { t.Fatal("non-OPT additional accepted") }
}
