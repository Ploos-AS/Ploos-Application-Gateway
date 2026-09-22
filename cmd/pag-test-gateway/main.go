// pag-test-gateway is integration-test infrastructure, not a protocol gateway.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"log"
	"net"
	"os"
)

type hello struct {
	ID string `json:"id"`
	Version string `json:"version"`
	Healthy bool `json:"healthy"`
}

func main() {
	socket := flag.String("socket", "", "Unix socket")
	id := flag.String("id", "pag-test", "gateway id")
	healthy := flag.Bool("healthy", true, "health state")
	flag.Parse()
	if *socket == "" { log.Fatal("-socket is required") }
	_ = os.Remove(*socket)
	ln, err := net.Listen("unix", *socket)
	if err != nil { log.Fatal(err) }
	defer ln.Close()
	for {
		c, err := ln.Accept()
		if err != nil { log.Fatal(err) }
		go func(conn net.Conn) {
			defer conn.Close()
			var req map[string]string
			if err := json.NewDecoder(bufio.NewReader(conn)).Decode(&req); err != nil { return }
			if req["op"] != "hello" { return }
			_ = json.NewEncoder(conn).Encode(hello{ID:*id, Version:"0.0.0-test", Healthy:*healthy})
		}(c)
	}
}
