package yal

import (
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const queueLen = 32

type (
	rotatedHandler struct {
		dir   string
		split int64
		keep  int
		ch    chan LogItem
		fhm   map[string]*os.File
		sync.Mutex
	}
)

func (rh *rotatedHandler) Emit(li LogItem) {
	rh.ch <- li
}

func (rh *rotatedHandler) getHandle(li *LogItem) (string, *os.File, error) {
	rh.Lock()
	defer rh.Unlock()
	base := "log"
	if v := li.popAttr("basename"); v != nil {
		base = v.(string)
	}
	if f := rh.fhm[base]; f != nil {
		return base, f, nil
	}
	fp := filepath.Join(rh.dir, base)
	f, err := os.OpenFile(fp, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return "", nil, err
	}
	rh.fhm[base] = f
	return base, f, nil
}

func (rh *rotatedHandler) delHandle(name string) {
	rh.Lock()
	defer rh.Unlock()
	delete(rh.fhm, name)
}

func (rh *rotatedHandler) ingest() {
	for {
		li := <-rh.ch
		base, f, err := rh.getHandle(&li)
		if onErr(err) {
			continue
		}
		li.flush(f)
		if rh.rotate(f) {
			rh.delHandle(base)
		}
	}
}

func (rh *rotatedHandler) rotate(f *os.File) bool {
	defer func() {
		if e := recover(); e != nil {
			onErr(e.(error))
		}
	}()
	st, err := f.Stat()
	assert(err)
	if st.Size() < rh.split {
		return false
	}
	assert(f.Close())
	fn := f.Name()
	assert(os.Rename(fn, fn+"."+time.Now().Format("20060102_150405")))
	go func(n string) {
		rh.Lock()
		defer rh.Unlock()
		oldLogs, _ := filepath.Glob(n + ".*")
		var backups []string
		for _, ol := range oldLogs {
			if strings.HasSuffix(ol, ".gz") {
				backups = append(backups, ol)
				continue
			}
			func(fn string) {
				defer func() {
					if e := recover(); err != nil {
						onErr(e.(error))
						return
					}
					os.Remove(fn)
					backups = append(backups, fn+".gz")
				}()
				f, err := os.Open(fn)
				assert(err)
				defer f.Close()
				g, err := os.Create(fn + ".gz")
				assert(err)
				defer func() { assert(g.Close()) }()
				zw, _ := gzip.NewWriterLevel(g, gzip.BestSpeed)
				defer func() { assert(zw.Close()) }()
				_, err = io.Copy(zw, f)
				assert(err)
			}(ol)
		}
		sort.Slice(backups, func(i, j int) bool { return backups[i] < backups[j] })
		for len(backups) >= rh.keep {
			os.Remove(backups[0])
			backups = backups[1:]
		}
	}(fn)
	return true
}

func RotatedHandler(dir string, split, keep int) (Handler, error) {
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
		split: int64(split),
		keep:  keep,
		ch:    make(chan LogItem, queueLen),
		fhm:   make(map[string]*os.File),
	}
	go rh.ingest()
	return &rh, nil
}
