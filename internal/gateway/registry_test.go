package gateway

import "testing"

func TestRegistry(t *testing.T) {
	r := NewRegistry()
	if err := r.Register(State{ID: "pag-dns", Version: "0.1", Socket: "/run/pag/pag-dns.sock", Healthy: true}); err != nil {
		t.Fatal(err)
	}
	g, ok := r.Get("pag-dns")
	if !ok || g.ID != "pag-dns" {
		t.Fatal("registered gateway not found")
	}
}

func TestDuplicateRegistrationRejected(t *testing.T) {
	r := NewRegistry()
	g := State{ID: "pag-dns", Socket: "/run/pag/pag-dns.sock"}
	if err := r.Register(g); err != nil {
		t.Fatal(err)
	}
	if err := r.Register(g); err == nil {
		t.Fatal("duplicate gateway must be rejected")
	}
}
