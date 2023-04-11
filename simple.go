package yal

import "io"

type simpleHandler struct {
	io.Writer
}

func (sh simpleHandler) Emit(li LogItem) {
	li.flush(sh)
}

func NewSimpleLogger(w io.Writer) *logger {
	sh := simpleHandler{w}
	return NewLogger(Options{}, sh)
}
