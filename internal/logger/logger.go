package logger

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

const maxEntries = 1000

// Level represents the log severity.
type Level string

const (
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
	LevelDebug Level = "debug"
)

// LogEntry represents a single log record.
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     Level     `json:"level"`
	Source    string    `json:"source"`
	Message   string    `json:"message"`
}

// ringBuffer is a thread-safe circular buffer of log entries.
type ringBuffer struct {
	mu      sync.Mutex
	entries []LogEntry
	pos     int
	count   int
}

var (
	instance *ringBuffer
	once     sync.Once
)

func getInstance() *ringBuffer {
	once.Do(func() {
		instance = &ringBuffer{
			entries: make([]LogEntry, maxEntries),
		}
	})
	return instance
}

func (rb *ringBuffer) add(entry LogEntry) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.entries[rb.pos] = entry
	rb.pos = (rb.pos + 1) % maxEntries
	if rb.count < maxEntries {
		rb.count++
	}
}

func (rb *ringBuffer) getAll() []LogEntry {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	if rb.count == 0 {
		return nil
	}

	result := make([]LogEntry, rb.count)
	if rb.count < maxEntries {
		// Buffer hasn't wrapped yet
		copy(result, rb.entries[:rb.count])
	} else {
		// Buffer has wrapped — read from pos (oldest) to end, then start to pos
		n := copy(result, rb.entries[rb.pos:])
		copy(result[n:], rb.entries[:rb.pos])
	}
	return result
}

func (rb *ringBuffer) clear() {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	rb.pos = 0
	rb.count = 0
}

func log(level Level, source, msg string) {
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Source:    source,
		Message:   msg,
	}

	// Write to stdout
	fmt.Printf("[%s] [%s] [%s] %s\n",
		entry.Timestamp.Format("2006-01-02 15:04:05.000"),
		entry.Level,
		entry.Source,
		entry.Message,
	)

	getInstance().add(entry)
}

// Info logs an informational message.
func Info(source, msg string) {
	log(LevelInfo, source, msg)
}

// Warn logs a warning message.
func Warn(source, msg string) {
	log(LevelWarn, source, msg)
}

// Error logs an error message.
func Error(source, msg string) {
	log(LevelError, source, msg)
}

// Debug logs a debug message.
func Debug(source, msg string) {
	log(LevelDebug, source, msg)
}

// GetAll returns all stored log entries in chronological order.
func GetAll() []LogEntry {
	return getInstance().getAll()
}

// Export returns all stored log entries as JSON bytes.
func Export() []byte {
	entries := GetAll()
	if entries == nil {
		entries = []LogEntry{}
	}
	data, _ := json.Marshal(entries)
	return data
}

// Clear flushes all stored log entries.
func Clear() {
	getInstance().clear()
}
