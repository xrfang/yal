package yal

import (
	"io"
	"reflect"
)

const badKey = "!BADKEY"

type (
	Handler interface {
		Emit(LogItem)
	}
	Options struct {
		Trace  byte
		Debug  bool
		Filter func(*LogItem)
	}
)

func Debug(on bool) {
	opt.Debug = on
}

func Trace(mode byte) {
	if mode < 2 {
		opt.Trace = mode
	} else {
		opt.Trace = 2
	}
}

func Peek(w io.Writer) {
	peek = w
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

var (
	opt  Options
	peek io.Writer
	lh   Handler
	skip []string
)

func init() {
	skip = []string{"runtime.", reflect.TypeOf(opt).PkgPath() + "."}
}
