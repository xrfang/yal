package yal

import (
	"errors"
	"fmt"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Emitter is the function prototype for generate [LogItem] from attributes
// and send to a [Handler].
type Emitter func(string, ...any)

func stringify(a any) any {
	switch v := a.(type) {
	case error:
		ss := strings.Split(trimRight(v.Error()), "\n")
		if len(ss) == 1 {
			return ss[0]
		}
		return ss
	case time.Duration:
		return v.String()
	case time.Time:
		return v.Format(time.RFC3339Nano)
	case string:
		ss := strings.Split(trimRight(v), "\n")
		if len(ss) == 1 {
			return ss[0]
		}
		return ss
	case bool:
		return strconv.FormatBool(v)
	case int:
		return strconv.FormatInt(int64(v), 10)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case uintptr:
		return fmt.Sprintf(ptrFmt, v)
	case Hex8:
		return fmt.Sprintf("%02x", v)
	case Hex16:
		return fmt.Sprintf("%04x", v)
	case Hex32:
		return fmt.Sprintf("%08x", v)
	case Hex64:
		return fmt.Sprintf("%016x", v)
	case float32:
		return strconv.FormatFloat(float64(v), 'g', -1, 64)
	case float64:
		return strconv.FormatFloat(v, 'g', -1, 64)
	case []byte:
		return v
	case complex64:
		return strconv.FormatComplex(complex128(v), 'g', -1, 128)
	case complex128:
		return strconv.FormatComplex(v, 'g', -1, 128)
	}
	return fmt.Sprintf("%v", a)
}

func parse(args ...any) map[string]any {
	attr := map[string]any{}
	for i := 0; i < len(args); i += 2 {
		if i+1 >= len(args) {
			break
		}
		switch args[i].(type) {
		case string:
			attr[args[i].(string)] = stringify(args[i+1])
		default:
			attr[badKey] = args[i]
			return attr
		}
	}
	return attr
}

func format(prop map[string]any, mesg string, args ...any) (string, map[string]any) {
	attr := parse(args...)
	ms := mtx.FindAllStringSubmatch(mesg, -1)
	for _, m := range ms {
		subst := attr[m[1]]
		switch v := subst.(type) {
		case string:
			mesg = strings.ReplaceAll(mesg, m[0], v)
		case []string:
			s := strings.Join(v, "\n")
			mesg = strings.ReplaceAll(mesg, m[0], s)
		case []byte:
			s := fmt.Sprintf("%x", v)
			mesg = strings.ReplaceAll(mesg, m[0], s)
		}
		delete(attr, m[1])
	}
	data := map[string]any{}
	for k, v := range prop {
		data[k] = v
	}
	for k, v := range attr {
		data[k] = v
	}
	return mesg, data
}

func trace(full bool) []string {
	var st []string
	n := 1
step:
	for {
		n++
		pc, file, line, ok := runtime.Caller(n)
		if !ok {
			break
		}
		f := runtime.FuncForPC(pc)
		name := f.Name()
		for _, pfx := range skip {
			if strings.HasPrefix(name, pfx) {
				continue step
			}
		}
		st = append(st, fmt.Sprintf("(%s:%d) %s", file, line, name))
		if !full {
			break
		}
	}
	return st
}

// NewLogger returns a log [Emitter] which is used to generate
// log records at "normal" log level.  The use of parameter(s)
// props are explained in the package level document.
func NewLogger(props ...any) Emitter {
	attr := parse(props...)
	return func(mesg string, args ...any) {
		Log(format(attr, mesg, args...))
	}
}

// NewDebugger returns a log [Emitter] which is used to generate
// log records at "debug" log level.  The use of parameter(s)
// props are explained in the package level document.
func NewDebugger(props ...any) Emitter {
	attr := parse(props...)
	return func(mesg string, args ...any) {
		if dbg {
			Log(format(attr, mesg, args...))
		}
	}
}

// Log makes a [LogItem] from mesg and attr, then send it to [Handler]. It
// is used internally by Emitters returned by [NewLogger] or [NewDebugger],
// although could also be used directly.
func Log(mesg string, attr map[string]any) {
	item := LogItem{time.Now(), mesg, attr}
	switch trc {
	case 1:
		item.Attr["callstack"] = trace(false)
	case 2:
		item.Attr["callstack"] = trace(true)
	}
	if flt != nil {
		flt(&item)
	}
	if peek != nil {
		item.flush(peek)
	}
	if lh != nil {
		lh.Emit(item)
	}
}

// Assert panics if the value e is an error or boolean value false.
// e can only be one of nil, error or boolean.  If e is nil or a
// boolean expression evaluates to true, Assert does nothing; if e
// is error, Assert is equivalent to
//
//	panic(e)
//
// if e is a boolean expression evaluates to false, it will panic
// with an error, like so:
//
//	yal.Assert(1 == 2, "1 must be equal to %v", 1)
//	// results in: panic(fmt.Errorf("1 must be equal to %v", 1))
//
// The parameter(s) ntfy is only used when e is a boolean expression.
func Assert(e any, ntfy ...any) {
	switch v := e.(type) {
	case nil:
	case bool:
		if !v {
			mesg := "assertion failed"
			if len(ntfy) > 0 {
				mesg = ntfy[0].(string)
				if len(ntfy) > 1 {
					mesg = fmt.Sprintf(mesg, ntfy[1:]...)
				}
			}
			panic(errors.New(mesg))
		}
	case error:
		panic(v)
	default:
		panic(fmt.Errorf("yal.Assert: expect <error> or <bool>, got <%T>", e))
	}
}

// Catch is used as a defer function which mimics try...catch logic
// in other languages such as Java. Internally, it uses recover() to
// catch panics, logs it (with callstack) and pass it on to h for
// processing. h could be one of nil, *error or func(error) error.
// If h is *error, the caught error will be stored in it and returned
// to the caller; if it is func(error) error, the caught error will be
// passed to that function and the return value from that function will
// be logged instead, if it is not nil.
//
// The use of parameter(s) args are explained in the package level document.
func Catch(h any, args ...any) {
	var e error
	switch v := recover().(type) {
	case nil:
	case string:
		e = errors.New(v)
	case error:
		e = v
	default:
		e = fmt.Errorf("%v", v)
	}
	switch proc := h.(type) {
	case nil:
	case *error:
		if e != nil {
			*proc = e
		}
	case func(error) error:
		e = proc(e)
	default:
		panic(fmt.Errorf("yal.Catch: invalid handler <%T>", h))
	}
	if e == nil {
		return
	}
	mesg, data := format(nil, e.Error(), args...)
	li := LogItem{
		When: time.Now(),
		Mesg: mesg,
		Attr: data,
	}
	st := trace(true)
	if len(st) > 0 {
		if li.Attr == nil {
			li.Attr = make(map[string]any)
		}
		li.Attr["callstack"] = st
	}
	if flt != nil {
		flt(&li)
	}
	if peek != nil {
		li.flush(peek)
	}
	if lh != nil {
		lh.Emit(li)
	}
}

var mtx *regexp.Regexp

func init() {
	mtx = regexp.MustCompile(`{{(\w+)}}`)
}
