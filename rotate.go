package yal

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const queueLen = 32

type (
	fileHandle struct {
		f *os.File
		t time.Time
	}
	rotatedHandler struct {
		dir   string
		split int
		keep  int
		ch    chan LogItem
		fhm   sync.Map //map[string]*fileHandle
	}
)

func (rh *rotatedHandler) Emit(li LogItem) {
	rh.ch <- li
}

func (rh *rotatedHandler) baseName(li LogItem) string {
	base, _ := li.Attr["basename"].(string)
	if base == "" {
		return "log"
	}
	return base
}

func (rh *rotatedHandler) getHandle(li LogItem) (*os.File, error) {
	base := rh.baseName(li)
	v, _ := rh.fhm.Load(base)
	if v == nil {
		fp := filepath.Join(rh.dir, base)
		//TODO: rotate fp!
		f, err := os.OpenFile(fp, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			return nil, err
		}
		rh.fhm.Store(base, &fileHandle{f: f, t: time.Now()})
		return f, nil
	}
	fh := v.(*fileHandle)
	fh.t = time.Now()
	rh.fhm.Store(base, fh)
	return fh.f, nil
}

func (rh *rotatedHandler) ingest() {
	for {
		li := <-rh.ch
		f, err := rh.getHandle(li)
		if err != nil {
			fmt.Fprintf(os.Stderr, "RotatedHandler.ingest: %v\n", err)
			continue
		}
		li.Flush(f)
	}
}

func (rh *rotatedHandler) rotate() {
	fmt.Println("TODO: log rotating...")
}

func (rh *rotatedHandler) Close() error { return nil }

func NewRotatedLogger(dir string, split, keep int) (*logger, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0777); err != nil {
		return nil, err
	}
	if split <= 0 {
		split = 10 * 1024 * 1024
	}
	if keep <= 0 {
		keep = 10
	}
	rh := rotatedHandler{
		dir:   dir,
		split: split,
		keep:  keep,
		ch:    make(chan LogItem, queueLen),
	}
	go rh.ingest()
	return NewLogger(Options{}, &rh), nil
}
