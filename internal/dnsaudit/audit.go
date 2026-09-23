package dnsaudit

import (
	"encoding/json"
	"io"
	"sync"
	"time"
)

type Event struct {
	Time string `json:"time"`
	Decision string `json:"decision"`
	Transport string `json:"transport"`
	QType uint16 `json:"qtype,omitempty"`
	Reason string `json:"reason,omitempty"`
}
type Logger struct { mu sync.Mutex; w io.Writer; now func() time.Time }
func New(w io.Writer) *Logger { return &Logger{w: w, now: time.Now} }
func (l *Logger) Log(decision, transport string, qtype uint16, reason string) {
	if l == nil || l.w == nil { return }
	e := Event{Time: l.now().UTC().Format(time.RFC3339), Decision: decision, Transport: transport, QType: qtype, Reason: reason}
	l.mu.Lock()
	defer l.mu.Unlock()
	_ = json.NewEncoder(l.w).Encode(e)
}
