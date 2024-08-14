package log

import (
	"io/ioutil"
	"testing"
)

func BenchmarkLoggerInfo(b *testing.B) {
	logger := NewLogger(ioutil.Discard)
	defer logger.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.log("info", "benchmark test message")
	}
}

func BenchmarkLoggerInfoWithFields(b *testing.B) {
	logger := NewLogger(ioutil.Discard)
	defer logger.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.log("info", "benchmark test message", "key1", "value1", "key2", 123)
	}
}

func BenchmarkDefaultLoggerInfo(b *testing.B) {
	defaultLogger = NewLogger(ioutil.Discard)
	defer Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Info("benchmark test message")
	}
}

func BenchmarkConcurrentLogging(b *testing.B) {
	logger := NewLogger(ioutil.Discard)
	defer logger.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.log("info", "benchmark concurrent test message")
		}
	})
}

func BenchmarkEntryPooling(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e := entryPool.Get().(*Entry)
		e.Level = "info"
		e.Message = "benchmark test message"
		e.Data["key"] = "value"
		resetEntry(e)
		entryPool.Put(e)
	}
}
