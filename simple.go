package yal

import "io"

type simpleHandler struct {
	io.Writer
}

func (sh simpleHandler) Emit(li LogItem) {
	li.Flush(sh)
}

func (sh simpleHandler) Close() error { return nil }

func NewSimpleLogger(w io.Writer) *logger {
	sh := simpleHandler{w}
	return NewLogger(Options{}, sh)
}
