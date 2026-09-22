package limit

import (
	"net"
	"sync"
	"time"
)

type bucket struct {
	tokens float64
	last   time.Time
	active int
}

type Limiter struct {
	mu sync.Mutex
	clients map[string]*bucket
	rate float64
	burst float64
	maxActive int
	now func() time.Time
}

func New(rate float64, burst, maxActive int) *Limiter {
	return &Limiter{clients:map[string]*bucket{}, rate:rate, burst:float64(burst), maxActive:maxActive, now:time.Now}
}

func (l *Limiter) key(addr net.Addr) string {
	host, _, err := net.SplitHostPort(addr.String())
	if err == nil { return host }
	return addr.String()
}

func (l *Limiter) Allow(addr net.Addr) bool {
	l.mu.Lock(); defer l.mu.Unlock()
	k:=l.key(addr); now:=l.now()
	b:=l.clients[k]
	if b==nil { b=&bucket{tokens:l.burst,last:now}; l.clients[k]=b }
	elapsed:=now.Sub(b.last).Seconds()
	b.tokens += elapsed*l.rate
	if b.tokens>l.burst { b.tokens=l.burst }
	b.last=now
	if b.tokens<1 { return false }
	b.tokens--
	return true
}

func (l *Limiter) Acquire(addr net.Addr) bool {
	l.mu.Lock(); defer l.mu.Unlock()
	k:=l.key(addr); b:=l.clients[k]
	if b==nil { b=&bucket{tokens:l.burst,last:l.now()}; l.clients[k]=b }
	if l.maxActive>0 && b.active>=l.maxActive { return false }
	b.active++
	return true
}

func (l *Limiter) Release(addr net.Addr) {
	l.mu.Lock(); defer l.mu.Unlock()
	if b:=l.clients[l.key(addr)]; b!=nil && b.active>0 { b.active-- }
}
