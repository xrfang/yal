package yal

import (
	"fmt"
	"os"
	"runtime"
	"strings"
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
