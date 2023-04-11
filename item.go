package yal

import (
	"encoding/hex"
	"fmt"
	"io"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

type LogItem struct {
	When time.Time
	Mesg string
	Attr map[string]any
}

var (
	yml = []byte{'|', '\n', ' ', ' ', ' ', ' ', '-', ' ', ':', ' '}
	stk = []byte("  callstack:\n")
)

func trace(full bool) []string {
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
		if !full {
			break
		}
	}
	return st
}

func trimRight(str string) string {
	for i := len(str); i > 0; i-- {
		if str[i-1] > 32 {
			return str[:i]
		}
	}
	return ""
}

func (li LogItem) Flush(w io.Writer) (err error) {
	write := func(buf []byte) {
		_, err = w.Write(buf)
		if err != nil {
			panic(err)
		}
	}
	defer func() { recover() }()
	write(yml[6:8]) //'-', ' '
	write([]byte(li.When.Format("20060102_150405.000")))
	write(yml[8:]) //':', ' '
	write([]byte(li.Mesg))
	write(yml[1:2]) //'\n'
	var keys []string
	var call []string
	for k, v := range li.Attr {
		if k == "callstack" {
			call = v.([]string)
		} else {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	for _, k := range keys {
		write(yml[2:4]) //' ', ' '
		write([]byte(k))
		write(yml[8:]) //':', ' '
		var ss []string
		switch v := li.Attr[k].(type) {
		case string:
			ss = strings.Split(trimRight(v), "\n")
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
		case []byte:
			ss = strings.Split(trimRight(hex.Dump(v)), "\n")
		case complex64:
			ss = []string{strconv.FormatComplex(complex128(v), 'g', -1, 128)}
		case complex128:
			ss = []string{strconv.FormatComplex(v, 'g', -1, 128)}
		default:
			ss = []string{badVal}
		}
		switch len(ss) {
		case 1:
			s := trimRight(ss[0])
			write([]byte(s))
			fallthrough
		case 0:
			write(yml[1:2]) //'\n'
		default:
			write(yml[:2]) //'|', '\n'
			for _, s := range ss {
				s = trimRight(s)
				write(yml[2:6]) //' ', ' ', ' ', ' '
				write([]byte(s))
				write(yml[1:2]) //'\n'
			}
		}
	}
	if len(call) > 0 {
		write(stk) //"  callstack:\n"
		for _, c := range call {
			write(yml[4:8]) //' ', ' ', '-', ' '
			write([]byte(c))
			write(yml[1:2]) //'\n'
		}
	}
	return
}
