package yal

import (
	"fmt"
	"io"
	"runtime"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type LogItem struct {
	When  time.Time
	Mesg  string
	Data  map[string]any
	Level byte //0=普通消息，1=DEBUG消息
}

func (li *LogItem) Trace() {
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
		if li.Data == nil {
			li.Data = make(map[string]any)
		}
		li.Data["callstack"] = st
	}
}

func (li LogItem) Flush(w io.Writer) error {
	data := li.Data
	if data == nil {
		data = make(map[string]any)
	}
	data[li.When.Format("20060102_150405.000")] = li.Mesg
	return yaml.NewEncoder(w).Encode([]map[string]any{data})
}
