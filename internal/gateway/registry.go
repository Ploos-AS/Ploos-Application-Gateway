package gateway

import (
	"fmt"
	"sort"
)

type State struct {
	ID      string
	Version string
	Socket  string
	Healthy bool
}

type Registry struct{ gateways map[string]State }

func NewRegistry() *Registry { return &Registry{gateways: make(map[string]State)} }

func (r *Registry) Register(g State) error {
	if g.ID == "" {
		return fmt.Errorf("gateway id is required")
	}
	if g.Socket == "" {
		return fmt.Errorf("gateway %q socket is required", g.ID)
	}
	if _, exists := r.gateways[g.ID]; exists {
		return fmt.Errorf("gateway %q already registered", g.ID)
	}
	r.gateways[g.ID] = g
	return nil
}

func (r *Registry) Get(id string) (State, bool) { g, ok := r.gateways[id]; return g, ok }

func (r *Registry) IDs() []string {
	ids := make([]string, 0, len(r.gateways))
	for id := range r.gateways {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}
