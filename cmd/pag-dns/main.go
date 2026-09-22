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

	"github.com/Ploos-AS/Ploos-Application-Gateway/internal/dnswire"
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
		if dnswire.ValidateQuery(q) != nil { continue }
		go func() {
			c, err := net.DialTimeout("udp", upstream, 2*time.Second)
			if err != nil { return }
			defer c.Close()
			_ = c.SetDeadline(time.Now().Add(3*time.Second))
			if _, err = c.Write(q); err != nil { return }
			r := make([]byte, 4096)
			n, err := c.Read(r)
			if err == nil && dnswire.ValidateResponse(q, r[:n]) == nil { _, _ = pc.WriteTo(r[:n], peer) }
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
	if _, err := io.ReadFull(client, q); err != nil || dnswire.ValidateQuery(q) != nil { return }

	up, err := net.DialTimeout("tcp", upstream, 2*time.Second)
	if err != nil { return }
	defer up.Close()
	_ = up.SetDeadline(time.Now().Add(5*time.Second))
	if _, err = up.Write(append(hdr[:], q...)); err != nil { return }
	if _, err = io.ReadFull(up, hdr[:]); err != nil { return }
	rn := int(binary.BigEndian.Uint16(hdr[:]))
	if rn < 12 || rn > 4096 { return }
	r := make([]byte, rn)
	if _, err = io.ReadFull(up, r); err != nil || dnswire.ValidateResponse(q, r) != nil { return }
	_, _ = client.Write(append(hdr[:], r...))
}

func init() {
	log.SetOutput(os.Stderr)
	log.SetPrefix("pag-dns: ")
}
