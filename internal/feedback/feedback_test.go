package feedback

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
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
