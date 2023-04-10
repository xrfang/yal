package yal

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	queueLen = 64
	badKey   = "!BADKEY"
)

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

func format(data map[string]any, mesg string, args ...any) (string, map[string]any) {
	attr := map[string]any{}
parse:
	for i := 0; i < len(args); i += 2 {
		if i+1 >= len(args) {
			break
		}
		switch args[i].(type) {
		case string:
			attr[args[i].(string)] = args[i+1]
		default:
			attr[badKey] = args[i]
			break parse
		}
	}
	ms := mtx.FindAllStringSubmatch(mesg, -1)
	for _, m := range ms {
		subst := attr[m[1]]
		if subst != nil {
			s := fmt.Sprintf("%v", subst)
			mesg = strings.ReplaceAll(mesg, m[0], s)
			delete(attr, m[1])
		}
	}
	if data == nil {
		return mesg, attr
	}
	for k, v := range attr {
		data[k] = v
	}
	return mesg, data
}

var mtx *regexp.Regexp

func init() {
	mtx = regexp.MustCompile(`{{(\w+)}}`)
}
