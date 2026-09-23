package gateway

import (
	"context"
	"encoding/json"
	"net"
	"path/filepath"
	"testing"

	"github.com/Ploos-AS/Ploos-Application-Gateway/internal/config"
)

func serveHello(t *testing.T, id string, healthy bool) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "gateway.sock")
	ln, err := net.Listen("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		defer ln.Close()
		c, err := ln.Accept()
		if err != nil {
			return
		}
		defer c.Close()
		var req map[string]string
		_ = json.NewDecoder(c).Decode(&req)
		_ = json.NewEncoder(c).Encode(Hello{ID: id, Version: "0.1.0", Healthy: healthy})
	}()
	return path
}

func TestDiscover(t *testing.T) {
	path := serveHello(t, "pag-dns", true)
	g, err := Discover(context.Background(), "pag-dns", config.Gateway{Enabled: true, Socket: path})
	if err != nil {
		t.Fatal(err)
	}
	if !g.Healthy || g.ID != "pag-dns" {
		t.Fatalf("unexpected gateway: %+v", g)
	}
}

func TestIdentityMismatchFailsClosed(t *testing.T) {
	path := serveHello(t, "pag-http", true)
	if _, err := Discover(context.Background(), "pag-dns", config.Gateway{Enabled: true, Socket: path}); err == nil {
		t.Fatal("identity mismatch must fail")
	}
}

func TestUnhealthyFailsClosed(t *testing.T) {
	path := serveHello(t, "pag-dns", false)
	if _, err := Discover(context.Background(), "pag-dns", config.Gateway{Enabled: true, Socket: path}); err == nil {
		t.Fatal("unhealthy gateway must fail")
	}
}
