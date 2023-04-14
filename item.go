package yal

import (
	"encoding/hex"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
	"unsafe"
)

type (
	Hex8    uint8
	Hex16   uint16
	Hex32   uint32
	Hex64   uint64
	LogItem struct {
		When time.Time
		Mesg string
		Attr map[string]any
	}
)

var (
	yml = []byte{'|', '\n', ' ', ' ', ' ', ' ', '-', ' ', ':', ' '}
	stk = []byte("  callstack:\n")
)

func trimRight(str string) string {
	for i := len(str); i > 0; i-- {
		if str[i-1] > 32 {
			return str[:i]
		}
	}
	return ""
}

func (li *LogItem) flush(w io.Writer) (err error) {
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
	msg := strings.Split(trimRight(li.Mesg), "\n")
	switch len(msg) {
	case 1:
		write([]byte(msg[0]))
		fallthrough
	case 0:
		write(yml[1:2]) //'\n'
	default:
		write(yml[:2]) //'|', '\n'
		for _, s := range msg {
			write(yml[2:4]) //' ', ' ',
			write([]byte(s))
			write(yml[1:2]) //'\n'
		}
	}
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
		switch v := li.Attr[k].(type) {
		case string:
			write([]byte(v))
			write(yml[1:2]) //'\n'
		case []string:
			write(yml[:2]) //'|', '\n'
			for _, s := range v {
				write(yml[2:6]) //' ', ' ', ' ', ' '
				write([]byte(s))
				write(yml[1:2]) //'\n'
			}
		case []byte:
			write(yml[:2]) //'|', '\n'
			for _, s := range strings.Split(trimRight(hex.Dump(v)), "\n") {
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

func (li *LogItem) Trace() {
	li.Attr["callstack"] = trace(true)
}

var ptrFmt string

func init() {
	ptrFmt = fmt.Sprintf("%%0%dx", unsafe.Sizeof(uintptr(0))*2)
}
