package yal

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type (
	rotatedHandler struct {
		dir   string
		split int64
		keep  int
		hkch  chan string //housekeeping
		fhm   map[string]*os.File
		sync.Mutex
	}
)

func assert(err error) {
	if err != nil {
		panic(err)
	}
}

func onErr(err error) bool {
	if err == nil {
		return false
	}
	fmt.Fprintf(os.Stderr, "ERROR: yal: %v\n", err)
	for _, s := range trace(true) {
		fmt.Fprintln(os.Stderr, "  ", s)
	}
	return true
}

func (rh *rotatedHandler) Emit(li LogItem) {
	rh.Lock()
	defer rh.Unlock()
	base := "log"
	switch bn := li.Attr["basename"].(type) {
	case string:
		if bn = strings.TrimSpace(bn); bn != "" {
			base = bn
		}
	}
	f := rh.fhm[base]
	if f == nil {
		fp := filepath.Join(rh.dir, base)
		if onErr(os.MkdirAll(filepath.Dir(fp), 0777)) {
			return
		}
		var err error
		f, err = os.OpenFile(fp, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if onErr(err) {
			return
		}
		rh.fhm[base] = f
	}
	delete(li.Attr, "basename")
	li.flush(f)
	if rh.rotate(f) {
		delete(rh.fhm, base)
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
	go func() { rh.hkch <- fn }()
	return true
}

// RotatedHandler saves log message to files under dir, and rotate
// out the current log file as it reaches split bytes.  If number
// of old log files exceeds keep, the oldest one is removed.
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
		hkch:  make(chan string),
		fhm:   make(map[string]*os.File),
	}
	go func() {
		for {
			n := <-rh.hkch
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
		}
	}()
	return &rh, nil
}
