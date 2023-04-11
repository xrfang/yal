package yal

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	badKey = "!BADKEY"
	badVal = "!BADVAL"
)

type (
	Handler interface {
		Emit(LogItem)
	}
	Options struct {
		Trace  bool
		Debug  bool
		Filter func(*LogItem)
	}
	logger struct {
		Options
		Handler
	}
)

func NewLogger(o Options, h Handler) *logger {
	return &logger{o, h}
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

func format(data map[string]any, mesg string, args ...any) (string, map[string]any) {
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
