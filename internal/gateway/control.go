package gateway

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"
)

type ControlServer struct {
	ID      string
	Version string
	Socket  string
	Healthy func() bool
}

func (s ControlServer) Serve() (net.Listener, error) {
	if s.ID == "" || s.Socket == "" { return nil, fmt.Errorf("gateway control id and socket are required") }
	dir:=filepath.Dir(s.Socket)
	if err:=os.MkdirAll(dir,0750); err!=nil { return nil, fmt.Errorf("create control socket directory: %w",err) }
	if fi,err:=os.Lstat(s.Socket); err==nil {
		if fi.Mode()&os.ModeSocket==0 { return nil, fmt.Errorf("refusing to replace non-socket control path %q",s.Socket) }
		if err:=os.Remove(s.Socket); err!=nil { return nil,err }
	} else if !os.IsNotExist(err) { return nil,err }

	ln, err := net.Listen("unix", s.Socket)
	if err != nil { return nil, err }
	if err:=os.Chmod(s.Socket,0660); err!=nil { ln.Close(); _=os.Remove(s.Socket); return nil,err }
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil { return }
			go s.handle(c)
		}
	}()
	return ln, nil
}

func (s ControlServer) handle(c net.Conn) {
	defer c.Close()
	_ = c.SetDeadline(time.Now().Add(2*time.Second))
	var req map[string]string
	if err := json.NewDecoder(c).Decode(&req); err != nil || req["op"] != "hello" { return }
	healthy := true
	if s.Healthy != nil { healthy = s.Healthy() }
	_ = json.NewEncoder(c).Encode(Hello{ID:s.ID, Version:s.Version, Healthy:healthy})
}
