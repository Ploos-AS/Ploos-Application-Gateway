package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Version  int                `json:"version"`
	Zones    map[string]Zone    `json:"zones"`
	Policy   Policy             `json:"policy"`
	Gateways map[string]Gateway `json:"gateways"`
}

type Zone struct {
	Interfaces []string `json:"interfaces"`
}
type Gateway struct {
	Enabled bool   `json:"enabled"`
	Socket  string `json:"socket"`
}
type Policy struct {
	Default string `json:"default"`
	Rules   []Rule `json:"rules"`
}

type Rule struct {
	Name     string `json:"name"`
	From     string `json:"from"`
	To       string `json:"to"`
	Protocol string `json:"protocol"`
	Action   string `json:"action"`
	Gateway  string `json:"gateway"`
}

func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c Config
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &c, nil
}

func (c *Config) Validate() error {
	if c.Version != 1 {
		return fmt.Errorf("unsupported config version %d", c.Version)
	}
	if c.Policy.Default != "deny" {
		return fmt.Errorf("policy.default must be deny in M1")
	}
	if len(c.Zones) == 0 {
		return fmt.Errorf("at least one zone is required")
	}

	seenInterfaces := map[string]string{}
	for name, z := range c.Zones {
		if name == "" {
			return fmt.Errorf("zone name must not be empty")
		}
		if len(z.Interfaces) == 0 {
			return fmt.Errorf("zone %q has no interfaces", name)
		}
		for _, iface := range z.Interfaces {
			if iface == "" {
				return fmt.Errorf("zone %q contains an empty interface", name)
			}
			if previous, exists := seenInterfaces[iface]; exists {
				return fmt.Errorf("interface %q belongs to both zone %q and %q", iface, previous, name)
			}
			seenInterfaces[iface] = name
		}
	}

	seenRules := map[string]bool{}
	for i, r := range c.Policy.Rules {
		if r.Name == "" {
			return fmt.Errorf("rule %d has no name", i)
		}
		if seenRules[r.Name] {
			return fmt.Errorf("duplicate rule name %q", r.Name)
		}
		seenRules[r.Name] = true
		if r.From == r.To {
			return fmt.Errorf("rule %q crosses no zone boundary", r.Name)
		}
		if _, ok := c.Zones[r.From]; !ok {
			return fmt.Errorf("rule %q references unknown source zone %q", r.Name, r.From)
		}
		if _, ok := c.Zones[r.To]; !ok {
			return fmt.Errorf("rule %q references unknown destination zone %q", r.Name, r.To)
		}
		if r.Protocol == "" {
			return fmt.Errorf("rule %q has no protocol", r.Name)
		}
		if r.Action != "proxy" && r.Action != "deny" {
			return fmt.Errorf("rule %q has unsupported action %q", r.Name, r.Action)
		}
		if r.Action == "deny" {
			if r.Gateway != "" {
				return fmt.Errorf("deny rule %q must not specify a gateway", r.Name)
			}
			continue
		}
		if r.Gateway == "" {
			return fmt.Errorf("proxy rule %q has no gateway", r.Name)
		}
		g, ok := c.Gateways[r.Gateway]
		if !ok {
			return fmt.Errorf("rule %q requires missing gateway %q", r.Name, r.Gateway)
		}
		if !g.Enabled {
			return fmt.Errorf("rule %q requires disabled gateway %q", r.Name, r.Gateway)
		}
		if g.Socket == "" {
			return fmt.Errorf("gateway %q has no socket", r.Gateway)
		}
	}
	return nil
}
