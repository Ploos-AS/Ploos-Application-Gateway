package gateway

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
)

type ControlServer struct {
	ID      string
	Version string
	Socket  string
	Healthy func() bool
}

func (s ControlServer) Serve() (net.Listener, error) {
	if s.ID == "" || s.Socket == "" { return nil, fmt.Errorf("gateway control id and socket are required") }
	_ = os.Remove(s.Socket)
	ln, err := net.Listen("unix", s.Socket)
	if err != nil { return nil, err }
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
	var req map[string]string
	if err := json.NewDecoder(c).Decode(&req); err != nil || req["op"] != "hello" { return }
	healthy := true
	if s.Healthy != nil { healthy = s.Healthy() }
	_ = json.NewEncoder(c).Encode(Hello{ID:s.ID, Version:s.Version, Healthy:healthy})
}
