package main

import (
	"net/http"
	"time"

	"go.xrfang.cn/yal"
)

var log, dbg yal.Emitter
var assert yal.Checker
var catch yal.Catcher

func task(g yal.Emitter, r *http.Request) {
	g("", "method", r.Method, "url", r.URL.String())
	panic("something wrong")
}

func main() {
	L, err := yal.NewRotatedLogger(".", 1024, 0)
	if err != nil {
		panic(err)
	}
	L.Debug = true
	L.Trace = true
	L.Filter = func(li *yal.LogItem) {
		li.Mesg += "!!!"
	}
	log = L.Log()
	dbg = L.Dbg()
	assert = L.Check()
	catch = L.Catch()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		src := r.RemoteAddr
		log := L.Log("client", src, "basename", "access.log")
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
