// Package feedback records human approve/reject decisions on agent output
// as an append-only JSONL annotation log — the RLHF-style signal the offline
// eval pipeline (Genkit, per the design spec) trains or judges against
// later. It is not a live training loop; it's the durable record a training
// or eval step reads.
package feedback

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Record is one human decision on one agent run.
type Record struct {
	Timestamp  time.Time `json:"timestamp"`
	Task       string    `json:"task"`
	Transcript string    `json:"transcript"`
	Approved   bool      `json:"approved"`
	Reason     string    `json:"reason"`
}

const fileName = "feedback.jsonl"

// Append writes rec as one JSON line to dir/feedback.jsonl, creating dir if
// needed. Append-only by design: annotations are a historical log, never
// edited in place.
//
// Concurrent-writer safety relies entirely on the OS: the file is opened
// with O_APPEND and the line is written in a single f.Write call, so each
// call's syscall(s) either fully land or don't race with another process's
// — true on POSIX-compliant local filesystems as long as the line stays
// under PIPE_BUF (historically 4096 bytes; a very long transcript could
// exceed it and risk interleaving). Not guaranteed on network filesystems
// (e.g. NFS) that don't implement atomic O_APPEND. This process itself has
// no in-process mutex, so two goroutines calling Append concurrently rely
// on the same OS guarantee, not additional locking here.
func Append(dir string, rec Record) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create feedback dir: %w", err)
	}

	f, err := os.OpenFile(filepath.Join(dir, fileName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open feedback log: %w", err)
	}
	defer f.Close()

	line, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("marshal feedback record: %w", err)
	}
	if _, err := f.Write(append(line, '\n')); err != nil {
		return fmt.Errorf("write feedback record: %w", err)
	}
	return nil
}
