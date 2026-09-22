package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/Ploos-AS/Ploos-Application-Gateway/internal/config"
)

type Hello struct {
	ID      string `json:"id"`
	Version string `json:"version"`
	Healthy bool   `json:"healthy"`
}

func Discover(ctx context.Context, id string, cfg config.Gateway) (State, error) {
	if !cfg.Enabled { return State{}, fmt.Errorf("gateway %q is disabled", id) }
	d := net.Dialer{Timeout: 2 * time.Second}
	conn, err := d.DialContext(ctx, "unix", cfg.Socket)
	if err != nil { return State{}, fmt.Errorf("gateway %q unavailable: %w", id, err) }
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(2 * time.Second))
	if _, err := conn.Write([]byte("{\"op\":\"hello\"}\n")); err != nil { return State{}, err }
	var h Hello
	if err := json.NewDecoder(conn).Decode(&h); err != nil { return State{}, fmt.Errorf("gateway %q invalid hello: %w", id, err) }
	if h.ID != id { return State{}, fmt.Errorf("gateway identity mismatch: configured %q, reported %q", id, h.ID) }
	if !h.Healthy { return State{}, fmt.Errorf("gateway %q is unhealthy", id) }
	return State{ID:h.ID, Version:h.Version, Socket:cfg.Socket, Healthy:true}, nil
}

func DiscoverAll(ctx context.Context, gateways map[string]config.Gateway) (*Registry, error) {
	r := NewRegistry()
	for id, cfg := range gateways {
		if !cfg.Enabled { continue }
		g, err := Discover(ctx, id, cfg)
		if err != nil { return nil, err }
		if err := r.Register(g); err != nil { return nil, err }
	}
	return r, nil
}
