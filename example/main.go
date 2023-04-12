package main

import (
	"net/http"
	"os"
	"time"

	"go.xrfang.cn/yal"
)

var log, dbg yal.Emitter
var assert yal.Checker
var catch yal.Catcher

func task(g yal.Emitter, r *http.Request) {
	g("handling task...", "method", r.Method, "url", r.URL.String())
	assert(1 == 2, "can you to math?")
}

func main() {
	yal.Debug(true)
	yal.Trace(true)
	yal.Filter(func(li *yal.LogItem) {
		li.Mesg += "!!!"
	})
	yal.Peek(os.Stderr)
	yal.Setup(func() (yal.Handler, error) {
		return yal.RotatedHandler(".", 1024, 0)
	})
	log = yal.NewLogger()
	dbg = yal.NewDebugger()
	assert = yal.ErrChecker()
	catch = yal.NewCatcher()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		src := r.RemoteAddr
		log := yal.NewLogger("client", src, "basename", "access.log")
		defer catch(nil, "client", src, "basename", "errors.log")
		task(log, r)
	})
	svr := http.Server{
		Addr:         ":1234",
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
	}
	assert(svr.ListenAndServe())
}
