// pag-test-dns is deterministic integration-test infrastructure.
package main

import (
	"encoding/binary"
	"flag"
	"io"
	"log"
	"net"
)

func response(q []byte) []byte {
	r:=append([]byte(nil),q...)
	if len(r)>=12 {
		r[2]|=0x80
		r[3]|=0x80
		binary.BigEndian.PutUint16(r[6:8],0)
	}
	return r
}

func main() {
	listen:=flag.String("listen","198.51.100.2:5353","UDP/TCP listen address")
	truncateUDP:=flag.Bool("truncate-udp",false,"set TC on UDP replies to exercise TCP fallback")
	flag.Parse()

	pc,err:=net.ListenPacket("udp",*listen)
	if err!=nil { log.Fatal(err) }
	defer pc.Close()
	ln,err:=net.Listen("tcp",*listen)
	if err!=nil { log.Fatal(err) }
	defer ln.Close()

	go func() {
		buf:=make([]byte,4096)
		for {
			n,peer,err:=pc.ReadFrom(buf); if err!=nil{return}
			r:=response(buf[:n])
			if *truncateUDP && len(r)>=4 { r[2]|=0x02 }
			_,_=pc.WriteTo(r,peer)
		}
	}()

	for {
		c,err:=ln.Accept(); if err!=nil{log.Fatal(err)}
		go func(c net.Conn) {
			defer c.Close()
			var h [2]byte
			if _,err:=io.ReadFull(c,h[:]);err!=nil{return}
			n:=int(binary.BigEndian.Uint16(h[:]))
			if n<12||n>4096{return}
			q:=make([]byte,n)
			if _,err:=io.ReadFull(c,q);err!=nil{return}
			r:=response(q)
			binary.BigEndian.PutUint16(h[:],uint16(len(r)))
			_,_=c.Write(append(h[:],r...))
		}(c)
	}
}
