package yal

import "io"

type SimpleHandler struct {
	io.Writer
}

func (sh SimpleHandler) Emit(li LogItem) {
	li.Flush(sh)
}

func (sh SimpleHandler) Close() error { return nil }

func NewSimpleLogger(w io.Writer) *logger {
	sh := SimpleHandler{w}
	return NewLogger(Options{}, sh)
}
