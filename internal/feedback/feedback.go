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
