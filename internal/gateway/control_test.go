package gateway

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/Ploos-AS/Ploos-Application-Gateway/internal/config"
)

func TestControlServerDiscovery(t *testing.T) {
	socket:=filepath.Join(t.TempDir(),"control.sock")
	ln,err:=(ControlServer{ID:"pag-dns",Version:"0.1.0",Socket:socket}).Serve()
	if err!=nil { t.Fatal(err) }
	defer ln.Close()
	g,err:=Discover(context.Background(),"pag-dns",config.Gateway{Enabled:true,Socket:socket})
	if err!=nil { t.Fatal(err) }
	if g.ID!="pag-dns" || !g.Healthy { t.Fatalf("unexpected gateway: %+v",g) }
}

func TestControlServerUnhealthy(t *testing.T) {
	socket:=filepath.Join(t.TempDir(),"control.sock")
	ln,err:=(ControlServer{ID:"pag-dns",Version:"0.1.0",Socket:socket,Healthy:func()bool{return false}}).Serve()
	if err!=nil { t.Fatal(err) }
	defer ln.Close()
	if _,err:=Discover(context.Background(),"pag-dns",config.Gateway{Enabled:true,Socket:socket}); err==nil {
		t.Fatal("unhealthy gateway discovered as healthy")
	}
}
