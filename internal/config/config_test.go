package config

import "testing"

func valid() *Config {
	return &Config{
		Version:1,
		Zones:map[string]Zone{"lan":{Interfaces:[]string{"lan0"}},"wan":{Interfaces:[]string{"wan0"}}},
		Policy:Policy{Default:"deny",Rules:[]Rule{{Name:"dns",From:"lan",To:"wan",Protocol:"dns",Action:"proxy",Gateway:"pag-dns"}}},
		Gateways:map[string]Gateway{"pag-dns":{Enabled:true,Socket:"/run/pag/pag-dns.sock"}},
	}
}

func TestValidConfig(t *testing.T) { if err:=valid().Validate(); err!=nil { t.Fatal(err) } }

func TestInterfaceCannotBelongToTwoZones(t *testing.T) {
	c:=valid(); c.Zones["wan"]=Zone{Interfaces:[]string{"lan0"}}
	if err:=c.Validate(); err==nil { t.Fatal("duplicate interface assignment must fail") }
}

func TestDuplicateRuleNameRejected(t *testing.T) {
	c:=valid(); c.Policy.Rules=append(c.Policy.Rules,c.Policy.Rules[0])
	if err:=c.Validate(); err==nil { t.Fatal("duplicate rule names must fail") }
}

func TestSameZoneRuleRejected(t *testing.T) {
	c:=valid(); c.Policy.Rules[0].To="lan"
	if err:=c.Validate(); err==nil { t.Fatal("same-zone boundary rule must fail") }
}

func TestUnknownActionRejected(t *testing.T) {
	c:=valid(); c.Policy.Rules[0].Action="accept"
	if err:=c.Validate(); err==nil { t.Fatal("plain accept must not be supported") }
}

func TestDenyRuleNeedsNoGateway(t *testing.T) {
	c:=valid(); c.Policy.Rules[0].Action="deny"; c.Policy.Rules[0].Gateway=""
	if err:=c.Validate(); err!=nil { t.Fatal(err) }
}
