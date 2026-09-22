package policy

import "github.com/Ploos-AS/Ploos-Application-Gateway/internal/config"

type Decision struct {
	Allowed bool
	Gateway string
	Rule    string
}

func Evaluate(c *config.Config, from, to, protocol string) Decision {
	for _, r := range c.Policy.Rules {
		if r.From == from && r.To == to && r.Protocol == protocol {
			if r.Action != "proxy" { return Decision{Rule: r.Name} }
			g, ok := c.Gateways[r.Gateway]
			if !ok || !g.Enabled { return Decision{Rule: r.Name} }
			return Decision{Allowed: true, Gateway: r.Gateway, Rule: r.Name}
		}
	}
	return Decision{}
}
