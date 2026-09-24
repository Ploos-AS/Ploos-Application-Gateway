package limit

import (
	"net"
	"testing"
	"time"
)

func addr(s string) net.Addr { a, _ := net.ResolveUDPAddr("udp", s); return a }

func TestBurstLimit(t *testing.T) {
	l := New(1, 2, 0)
	a := addr("192.0.2.1:1234")
	if !l.Allow(a) || !l.Allow(a) {
		t.Fatal("initial burst rejected")
	}
	if l.Allow(a) {
		t.Fatal("burst limit not enforced")
	}
}

func TestClientsIndependent(t *testing.T) {
	l := New(1, 1, 0)
	a := addr("192.0.2.1:1")
	b := addr("192.0.2.2:1")
	if !l.Allow(a) || !l.Allow(b) {
		t.Fatal("clients must have independent buckets")
	}
}

func TestRefill(t *testing.T) {
	l := New(1, 1, 0)
	now := time.Unix(0, 0)
	l.now = func() time.Time { return now }
	a := addr("192.0.2.1:1")
	if !l.Allow(a) || l.Allow(a) {
		t.Fatal("unexpected initial bucket behavior")
	}
	now = now.Add(time.Second)
	if !l.Allow(a) {
		t.Fatal("token did not refill")
	}
}

func TestConcurrency(t *testing.T) {
	l := New(1, 1, 1)
	a := addr("192.0.2.1:1")
	if !l.Acquire(a) {
		t.Fatal("first acquire rejected")
	}
	if l.Acquire(a) {
		t.Fatal("concurrency limit not enforced")
	}
	l.Release(a)
	if !l.Acquire(a) {
		t.Fatal("release did not restore capacity")
	}
}

func TestIdleClientsExpire(t *testing.T) {
	l := New(1, 1, 0)
	now := time.Unix(0, 0)
	l.now = func() time.Time { return now }
	l.idleTTL = time.Minute
	a := addr("192.0.2.1:1")
	if !l.Allow(a) {
		t.Fatal("initial request rejected")
	}
	if len(l.clients) != 1 {
		t.Fatal("client state not created")
	}
	now = now.Add(2 * time.Minute)
	b := addr("192.0.2.2:1")
	if !l.Allow(b) {
		t.Fatal("new client rejected")
	}
	if _, ok := l.clients["192.0.2.1"]; ok {
		t.Fatal("idle client state not expired")
	}
}

func TestActiveClientDoesNotExpire(t *testing.T) {
	l := New(1, 1, 1)
	now := time.Unix(0, 0)
	l.now = func() time.Time { return now }
	l.idleTTL = time.Minute
	a := addr("192.0.2.1:1")
	if !l.Acquire(a) {
		t.Fatal("acquire failed")
	}
	now = now.Add(2 * time.Minute)
	_ = l.Allow(addr("192.0.2.2:1"))
	if _, ok := l.clients["192.0.2.1"]; !ok {
		t.Fatal("active client expired")
	}
}


func TestClientStateLimitFailsClosed(t *testing.T) {
	l := New(1, 1, 1)
	l.maxClients = 2
	a := addr("192.0.2.1:1")
	b := addr("192.0.2.2:1")
	c := addr("192.0.2.3:1")
	if !l.Allow(a) || !l.Allow(b) {
		t.Fatal("clients within state limit rejected")
	}
	if l.Allow(c) {
		t.Fatal("new client admitted after state limit reached")
	}
	if len(l.clients) != 2 {
		t.Fatalf("client state exceeded limit: %d", len(l.clients))
	}
}

func TestClientStateLimitPreservesActiveClient(t *testing.T) {
	l := New(1, 1, 1)
	l.maxClients = 1
	a := addr("192.0.2.1:1")
	if !l.Acquire(a) {
		t.Fatal("active client rejected")
	}
	if l.Acquire(addr("192.0.2.2:1")) {
		t.Fatal("new client admitted while state table full")
	}
	if _, ok := l.clients["192.0.2.1"]; !ok {
		t.Fatal("active client state was evicted")
	}
	l.Release(a)
}
