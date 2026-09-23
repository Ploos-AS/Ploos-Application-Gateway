package nft

import (
	"strings"
	"testing"

	"github.com/Ploos-AS/Ploos-Application-Gateway/internal/config"
)

func testConfig() *config.Config {
	return &config.Config{
		Version: 1,
		Zones: map[string]config.Zone{
			"lan": {Interfaces: []string{"lan0"}},
			"wan": {Interfaces: []string{"wan0"}},
		},
		Policy: config.Policy{
			Default: "deny",
			Rules: []config.Rule{{
				Name: "lan-dns", From: "lan", To: "wan", Protocol: "dns",
				Action: "proxy", Gateway: "pag-dns",
			}},
		},
		Gateways: map[string]config.Gateway{
			"pag-dns": {Enabled: true, Socket: "/run/pag/pag-dns.sock"},
		},
	}
}

func TestRenderIsDefaultDeny(t *testing.T) {
	out, err := Render(testConfig())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "policy drop;") {
		t.Fatal("forward chain must default to drop")
	}
	if strings.Contains(out, " accept") {
		t.Fatal("M1 renderer must not create forwarding accept rules")
	}
}

func TestRenderDocumentsAntiBypass(t *testing.T) {
	out, err := Render(testConfig())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "direct forwarding stays blocked") {
		t.Fatal("proxy policy must preserve anti-bypass behavior")
	}
}

func TestUnsafeInterfaceRejected(t *testing.T) {
	c := testConfig()
	c.Zones["lan"] = config.Zone{Interfaces: []string{"lan0; flush ruleset"}}
	if _, err := Render(c); err == nil {
		t.Fatal("unsafe interface name must be rejected")
	}
}
