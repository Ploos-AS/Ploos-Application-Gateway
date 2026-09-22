// pag-dns is PAG's DNS application gateway prototype.
package main

import (
	"encoding/binary"
	"flag"
	"io"
	"log"
	"net"
	"os"
	"time"
)

func main() {
	listen := flag.String("listen", "127.0.0.1:5353", "UDP/TCP listen address")
	upstream := flag.String("upstream", "1.1.1.1:53", "DNS upstream address")
	flag.Parse()

	go serveUDP(*listen, *upstream)
	serveTCP(*listen, *upstream)
}

func serveUDP(addr, upstream string) {
	pc, err := net.ListenPacket("udp", addr)
	if err != nil { log.Fatal(err) }
	defer pc.Close()
	buf := make([]byte, 4096)
	for {
		n, peer, err := pc.ReadFrom(buf)
		if err != nil { log.Fatal(err) }
		q := append([]byte(nil), buf[:n]...)
		if !validDNS(q) { continue }
		go func() {
			c, err := net.DialTimeout("udp", upstream, 2*time.Second)
			if err != nil { return }
			defer c.Close()
			_ = c.SetDeadline(time.Now().Add(3*time.Second))
			if _, err = c.Write(q); err != nil { return }
			r := make([]byte, 4096)
			n, err := c.Read(r)
			if err == nil && validDNS(r[:n]) { _, _ = pc.WriteTo(r[:n], peer) }
		}()
	}
}

func serveTCP(addr, upstream string) {
	ln, err := net.Listen("tcp", addr)
	if err != nil { log.Fatal(err) }
	defer ln.Close()
	for {
		c, err := ln.Accept()
		if err != nil { log.Fatal(err) }
		go handleTCP(c, upstream)
	}
}

func handleTCP(client net.Conn, upstream string) {
	defer client.Close()
	_ = client.SetDeadline(time.Now().Add(5*time.Second))
	var hdr [2]byte
	if _, err := io.ReadFull(client, hdr[:]); err != nil { return }
	n := int(binary.BigEndian.Uint16(hdr[:]))
	if n < 12 || n > 4096 { return }
	q := make([]byte, n)
	if _, err := io.ReadFull(client, q); err != nil || !validDNS(q) { return }

	up, err := net.DialTimeout("tcp", upstream, 2*time.Second)
	if err != nil { return }
	defer up.Close()
	_ = up.SetDeadline(time.Now().Add(5*time.Second))
	if _, err = up.Write(append(hdr[:], q...)); err != nil { return }
	if _, err = io.ReadFull(up, hdr[:]); err != nil { return }
	rn := int(binary.BigEndian.Uint16(hdr[:]))
	if rn < 12 || rn > 4096 { return }
	r := make([]byte, rn)
	if _, err = io.ReadFull(up, r); err != nil || !validDNS(r) { return }
	_, _ = client.Write(append(hdr[:], r...))
}

func validDNS(m []byte) bool {
	if len(m) < 12 { return false }
	qd := binary.BigEndian.Uint16(m[4:6])
	// M2 prototype accepts exactly one question and rejects malformed/minimal abuse.
	return qd == 1
}

func init() {
	log.SetOutput(os.Stderr)
	log.SetPrefix("pag-dns: ")
}
