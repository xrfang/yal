package main

import (
	"go.xrfang.cn/yal"
	"golang.org/x/exp/slog"
	"io"
	"testing"
	"time"
)

func BenchmarkSlog(b *testing.B) {
	b.StopTimer()
	attrs := []slog.Attr{
		{Key: "a", Value: slog.StringValue("1")},
		{Key: "b", Value: slog.StringValue("2")},
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(io.Discard).WithAttrs(attrs)))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		slog.Info("this is a test")
	}
	return
}

func BenchmarkYal(b *testing.B) {
	b.StopTimer()
	log := yal.NewLogger("a", "1", "b", "2")
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		log("this is a test")
	}
	return
}

func BenchmarkYalRaw(b *testing.B) {
	b.StopTimer()
	now := time.Now()
	attr := map[string]any{"a": "1", "b": "2"}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		yal.Log(yal.LogItem{
			When: now,
			Mesg: "this is a test",
			Attr: attr,
		})
	}
	return
}
