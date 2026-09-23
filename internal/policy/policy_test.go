package policy

import (
	"github.com/Ploos-AS/Ploos-Application-Gateway/internal/config"
	"testing"
)

func TestDefaultDeny(t *testing.T) {
	c := &config.Config{Policy: config.Policy{Default: "deny"}}
	if d := Evaluate(c, "lan", "wan", "smtp"); d.Allowed {
		t.Fatal("unmatched traffic must be denied")
	}
}

func TestProxyRule(t *testing.T) {
	c := &config.Config{
		Gateways: map[string]config.Gateway{"pag-dns": {Enabled: true}},
		Policy:   config.Policy{Default: "deny", Rules: []config.Rule{{Name: "dns", From: "lan", To: "wan", Protocol: "dns", Action: "proxy", Gateway: "pag-dns"}}},
	}
	d := Evaluate(c, "lan", "wan", "dns")
	if !d.Allowed || d.Gateway != "pag-dns" {
		t.Fatalf("unexpected decision: %+v", d)
	}
}

func TestMissingGatewayFailsClosed(t *testing.T) {
	c := &config.Config{Policy: config.Policy{Default: "deny", Rules: []config.Rule{{Name: "smtp", From: "lan", To: "wan", Protocol: "smtp", Action: "proxy", Gateway: "pag-smtp"}}}}
	if d := Evaluate(c, "lan", "wan", "smtp"); d.Allowed {
		t.Fatal("missing gateway must fail closed")
	}
}
