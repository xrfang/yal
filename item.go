package yal

import (
	"fmt"
	"io"
	"runtime"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type item struct {
	when  time.Time
	mesg  string
	data  map[string]any
	debug bool //是否为debug信息
	/*
		- 20230406_114456.123: something....
		  callstack:
		  - line1
		  - line2
		  arg1: value1
		  arg2: value2
	*/
}

func (li *item) Trace() {
	var st []string
	n := 1
	for {
		n++
		pc, file, line, ok := runtime.Caller(n)
		if !ok {
			break
		}
		f := runtime.FuncForPC(pc)
		name := f.Name()
		if strings.HasPrefix(name, "runtime.") {
			continue
		}
		fn := strings.Split(file, "/")
		if len(fn) > 1 {
			file = strings.Join(fn[len(fn)-2:], "/")
		}
		st = append(st, fmt.Sprintf("(%s:%d) %s", file, line, name))
	}
	if len(st) > 0 {
		if li.data == nil {
			li.data = make(map[string]any)
		}
		li.data["callstack"] = st
	}
}

func (li item) Flush(w io.Writer) error {
	data := li.data
	if data == nil {
		data = make(map[string]any)
	}
	data[li.when.Format("20060102_150405.000")] = li.mesg
	return yaml.NewEncoder(w).Encode([]map[string]any{data})
}
