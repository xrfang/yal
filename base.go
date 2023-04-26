package yal

import (
	"io"
	"reflect"
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

var (
	opt  Options
	peek io.Writer
	lh   Handler
	skip []string
)

func init() {
	skip = []string{"runtime.", reflect.TypeOf(opt).PkgPath() + "."}
}
