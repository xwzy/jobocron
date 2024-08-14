package log

import (
	"bufio"
	"bytes"
	_ "io"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
	_ "time"
)

func TestConcurrentLogging(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			logger.log("info", "concurrent test", "goroutine", i)
		}(i)
	}
	wg.Wait()
	logger.Close()

	scanner := bufio.NewScanner(&buf)
	count := 0
	for scanner.Scan() {
		count++
	}
	if count != 100 {
		t.Errorf("Expected 100 log entries, got %d", count)
	}
}

func TestLoggerClose(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf)

	logger.log("info", "test message")
	if err := logger.Close(); err != nil {
		t.Errorf("Unexpected error on close: %v", err)
	}

	// Try to log after closing
	logger.log("info", "should not be logged")

	scanner := bufio.NewScanner(&buf)
	count := 0
	for scanner.Scan() {
		count++
	}
	if count != 1 {
		t.Errorf("Expected 1 log entry after close, got %d", count)
	}
}

func TestLoggerFlush(t *testing.T) {
	var buf bytes.Buffer
	defaultLogger = NewLogger(&buf)

	Info("test flush")
	Flush()

	scanner := bufio.NewScanner(&buf)
	if !scanner.Scan() {
		t.Error("Expected log entry after flush, got none")
	}
}

func TestLoggerWithCustomWriter(t *testing.T) {
	customWriter := &customTestWriter{
		content: new(bytes.Buffer),
	}
	logger := NewLogger(customWriter)

	logger.log("info", "test custom writer")
	logger.Close()

	if !strings.Contains(customWriter.content.String(), "test custom writer") {
		t.Error("Expected log message in custom writer, not found")
	}
}

func TestEntryPool(t *testing.T) {
	e1 := entryPool.Get().(*Entry)
	e1.Level = "info"
	e1.Message = "test message"
	e1.Data["key"] = "value"

	resetEntry(e1) // 使用新的 resetEntry 函数
	entryPool.Put(e1)

	e2 := entryPool.Get().(*Entry)
	if e2.Level != "" || e2.Message != "" || len(e2.Data) != 0 {
		t.Error("Entry from pool should be reset")
	}

	// 额外的检查
	if !e2.Time.IsZero() {
		t.Error("Entry time should be zero")
	}
}

func TestLoggerWithNilOutput(t *testing.T) {
	logger := NewLogger(nil)

	// This should not panic
	logger.log("info", "test nil output")

	// Wait for log processing
	time.Sleep(200 * time.Millisecond)

	// Close the logger
	if err := logger.Close(); err != nil {
		t.Errorf("Unexpected error on close: %v", err)
	}

	// Verify that the logger's output is set to ioutil.Discard
	if logger.output != ioutil.Discard {
		t.Error("Expected logger output to be set to ioutil.Discard")
	}
}

type customTestWriter struct {
	content *bytes.Buffer
}

func (w *customTestWriter) Write(p []byte) (n int, err error) {
	return w.content.Write(p)
}

// Mock os.Exit for testing
var osExit = os.Exit
