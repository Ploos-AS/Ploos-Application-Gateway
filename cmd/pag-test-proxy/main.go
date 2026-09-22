// pag-test-proxy is integration-test infrastructure, not a production gateway.
package main

import (
	"flag"
	"io"
	"log"
	"net"
)

func main() {
	listen := flag.String("listen", "", "TCP listen address")
	upstream := flag.String("upstream", "", "TCP upstream address")
	flag.Parse()
	if *listen == "" || *upstream == "" { log.Fatal("-listen and -upstream are required") }

	ln, err := net.Listen("tcp", *listen)
	if err != nil { log.Fatal(err) }
	defer ln.Close()

	for {
		client, err := ln.Accept()
		if err != nil { log.Fatal(err) }
		go proxy(client, *upstream)
	}
}

func proxy(client net.Conn, upstreamAddr string) {
	defer client.Close()
	upstream, err := net.Dial("tcp", upstreamAddr)
	if err != nil { return }
	defer upstream.Close()

	done := make(chan struct{}, 2)
	go func() { _, _ = io.Copy(upstream, client); done <- struct{}{} }()
	go func() { _, _ = io.Copy(client, upstream); done <- struct{}{} }()
	<-done
}
