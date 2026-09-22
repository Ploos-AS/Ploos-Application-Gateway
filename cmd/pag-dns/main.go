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

	"github.com/Ploos-AS/Ploos-Application-Gateway/internal/dnsaudit"
	"github.com/Ploos-AS/Ploos-Application-Gateway/internal/dnswire"
	"github.com/Ploos-AS/Ploos-Application-Gateway/internal/gateway"
	"github.com/Ploos-AS/Ploos-Application-Gateway/internal/limit"
	"github.com/Ploos-AS/Ploos-Application-Gateway/internal/version"
)

func main() {
	listen := flag.String("listen", "127.0.0.1:5353", "UDP/TCP listen address")
	upstream := flag.String("upstream", "1.1.1.1:53", "DNS upstream address")
	control := flag.String("control", "/run/pag/pag-dns.sock", "PAG control Unix socket")
	rate := flag.Float64("rate", 50, "queries per second per client")
	burst := flag.Int("burst", 100, "per-client query burst")
	maxTCP := flag.Int("max-tcp-per-client", 16, "maximum concurrent TCP sessions per client")
	auditEnabled := flag.Bool("audit", true, "emit privacy-minimal structured DNS audit events")
	flag.Parse()

	controlListener, err := (gateway.ControlServer{ID:"pag-dns", Version:version.Version, Socket:*control}).Serve()
	if err != nil { log.Fatalf("control socket: %v", err) }
	defer controlListener.Close()

	var audit *dnsaudit.Logger
	if *auditEnabled { audit = dnsaudit.New(os.Stderr) }
	_ = audit

	udpLimit := limit.New(*rate, *burst, 0)
	tcpLimit := limit.New(*rate, *burst, *maxTCP)
	go serveUDP(*listen, *upstream, udpLimit, audit)
	serveTCP(*listen, *upstream, tcpLimit, audit)
}

func serveUDP(addr, upstream string, limiter *limit.Limiter, audit *dnsaudit.Logger) {
	pc, err := net.ListenPacket("udp", addr)
	if err != nil { log.Fatal(err) }
	defer pc.Close()
	buf := make([]byte, 4096)
	for {
		n, peer, err := pc.ReadFrom(buf)
		if err != nil { log.Fatal(err) }
		q := append([]byte(nil), buf[:n]...)
		if !limiter.Allow(peer) { audit.Log("deny","udp",0,"rate_limit"); continue }
		if err:=dnswire.ValidateQueryPolicy(q, dnswire.DefaultPolicy()); err != nil {
			var qt uint16
			if parsed,e:=dnswire.ParseQuestion(q); e==nil { qt=parsed.Type }
			audit.Log("deny","udp",qt,err.Error())
			continue
		}
		go func() {
			c, err := net.DialTimeout("udp", upstream, 2*time.Second)
			if err != nil { return }
			defer c.Close()
			_ = c.SetDeadline(time.Now().Add(3*time.Second))
			if _, err = c.Write(q); err != nil { return }
			r := make([]byte, 4096)
			n, err := c.Read(r)
			if err != nil { return }
			resp := r[:n]
			if dnswire.ValidateResponse(q, resp) != nil { return }
			if dnswire.IsTruncated(resp) {
				resp, err = exchangeTCP(upstream, q)
				if err != nil { return }
			}
			_, _ = pc.WriteTo(resp, peer)
			audit.Log("allow","udp",mustQType(q),"proxied")
		}()
	}
}

func serveTCP(addr, upstream string, limiter *limit.Limiter, audit *dnsaudit.Logger) {
	ln, err := net.Listen("tcp", addr)
	if err != nil { log.Fatal(err) }
	defer ln.Close()
	for {
		c, err := ln.Accept()
		if err != nil { log.Fatal(err) }
		if !limiter.Allow(c.RemoteAddr()) { audit.Log("deny","tcp",0,"rate_limit"); _=c.Close(); continue }
		if !limiter.Acquire(c.RemoteAddr()) { audit.Log("deny","tcp",0,"concurrency_limit"); _=c.Close(); continue }
		go func() { defer limiter.Release(c.RemoteAddr()); handleTCP(c, upstream, audit) }()
	}
}

func exchangeTCP(upstream string, q []byte) ([]byte, error) {
	up, err := net.DialTimeout("tcp", upstream, 2*time.Second)
	if err != nil { return nil, err }
	defer up.Close()
	_ = up.SetDeadline(time.Now().Add(5*time.Second))
	var hdr [2]byte
	binary.BigEndian.PutUint16(hdr[:], uint16(len(q)))
	if _, err = up.Write(append(hdr[:], q...)); err != nil { return nil, err }
	if _, err = io.ReadFull(up, hdr[:]); err != nil { return nil, err }
	rn := int(binary.BigEndian.Uint16(hdr[:]))
	if rn < 12 || rn > 4096 { return nil, io.ErrUnexpectedEOF }
	r := make([]byte, rn)
	if _, err = io.ReadFull(up, r); err != nil { return nil, err }
	if err = dnswire.ValidateResponse(q, r); err != nil { return nil, err }
	return r, nil
}

func handleTCP(client net.Conn, upstream string, audit *dnsaudit.Logger) {
	defer client.Close()
	_ = client.SetDeadline(time.Now().Add(5*time.Second))
	var hdr [2]byte
	if _, err := io.ReadFull(client, hdr[:]); err != nil { return }
	n := int(binary.BigEndian.Uint16(hdr[:]))
	if n < 12 || n > 4096 { return }
	q := make([]byte, n)
	if _, err := io.ReadFull(client, q); err != nil { audit.Log("deny","tcp",0,"read_error"); return }
	if err:=dnswire.ValidateQueryPolicy(q,dnswire.DefaultPolicy()); err!=nil { audit.Log("deny","tcp",mustQType(q),err.Error()); return }

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
	audit.Log("allow","tcp",mustQType(q),"proxied")
}

func mustQType(q []byte) uint16 {
	parsed,err:=dnswire.ParseQuestion(q)
	if err!=nil{return 0}
	return parsed.Type
}

func init() {
	log.SetOutput(os.Stderr)
	log.SetPrefix("pag-dns: ")
}
