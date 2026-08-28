package model

import (
	"encoding/json"
	"fmt"
	"strings"
)

type TicketBatch struct {
	ID          string `json:"id"`
	Source      string `json:"source"`
	CreatedBy   string `json:"created_by"`
	Status      string `json:"status"`
	Total       int    `json:"total"`
	Processed   int    `json:"processed"`
	Succeeded   int    `json:"succeeded"`
	Failed      int    `json:"failed"`
	RetryCount  int    `json:"retry_count"`
	CreatedAt   int64  `json:"created_at"`
	CompletedAt int64  `json:"completed_at"`
}

type TicketCode struct {
	BatchID   string `json:"batch_id"`
	Code      string `json:"code"`
	Holder    string `json:"holder"`
	State     string `json:"state"`
	Reason    string `json:"reason"`
	Attempts  int    `json:"attempts"`
	Validated bool   `json:"validated"`
}

type ValidationAttempt struct {
	ID       string `json:"id"`
	BatchID  string `json:"batch_id"`
	Code     string `json:"code"`
	Worker   string `json:"worker"`
	Outcome  string `json:"outcome"`
	Message  string `json:"message"`
	Sequence int    `json:"sequence"`
}

type FailureDetail struct {
	ID        string `json:"id"`
	BatchID   string `json:"batch_id"`
	Code      string `json:"code"`
	Category  string `json:"category"`
	Message   string `json:"message"`
	Retryable bool   `json:"retryable"`
	Resolved  bool   `json:"resolved"`
}

type WorkerTask struct {
	ID      string `json:"id"`
	BatchID string `json:"batch_id"`
	Code    string `json:"code"`
	Worker  string `json:"worker"`
	State   string `json:"state"`
	Tries   int    `json:"tries"`
}

type BatchSummary struct {
	Batch    TicketBatch         `json:"batch"`
	Codes    []TicketCode        `json:"codes"`
	Failures []FailureDetail     `json:"failures"`
	Attempts []ValidationAttempt `json:"attempts"`
	Pending  int                 `json:"pending"`
	Percent  int                 `json:"percent"`
}

func (b TicketBatch) IsComplete() bool {
	return b.Total > 0 && b.Processed >= b.Total && (b.Status == "complete" || b.Status == "failed")
}

func (b TicketBatch) Progress() int {
	if b.Total <= 0 {
		return 0
	}
	p := b.Processed * 100 / b.Total
	if p > 100 {
		return 100
	}
	return p
}

func (c TicketCode) Key() string { return c.BatchID + ":" + c.Code }

func (c TicketCode) IsUsable() bool {
	return strings.TrimSpace(c.Code) != "" && c.State == "pending"
}

func (f FailureDetail) Key() string { return f.BatchID + ":" + f.Code + ":" + f.ID }

func Encode[T any](value T) ([]byte, error) { return json.Marshal(value) }

func Decode[T any](data []byte, target *T) error {
	if len(data) == 0 {
		return fmt.Errorf("empty record")
	}
	return json.Unmarshal(data, target)
}

func NormalizeCode(code string) string { return strings.ToUpper(strings.TrimSpace(code)) }
