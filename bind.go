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
	Catch func(*error)
)

func (l *logger) Bind(data map[string]any) (Print, Debug, Check, Catch) {
	return func(mesg string, args ...any) {
			l.ch <- &item{
				when: time.Now(),
				mesg: fmt.Sprintf(mesg, args...),
				data: data,
			}
		}, func(mesg string, args ...any) {
			l.ch <- &item{
				when:  time.Now(),
				mesg:  fmt.Sprintf(mesg, args...),
				data:  data,
				debug: true,
			}
		}, func(e any, ntfy ...any) {
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
				panic(fmt.Errorf("%v", e))
			}
		}, func(err *error) {
			switch e := recover().(type) {
			case nil:
				return
			case error:
				*err = e
				li := item{
					when: time.Now(),
					mesg: e.Error(),
					data: data,
				}
				li.Trace()
				l.ch <- &li
			default:
				panic(fmt.Errorf("expect nil or error, got %T", e))
			}
		}
}
