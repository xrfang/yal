package yal

import (
	"io"
	"reflect"
)

const badKey = "!BADKEY"

type (
	// Handler sends log items to "backends".
	Handler interface {
		Emit(LogItem)
	}
)

// Debug controls "log level", when set to false (which is default ), only
// items emitted with "normal logger" (see [NewLogger]) will be sent to handler,
// otherwise, items emitted with "debug logger" (see [NewDebugger]) will also
// be sent.
func Debug(on bool) {
	dbg = on
}

// Trace controls stack tracing of a log item.  If Trace is 0, no stack
// info will be added to a normal log item; if Trace is 1, a callstack
// item indicating the source code position of the log statement will be
// added; otherwise, full callstack will be added.
//
// NOTE: this option does NOT affect stack tracing of panics. In case
// of "exception", the log item will always contain full callstack, see
// [Catch] for more info.
func Trace(mode byte) {
	if mode < 2 {
		trc = mode
	} else {
		trc = 2
	}
}

// RemoveFromCallStack removes certain statements from callstack.
// By default, stack tracing will ignore function calls within the
// Go runtime and the "yal" package itself.
func RemoveFromCallStack(patterns ...string) {
	skip = append([]string{"runtime.", pkgp}, patterns...)
}

// Peek is used to setup a "multi-writer", so that all log items sent
// to the [Handler] are also copied to that writer.  Typical use is:
//
//	yal.Peek(os.Stdout)
func Peek(w io.Writer) {
	peek = w
}

// Filter can be used to modify attributes of a log item before it is output
// to a backend, for example, to sanitize personal information, or add/remove
// attributes.
func Filter(f func(*LogItem)) {
	flt = f
}

// Setup add handler to the logging system.  The function f
// inintializes and return a [Handler] along with an error,
// if any.  If f returns an error, it will be returned by
// Setup directly.
func Setup(f func() (Handler, error)) error {
	h, err := f()
	if err != nil {
		return err
	}
	lh = h
	return nil
}

var (
	dbg  bool
	trc  byte
	flt  func(*LogItem)
	peek io.Writer
	lh   Handler
	pkgp string
	skip []string
)

func init() {
	pkgp = reflect.TypeOf(LogItem{}).PkgPath() + "."
	skip = []string{"runtime.", pkgp}
}
