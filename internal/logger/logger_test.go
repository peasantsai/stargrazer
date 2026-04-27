package logger

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"
)

// resetLogger resets the singleton so each test starts fresh.
func resetLogger() {
	once = sync.Once{}
	instance = nil
}

func TestInfoAddsEntry(t *testing.T) {
	resetLogger()
	defer resetLogger()

	Info("test", "info message")
	entries := GetAll()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Level != LevelInfo {
		t.Errorf("expected level info, got %s", entries[0].Level)
	}
	if entries[0].Source != "test" {
		t.Errorf("expected source 'test', got %q", entries[0].Source)
	}
	if entries[0].Message != "info message" {
		t.Errorf("expected message 'info message', got %q", entries[0].Message)
	}
}

func TestWarnAddsEntry(t *testing.T) {
	resetLogger()
	defer resetLogger()

	Warn("src", "warn msg")
	entries := GetAll()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Level != LevelWarn {
		t.Errorf("expected level warn, got %s", entries[0].Level)
	}
}

func TestErrorAddsEntry(t *testing.T) {
	resetLogger()
	defer resetLogger()

	Error("src", "error msg")
	entries := GetAll()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Level != LevelError {
		t.Errorf("expected level error, got %s", entries[0].Level)
	}
}

func TestDebugAddsEntry(t *testing.T) {
	resetLogger()
	defer resetLogger()

	Debug("src", "debug msg")
	entries := GetAll()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Level != LevelDebug {
		t.Errorf("expected level debug, got %s", entries[0].Level)
	}
}

func TestGetAllReturnsInOrder(t *testing.T) {
	resetLogger()
	defer resetLogger()

	Info("src", "first")
	Warn("src", "second")
	Error("src", "third")

	entries := GetAll()
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	if entries[0].Message != "first" {
		t.Errorf("expected first entry 'first', got %q", entries[0].Message)
	}
	if entries[1].Message != "second" {
		t.Errorf("expected second entry 'second', got %q", entries[1].Message)
	}
	if entries[2].Message != "third" {
		t.Errorf("expected third entry 'third', got %q", entries[2].Message)
	}
}

func TestRingBufferWrapsAfter1000(t *testing.T) {
	resetLogger()
	defer resetLogger()

	// Fill beyond capacity.
	for i := 0; i < 1050; i++ {
		Info("src", "msg")
	}

	entries := GetAll()
	if len(entries) != 1000 {
		t.Errorf("expected 1000 entries after wrap, got %d", len(entries))
	}
}

func TestRingBufferPreservesOrderAfterWrap(t *testing.T) {
	resetLogger()
	defer resetLogger()

	for i := 0; i < 1005; i++ {
		Info("src", "msg")
	}

	entries := GetAll()
	if len(entries) != 1000 {
		t.Fatalf("expected 1000 entries, got %d", len(entries))
	}

	// All entries should have valid timestamps in non-decreasing order.
	for i := 1; i < len(entries); i++ {
		if entries[i].Timestamp.Before(entries[i-1].Timestamp) {
			t.Errorf("entry %d timestamp before entry %d", i, i-1)
		}
	}
}

func TestExportReturnsValidJSON(t *testing.T) {
	resetLogger()
	defer resetLogger()

	Info("src", "hello")
	Warn("src", "world")

	data := Export()
	var entries []LogEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		t.Fatalf("Export returned invalid JSON: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries in export, got %d", len(entries))
	}
}

func TestExportEmptyReturnsEmptyArray(t *testing.T) {
	resetLogger()
	defer resetLogger()

	data := Export()
	if string(data) != "[]" {
		t.Errorf("expected '[]' for empty export, got %q", string(data))
	}
}

func TestClearEmptiesBuffer(t *testing.T) {
	resetLogger()
	defer resetLogger()

	Info("src", "a")
	Info("src", "b")
	Info("src", "c")

	Clear()

	entries := GetAll()
	if entries != nil {
		t.Errorf("expected nil after Clear, got %d entries", len(entries))
	}
}

func TestConcurrentWritesAreSafe(t *testing.T) {
	resetLogger()
	defer resetLogger()

	var wg sync.WaitGroup
	goroutines := 50
	messagesPerGoroutine := 100

	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func(id int) {
			defer wg.Done()
			for i := 0; i < messagesPerGoroutine; i++ {
				switch i % 4 {
				case 0:
					Info("concurrent", "info")
				case 1:
					Warn("concurrent", "warn")
				case 2:
					Error("concurrent", "error")
				case 3:
					Debug("concurrent", "debug")
				}
			}
		}(g)
	}
	wg.Wait()

	entries := GetAll()
	total := goroutines * messagesPerGoroutine
	if total > 1000 {
		total = 1000
	}
	if len(entries) != total {
		t.Errorf("expected %d entries, got %d", total, len(entries))
	}
}

func TestGetAllReturnsNilWhenEmpty(t *testing.T) {
	resetLogger()
	defer resetLogger()

	entries := GetAll()
	if entries != nil {
		t.Errorf("expected nil for empty buffer, got %d entries", len(entries))
	}
}

func TestLevelConstants(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{LevelInfo, "info"},
		{LevelWarn, "warn"},
		{LevelError, "error"},
		{LevelDebug, "debug"},
	}
	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			if string(tc.level) != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, string(tc.level))
			}
		})
	}
}

func TestLogEntryHasTimestamp(t *testing.T) {
	resetLogger()
	defer resetLogger()

	before := time.Now()
	Info("src", "timestamp test")
	after := time.Now()

	entries := GetAll()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	ts := entries[0].Timestamp
	if ts.Before(before) || ts.After(after) {
		t.Errorf("timestamp %v not in expected range [%v, %v]", ts, before, after)
	}
}

func TestExportContainsAllFields(t *testing.T) {
	resetLogger()
	defer resetLogger()

	Info("testsource", "testmessage")

	data := Export()
	s := string(data)

	for _, want := range []string{"testsource", "testmessage", "info", "timestamp"} {
		if !strings.Contains(s, want) {
			t.Errorf("export JSON missing field/value %q; got: %s", want, s)
		}
	}
}

func TestMaxEntriesConstant(t *testing.T) {
	if maxEntries != 1000 {
		t.Errorf("expected maxEntries 1000, got %d", maxEntries)
	}
}

func TestClearThenAddWorks(t *testing.T) {
	resetLogger()
	defer resetLogger()

	Info("src", "before clear")
	Clear()
	Info("src", "after clear")

	entries := GetAll()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry after clear+add, got %d", len(entries))
	}
	if entries[0].Message != "after clear" {
		t.Errorf("expected 'after clear', got %q", entries[0].Message)
	}
}
