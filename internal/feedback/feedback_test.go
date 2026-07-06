package feedback

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestAppend_WritesOneJSONLineAndCreatesDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "feedback")

	rec := Record{
		Timestamp:  time.Date(2026, 7, 4, 10, 0, 0, 0, time.UTC),
		Task:       "review foo.go",
		Transcript: "generator: draft\nreview: APPROVE",
		Approved:   true,
		Reason:     "looks good",
	}

	if err := Append(dir, rec); err != nil {
		t.Fatalf("Append: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, fileName))
	if err != nil {
		t.Fatalf("read feedback log: %v", err)
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	if !scanner.Scan() {
		t.Fatal("expected one line in feedback log, got none")
	}
	var got Record
	if err := json.Unmarshal(scanner.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal line: %v", err)
	}
	if got.Task != rec.Task || got.Approved != rec.Approved || got.Reason != rec.Reason {
		t.Errorf("got %+v, want %+v", got, rec)
	}
	if scanner.Scan() {
		t.Errorf("expected exactly one line, got a second: %s", scanner.Text())
	}
}

func TestAppend_AppendsSecondRecordAsNewLine(t *testing.T) {
	dir := t.TempDir()

	if err := Append(dir, Record{Task: "first", Approved: true}); err != nil {
		t.Fatalf("first Append: %v", err)
	}
	if err := Append(dir, Record{Task: "second", Approved: false, Reason: "needs work"}); err != nil {
		t.Fatalf("second Append: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, fileName))
	if err != nil {
		t.Fatalf("read feedback log: %v", err)
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	lines := 0
	for scanner.Scan() {
		lines++
	}
	if lines != 2 {
		t.Errorf("got %d lines, want 2", lines)
	}
}

// TestAppend_ConcurrentWritersProduceNoCorruptedLines documents and checks
// the concurrency guarantee described on Append's docstring: relying on
// O_APPEND + one f.Write per call, not an in-process lock. Every launched
// goroutine's line must appear intact and independently JSON-parseable —
// no torn/interleaved writes — though this doesn't prove atomicity beyond
// what the OS itself guarantees (it can't turn a flaky guarantee into a
// hard one, only catch a regression like splitting Append across writes).
func TestAppend_ConcurrentWritersProduceNoCorruptedLines(t *testing.T) {
	dir := t.TempDir()
	const n = 50

	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			err := Append(dir, Record{Task: fmt.Sprintf("task %d", i), Approved: i%2 == 0})
			if err != nil {
				t.Errorf("Append(%d): %v", i, err)
			}
		}(i)
	}
	wg.Wait()

	data, err := os.ReadFile(filepath.Join(dir, fileName))
	if err != nil {
		t.Fatalf("read feedback log: %v", err)
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	seen := map[string]bool{}
	lines := 0
	for scanner.Scan() {
		lines++
		var rec Record
		if err := json.Unmarshal(scanner.Bytes(), &rec); err != nil {
			t.Fatalf("line %d did not parse as JSON (torn write): %v\nline: %s", lines, err, scanner.Text())
		}
		if seen[rec.Task] {
			t.Errorf("task %q recorded more than once", rec.Task)
		}
		seen[rec.Task] = true
	}
	if lines != n {
		t.Errorf("got %d lines, want %d — a lost or duplicated write under concurrency", lines, n)
	}
}
