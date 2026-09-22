package gateway

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

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


func TestControlServerRefusesNonSocketPath(t *testing.T) {
	path:=filepath.Join(t.TempDir(),"control.sock")
	if err:=os.WriteFile(path,[]byte("do not replace"),0600);err!=nil{t.Fatal(err)}
	if _,err:=(ControlServer{ID:"pag-dns",Version:"0.1.0",Socket:path}).Serve();err==nil{
		t.Fatal("non-socket path was replaced")
	}
}

func TestControlServerSocketPermissions(t *testing.T) {
	path:=filepath.Join(t.TempDir(),"nested","control.sock")
	ln,err:=(ControlServer{ID:"pag-dns",Version:"0.1.0",Socket:path}).Serve()
	if err!=nil{t.Fatal(err)}
	defer ln.Close()
	fi,err:=os.Stat(path);if err!=nil{t.Fatal(err)}
	if fi.Mode().Perm()!=0660{t.Fatalf("socket permissions=%o want 660",fi.Mode().Perm())}
}

func TestControlServerStalledClientTimesOut(t *testing.T) {
	path:=filepath.Join(t.TempDir(),"control.sock")
	ln,err:=(ControlServer{ID:"pag-dns",Version:"0.1.0",Socket:path}).Serve()
	if err!=nil{t.Fatal(err)}
	defer ln.Close()
	c,err:=net.Dial("unix",path);if err!=nil{t.Fatal(err)}
	defer c.Close()
	time.Sleep(2200*time.Millisecond)
	_ = c.SetReadDeadline(time.Now().Add(500*time.Millisecond))
	var b [1]byte
	if _,err=c.Read(b[:]);err==nil{t.Fatal("stalled control connection remained usable")}
}


func TestControlServerCloseRemovesSocket(t *testing.T) {
	path:=filepath.Join(t.TempDir(),"control.sock")
	ln,err:=(ControlServer{ID:"pag-dns",Version:"0.1.0",Socket:path}).Serve()
	if err!=nil{t.Fatal(err)}
	if _,err:=os.Stat(path);err!=nil{t.Fatal(err)}
	if err:=ln.Close();err!=nil{t.Fatal(err)}
	if _,err:=os.Stat(path);!os.IsNotExist(err){t.Fatalf("control socket remains after close: %v",err)}
}
