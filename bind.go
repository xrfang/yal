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

func (l *logger) Log(props ...any) Emitter {
	attr := parse(props...)
	return func(mesg string, args ...any) {
		mesg, data := format(attr, mesg, args...)
		item := LogItem{
			When: time.Now(),
			Mesg: mesg,
			Attr: data,
		}
		st := trace(false)
		if len(st) > 0 {
			item.Attr["~src~"] = st[0]
		}
		if l.Filter != nil {
			l.Filter(&item)
		}
		l.Emit(item)
	}
}

func (l *logger) Dbg(props ...any) Emitter {
	attr := parse(props...)
	return func(mesg string, args ...any) {
		if !l.Debug {
			return
		}
		mesg, data := format(attr, mesg, args...)
		item := LogItem{
			When: time.Now(),
			Mesg: mesg,
			Attr: data,
		}
		st := trace(false)
		if len(st) > 0 {
			item.Attr["~src~"] = st[0]
		}
		if l.Filter != nil {
			l.Filter(&item)
		}
		l.Emit(item)
	}
}

func (l *logger) Check() Checker {
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
		}
	}
}

func (l *logger) Catch(props ...any) Catcher {
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
			e = fmt.Errorf("yal.Catcher: unexpected data type %T", v)
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
		if l.Filter != nil {
			l.Filter(&li)
		}
		l.Emit(li)
	}
}
