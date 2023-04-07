package yal

import (
	"fmt"
	"time"
)

const queueLen = 64

type (
	Handler interface {
		Done()
		Emit(*LogItem)
		Name() string
		SetLevel(int)
		SetRotate(int, int)
	}
	logger struct {
		ch chan *LogItem
		lh Handler
	}
)

func NewLogger(h Handler) *logger {
	l := logger{ch: make(chan *LogItem, queueLen), lh: h}
	go func() {
		for {
			li := <-l.ch
			if li == nil {
				l.lh.Done()
				break
			}
			l.lh.Emit(li)
		}
	}()
	return &l
}

func (l *logger) Log(data map[string]any, mesg string, args ...any) {
	l.ch <- &LogItem{
		When: time.Now(),
		Mesg: fmt.Sprintf(mesg, args...),
		Data: data,
	}
}

func (l *logger) Dbg(data map[string]any, mesg string, args ...any) {
	l.ch <- &LogItem{
		When:  time.Now(),
		Mesg:  fmt.Sprintf(mesg, args...),
		Data:  data,
		Level: 1,
	}
}
