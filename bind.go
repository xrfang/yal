package yal

import (
	"errors"
	"fmt"
	"time"
)

type (
	Print func(string, ...any)
	Debug func(string, ...any)
	Check func(any, ...any)
	Catch func(*error, ...any)
)

func (l *logger) Printer(data map[string]any) Print {
	return func(mesg string, args ...any) {
		mesg, data := format(data, mesg, args...)
		l.ch <- &LogItem{
			When: time.Now(),
			Mesg: mesg,
			Data: data,
		}
	}
}

func (l *logger) Debugger(data map[string]any) Debug {
	return func(mesg string, args ...any) {
		mesg, data := format(data, mesg, args...)
		l.ch <- &LogItem{
			When:  time.Now(),
			Mesg:  mesg,
			Data:  data,
			Level: 1,
		}
	}
}

func (l *logger) Checker() Check {
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

func (l *logger) Catcher(data map[string]any) Catch {
	return func(err *error, attr ...any) {
		switch e := recover().(type) {
		case nil:
			return
		case error:
			if err != nil {
				*err = e
			}
			mesg, data := format(data, e.Error(), attr...)
			li := LogItem{
				When: time.Now(),
				Mesg: mesg,
				Data: data,
			}
			li.Trace()
			l.ch <- &li
		default:
			panic(fmt.Errorf("expect nil or error, got %T", e))
		}
	}
}
