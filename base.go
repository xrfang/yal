package yal

import (
	"fmt"
	"sync"
	"time"
)

const queueLen = 64

type (
	Handler interface {
		Done()
		Emit(*item)
		Name() string
	}
	logger struct {
		ch  chan *item
		dbg bool
		hm  map[string]Handler
		sync.Mutex
	}
)

func NewLogger(dbg bool) *logger {
	l := logger{
		ch:  make(chan *item, queueLen),
		dbg: dbg,
		hm:  make(map[string]Handler),
	}
	go func() {
		for {
			li := <-l.ch
			if li == nil {
				for _, h := range l.hm {
					h.Done()
				}
				break
			}
			for _, h := range l.hm {
				if !li.debug && l.dbg {
					h.Emit(li)
				}
			}
		}
	}()
	return &l
}

func (l *logger) GetDebug() bool {
	return l.dbg
}

func (l *logger) SetDebug(dbg bool) {
	l.dbg = dbg
}

func (l *logger) AddHandler(h Handler) {
	l.Lock()
	defer l.Unlock()
	l.hm[h.Name()] = h
}

func (l *logger) DelHandler(name string) Handler {
	l.Lock()
	defer l.Unlock()
	h := l.hm[name]
	delete(l.hm, name)
	return h
}

func (l *logger) Log(data map[string]any, mesg string, args ...any) {
	l.ch <- &item{
		when:  time.Now(),
		mesg:  fmt.Sprintf(mesg, args...),
		data:  data,
		debug: false,
	}
}

func (l *logger) Dbg(data map[string]any, mesg string, args ...any) {
	l.ch <- &item{
		when:  time.Now(),
		mesg:  fmt.Sprintf(mesg, args...),
		data:  data,
		debug: true,
	}
}
