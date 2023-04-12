package yal

import (
	"errors"
	"fmt"
	"time"
)

type (
	Emitter func(string, ...any)
	Checker func(any, ...any)
	Catcher func(*error, ...any)
)

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

func ErrChecker() Checker {
	return func(e any, ntfy ...any) {
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
			panic(fmt.Errorf("yal.Checker: expect <error> or <bool>, got <%T>", e))
		}
	}
}

func NewCatcher(props ...any) Catcher {
	attr := parse(props...)
	return func(err *error, args ...any) {
		var e error
		switch v := recover().(type) {
		case nil:
			return
		case string:
			e = errors.New(v)
		case error:
			e = v
		default:
			e = fmt.Errorf("%v", v)
		}
		if err != nil {
			*err = e
		}
		mesg, data := format(attr, e.Error(), args...)
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
}
