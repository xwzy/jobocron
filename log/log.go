package log

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/enriquebris/goconcurrentqueue"
)

const (
	defaultBufferSize = 256 * 1024  // 256 KB
	maxEntrySize      = 1024 * 1024 // 1 MB
)

type Logger struct {
	output    io.Writer
	writer    *bufio.Writer
	queue     goconcurrentqueue.Queue
	wg        sync.WaitGroup
	closed    int32
	batchSize int
	mu        sync.Mutex
}

type Entry struct {
	Time    time.Time              `json:"time"`
	Level   string                 `json:"level"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

var defaultLogger *Logger

var entryPool = sync.Pool{
	New: func() interface{} {
		return &Entry{Data: make(map[string]interface{})}
	},
}

func init() {
	defaultLogger = NewLogger(os.Stdout)
}

func NewLogger(output io.Writer) *Logger {
	if output == nil {
		output = io.Discard
	}
	writer := bufio.NewWriterSize(output, defaultBufferSize)
	logger := &Logger{
		output:    output,
		writer:    writer,
		queue:     goconcurrentqueue.NewFIFO(),
		batchSize: 100,
	}

	for i := 0; i < runtime.NumCPU(); i++ {
		go logger.processLogs()
	}

	return logger
}

func (l *Logger) processLogs() {
	batch := make([]*Entry, 0, l.batchSize)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			l.writeBatch(batch)
			batch = batch[:0]
		default:
			entry, err := l.queue.Dequeue()
			if err == nil {
				batch = append(batch, entry.(*Entry))
				if len(batch) >= l.batchSize {
					l.writeBatch(batch)
					batch = batch[:0]
				}
			} else if atomic.LoadInt32(&l.closed) == 1 && l.queue.GetLen() == 0 {
				l.writeBatch(batch)
				return
			} else {
				runtime.Gosched()
			}
		}
	}
}

func (l *Logger) writeBatch(batch []*Entry) {
	if len(batch) == 0 {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	for _, entry := range batch {
		data, err := json.Marshal(entry)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to marshal log entry: %v\n", err)
			continue
		}

		if len(data) > maxEntrySize {
			fmt.Fprintf(os.Stderr, "Log entry too large: %d bytes\n", len(data))
			continue
		}

		// If the buffer is full, flush it
		if l.writer.Available() < len(data)+1 { // +1 for newline
			if err := l.writer.Flush(); err != nil && err != io.ErrShortWrite {
				fmt.Fprintf(os.Stderr, "Failed to flush buffer: %v\n", err)
			}
		}

		if _, err := l.writer.Write(data); err != nil && err != io.ErrShortWrite {
			fmt.Fprintf(os.Stderr, "Failed to write log entry: %v\n", err)
		}
		if err := l.writer.WriteByte('\n'); err != nil && err != io.ErrShortWrite {
			fmt.Fprintf(os.Stderr, "Failed to write newline: %v\n", err)
		}

		resetEntry(entry) // Reset the entry before putting it back in the pool
		entryPool.Put(entry)
	}

	if err := l.writer.Flush(); err != nil && err != io.ErrShortWrite {
		fmt.Fprintf(os.Stderr, "Failed to flush buffer: %v\n", err)
	}

	l.wg.Add(-len(batch))
}

func (l *Logger) Close() error {
	if atomic.CompareAndSwapInt32(&l.closed, 0, 1) {
		l.wg.Wait()
		l.mu.Lock()
		defer l.mu.Unlock()
		return l.writer.Flush()
	}
	return nil
}

func (l *Logger) log(level, msg string, kvs ...interface{}) {
	e := entryPool.Get().(*Entry)
	resetEntry(e) // Reset the entry before using it
	e.Time = time.Now()
	e.Level = level
	e.Message = msg

	for i := 0; i < len(kvs); i += 2 {
		if i+1 < len(kvs) {
			key, ok := kvs[i].(string)
			if !ok {
				continue
			}
			e.Data[key] = kvs[i+1]
		}
	}

	if atomic.LoadInt32(&l.closed) == 0 {
		l.queue.Enqueue(e)
		l.wg.Add(1)
	} else {
		entryPool.Put(e)
	}
}

func resetEntry(e *Entry) {
	e.Time = time.Time{}
	e.Level = ""
	e.Message = ""
	for k := range e.Data {
		delete(e.Data, k)
	}
}

func Info(msg string, kvs ...interface{})  { defaultLogger.log("info", msg, kvs...) }
func Error(msg string, kvs ...interface{}) { defaultLogger.log("error", msg, kvs...) }
func Warn(msg string, kvs ...interface{})  { defaultLogger.log("warn", msg, kvs...) }
func Fatal(msg string, kvs ...interface{}) {
	defaultLogger.log("fatal", msg, kvs...)
	Flush()
	os.Exit(1)
}

func Flush() {
	defaultLogger.wg.Wait()
	defaultLogger.mu.Lock()
	defer defaultLogger.mu.Unlock()
	defaultLogger.writer.Flush()
}

func Close() error {
	return defaultLogger.Close()
}
