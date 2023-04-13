package yal

import (
	"errors"
	"fmt"
	"regexp"
	"runtime"
	"strings"
	"time"
)

type (
	Emitter func(string, ...any)
	ErrProc = func(error) error
)

func parse(args ...any) map[string]any {
	attr := map[string]any{}
	for i := 0; i < len(args); i += 2 {
		if i+1 >= len(args) {
			break
		}
		switch args[i].(type) {
		case string:
			attr[args[i].(string)] = args[i+1]
		default:
			attr[badKey] = args[i]
			return attr
		}
	}
	return attr
}

func format(prop map[string]any, mesg string, args ...any) (string, map[string]any) {
	attr := parse(args...)
	ms := mtx.FindAllStringSubmatch(mesg, -1)
	for _, m := range ms {
		subst := attr[m[1]]
		if subst != nil {
			s := fmt.Sprintf("%v", subst)
			mesg = strings.ReplaceAll(mesg, m[0], s)
			delete(attr, m[1])
		}
	}
	data := map[string]any{}
	for k, v := range prop {
		data[k] = v
	}
	for k, v := range attr {
		data[k] = v
	}
	return mesg, data
}

func trace(full bool) []string {
	var st []string
	n := 1
	for {
		n++
		pc, file, line, ok := runtime.Caller(n)
		if !ok {
			break
		}
		f := runtime.FuncForPC(pc)
		name := f.Name()
		if strings.HasPrefix(name, "runtime.") {
			continue
		}
		fn := strings.Split(file, "/")
		if len(fn) > 1 {
			file = strings.Join(fn[len(fn)-2:], "/")
		}
		if file == "yal/bind.go" {
			continue
		}
		st = append(st, fmt.Sprintf("(%s:%d) %s", file, line, name))
		if !full {
			break
		}
	}
	return st
}

func NewLogger(props ...any) Emitter {
	attr := parse(props...)
	return func(mesg string, args ...any) {
		mesg, data := format(attr, mesg, args...)
		item := LogItem{
			When: time.Now(),
			Mesg: mesg,
			Attr: data,
		}
		if opt.Trace {
			st := trace(false)
			if len(st) > 0 {
				item.Attr["~src~"] = st[0]
			}
		}
		if opt.Filter != nil {
			opt.Filter(&item)
		}
		if peek != nil {
			item.flush(peek)
		}
		if lh != nil {
			lh.Emit(item)
		}
	}
}

func NewDebugger(props ...any) Emitter {
	attr := parse(props...)
	return func(mesg string, args ...any) {
		if !opt.Debug {
			return
		}
		mesg, data := format(attr, mesg, args...)
		item := LogItem{
			When: time.Now(),
			Mesg: mesg,
			Attr: data,
		}
		if opt.Trace {
			st := trace(false)
			if len(st) > 0 {
				item.Attr["~src~"] = st[0]
			}
		}
		if opt.Filter != nil {
			opt.Filter(&item)
		}
		if peek != nil {
			item.flush(peek)
		}
		if lh != nil {
			lh.Emit(item)
		}
	}
}

func Assert(e any, ntfy ...any) {
	switch v := e.(type) {
	case nil:
	case bool:
		if !v {
			mesg := "assertion failed"
			if len(ntfy) > 0 {
				mesg = ntfy[0].(string)
				if len(ntfy) > 1 {
					mesg = fmt.Sprintf(mesg, ntfy[1:]...)
				}
			}
			panic(errors.New(mesg))
		}
	case error:
		panic(v)
	default:
		panic(fmt.Errorf("yal.Assert: expect <error> or <bool>, got <%T>", e))
	}
}

func Catch(h any, args ...any) {
	var e error
	switch v := recover().(type) {
	case nil:
	case string:
		e = errors.New(v)
	case error:
		e = v
	default:
		e = fmt.Errorf("%v", v)
	}
	switch proc := h.(type) {
	case nil:
	case *error:
		*proc = e
	case ErrProc:
		e = proc(e)
	default:
		panic(fmt.Errorf("yal.Catch: invalid handler <%T>", h))
	}
	if e == nil {
		return
	}
	mesg, data := format(nil, e.Error(), args...)
	li := LogItem{
		When: time.Now(),
		Mesg: mesg,
		Attr: data,
	}
	st := trace(true)
	if len(st) > 0 {
		if li.Attr == nil {
			li.Attr = make(map[string]any)
		}
		li.Attr["callstack"] = st
	}
	if opt.Filter != nil {
		opt.Filter(&li)
	}
	if peek != nil {
		li.flush(peek)
	}
	if lh != nil {
		lh.Emit(li)
	}
}

var mtx *regexp.Regexp

func init() {
	mtx = regexp.MustCompile(`{{(\w+)}}`)
}
