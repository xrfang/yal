package yal

import (
	"fmt"
	"io"
	"regexp"
	"strings"
)

const (
	badKey = "!BADKEY"
	badVal = "!BADVAL"
)

type (
	Switch  int8
	Handler interface {
		Emit(LogItem)
	}
	Options struct {
		Trace  bool
		Debug  bool
		Filter func(*LogItem)
	}
)

func Debug(on bool) {
	opt.Debug = on
}

func Trace(on bool) {
	opt.Trace = on
}

func Peek(w io.Writer) {
	if w == nil {
		peek = io.Discard
	} else {
		peek = w
	}
}

func Filter(f func(*LogItem)) {
	opt.Filter = f
}

func Setup(f func() (Handler, error)) error {
	h, err := f()
	if err != nil {
		return err
	}
	lh = h
	return nil
}

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

var (
	opt  Options
	mtx  *regexp.Regexp
	peek io.Writer
	lh   Handler
)

func init() {
	mtx = regexp.MustCompile(`{{(\w+)}}`)
}
