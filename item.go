package yal

import (
	"fmt"
	"io"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
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

func (li LogItem) Flush(w io.Writer) (err error) {
	write := func(buf []byte) {
		_, err = w.Write(buf)
		if err != nil {
			panic(err)
		}
	}
	defer func() { recover() }()
	write([]byte(li.When.Format("20060102_150405.000")))
	write([]byte{':', ' '})
	write([]byte(li.Mesg))
	write([]byte{'\n'})
	var keys []string
	var call []string
	for k, v := range li.Data {
		if k == "callstack" {
			call = v.([]string)
		} else {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	for _, k := range keys {
		write([]byte{' ', ' '})
		write([]byte(k))
		write([]byte{':'})
		var ss []string
		switch v := li.Data[k].(type) {
		case string:
			ss = strings.Split(strings.TrimRight(v, " \n\r\t\v"), "\n")
		case bool:
			ss = []string{strconv.FormatBool(v)}
		case int:
			ss = []string{strconv.FormatInt(int64(v), 10)}
		case int8:
			ss = []string{strconv.FormatInt(int64(v), 10)}
		case int16:
			ss = []string{strconv.FormatInt(int64(v), 10)}
		case int32:
			ss = []string{strconv.FormatInt(int64(v), 10)}
		case int64:
			ss = []string{strconv.FormatInt(v, 10)}
		case uint:
			ss = []string{strconv.FormatUint(uint64(v), 10)}
		case uint8:
			ss = []string{strconv.FormatUint(uint64(v), 10)}
		case uint16:
			ss = []string{strconv.FormatUint(uint64(v), 10)}
		case uint32:
			ss = []string{strconv.FormatUint(uint64(v), 10)}
		case uint64:
			ss = []string{strconv.FormatUint(v, 10)}
		case uintptr:
			ss = []string{strconv.FormatUint(uint64(v), 16)}
		case float32:
			ss = []string{strconv.FormatFloat(float64(v), 'g', -1, 64)}
		case float64:
			ss = []string{strconv.FormatFloat(v, 'g', -1, 64)}
		case complex64:
			ss = []string{strconv.FormatComplex(complex128(v), 'g', -1, 128)}
		case complex128:
			ss = []string{strconv.FormatComplex(v, 'g', -1, 128)}
		default:
			ss = []string{"!BADVALUE"}
		}
		switch len(ss) {
		case 1:
			write([]byte{' '})
			s := strings.TrimRight(ss[0], " \n\r\t\v")
			write([]byte(s))
			fallthrough
		case 0:
			write([]byte{'\n'})
		default:
			write([]byte{' ', '|', '\n'})
			for _, s := range ss {
				s = strings.TrimRight(s, " \n\r\t\v")
				write([]byte{' ', ' ', ' ', ' '}) //extra indent 2 spaces
				write([]byte(s))
				write([]byte{'\n'})
			}
		}
	}
	if len(call) > 0 {
		write([]byte("  callstack:\n"))
		for _, c := range call {
			write([]byte{' ', ' ', '-', ' '})
			write([]byte(c))
			write([]byte{'\n'})
		}
	}
	return
}
