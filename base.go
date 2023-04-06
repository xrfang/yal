package yal

import (
	"fmt"
	"sync"
	"time"
)

type (
	Handler interface {
		Debug() bool
		Done()
		Emit(*item)
		Name() string
	}
	Logger struct {
		ch chan *item
		hm map[string]Handler
		sync.Mutex
	}
)

func NewLogger(qlen int) *Logger {
	l := Logger{ch: make(chan *item, qlen), hm: make(map[string]Handler)}
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
				if !li.debug && h.Debug() {
					h.Emit(li)
				}
			}
		}
	}()
	return &l
}

func (l *Logger) AddHandler(h Handler) {
	l.Lock()
	defer l.Unlock()
	l.hm[h.Name()] = h
}

func (l *Logger) DelHandler(name string) Handler {
	l.Lock()
	defer l.Unlock()
	h := l.hm[name]
	delete(l.hm, name)
	return h
}

func (l *Logger) Log(data map[string]any, mesg string, args ...any) {
	l.ch <- &item{
		when:  time.Now(),
		mesg:  fmt.Sprintf(mesg, args...),
		data:  data,
		debug: false,
	}
}

func (l *Logger) Dbg(data map[string]any, mesg string, args ...any) {
	l.ch <- &item{
		when:  time.Now(),
		mesg:  fmt.Sprintf(mesg, args...),
		data:  data,
		debug: true,
	}
}
