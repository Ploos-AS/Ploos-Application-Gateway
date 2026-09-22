package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Version int               `json:"version"`
	Zones   map[string]Zone   `json:"zones"`
	Policy  Policy            `json:"policy"`
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
	if err != nil { return nil, err }
	var c Config
	if err := json.Unmarshal(b, &c); err != nil { return nil, err }
	if err := c.Validate(); err != nil { return nil, err }
	return &c, nil
}

func (c *Config) Validate() error {
	if c.Version != 1 { return fmt.Errorf("unsupported config version %d", c.Version) }
	if c.Policy.Default != "deny" { return fmt.Errorf("policy.default must be deny in M1") }
	for _, r := range c.Policy.Rules {
		if _, ok := c.Zones[r.From]; !ok { return fmt.Errorf("rule %q references unknown source zone %q", r.Name, r.From) }
		if _, ok := c.Zones[r.To]; !ok { return fmt.Errorf("rule %q references unknown destination zone %q", r.Name, r.To) }
		if r.Action == "proxy" {
			g, ok := c.Gateways[r.Gateway]
			if !ok { return fmt.Errorf("rule %q requires missing gateway %q", r.Name, r.Gateway) }
			if !g.Enabled { return fmt.Errorf("rule %q requires disabled gateway %q", r.Name, r.Gateway) }
		}
	}
	return nil
}
