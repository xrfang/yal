package main

import (
	"io"
	"math/rand"
	"net/http"
	"os"
	"time"

	"go.xrfang.cn/yal"
)

var dbg yal.Emitter

func task1(g yal.Emitter, r *http.Request) {
	g("handling task1...", "method", r.Method, "url", r.URL.String())
	yal.Assert(1 == 2, "can you do math?")
}

func task2(g yal.Emitter, r *http.Request) {
	g("processing task2...")
	panic(io.EOF) //EOF is not an error!
}

func main() {
	yal.Debug(true)
	//yal.Trace(true)
	yal.Filter(func(li *yal.LogItem) {
		li.Mesg += "!!!"
	})
	yal.Peek(os.Stderr)
	yal.Setup(func() (yal.Handler, error) {
		return yal.RotatedHandler(".", 1024, 0)
	})
	dbg = yal.NewDebugger("basename", "main.log")
	dbg("program started, port={{port}}", "port", 1234)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		src := r.RemoteAddr
		log := yal.NewLogger("client", src, "basename", "access.log")
		defer yal.Catch(func(e error) error {
			if e == io.EOF {
				log("suppressed EOF")
				return nil
			}
			return e
		}, "client", src, "basename", "errors.log")
		if rand.Int()%2 == 0 {
			task1(log, r)
		} else {
			task2(log, r)
		}
	})
	svr := http.Server{
		Addr:         ":1234",
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
	}
	yal.Assert(svr.ListenAndServe())
}
