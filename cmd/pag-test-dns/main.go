// pag-test-dns is deterministic integration-test infrastructure.
package main

import (
	"encoding/binary"
	"flag"
	"log"
	"net"
)

func main() {
	listen:=flag.String("listen","198.51.100.2:5353","UDP listen address")
	flag.Parse()
	pc,err:=net.ListenPacket("udp",*listen)
	if err!=nil { log.Fatal(err) }
	defer pc.Close()
	buf:=make([]byte,4096)
	for {
		n,peer,err:=pc.ReadFrom(buf)
		if err!=nil { log.Fatal(err) }
		if n<12 { continue }
		q:=append([]byte(nil),buf[:n]...)
		q[2]|=0x80
		q[3]|=0x80
		// No answers are needed: the integration test validates the proxy path,
		// transaction/question matching and response delivery.
		binary.BigEndian.PutUint16(q[6:8],0)
		_,_=pc.WriteTo(q,peer)
	}
}
