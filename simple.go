package yal

import "io"

type SimpleHandler struct {
	io.Writer
}

func (sh SimpleHandler) Name() string {
	return "simple"
}

func (sh SimpleHandler) Done() {}

func (sh SimpleHandler) Emit(li *LogItem) {
	li.Flush(sh)
}
