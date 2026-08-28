package model

import "strings"

func NormalizeSource(source string) string { return strings.TrimSpace(strings.ToLower(source)) }

func NormalizeBatch(batch TicketBatch) TicketBatch {
	batch.ID = strings.TrimSpace(batch.ID)
	batch.Source = NormalizeSource(batch.Source)
	batch.CreatedBy = CanonicalWorker(batch.CreatedBy)
	return batch
}

func (b TicketBatch) OutcomeCounts() (int, int, int) { return b.Processed, b.Succeeded, b.Failed }

func (c TicketCode) DisplayLabel() string {
	if c.Holder == "" {
		return c.Code
	}
	return c.Code + " (" + c.Holder + ")"
}

func (f FailureDetail) DisplayLabel() string {
	if f.Category == "" {
		return f.Code + ": " + f.Message
	}
	return f.Code + " [" + f.Category + "]: " + f.Message
}

func (a ValidationAttempt) Succeeded() bool { return a.Outcome == AttemptSuccess }

func (t WorkerTask) Ready() bool { return t.State == TaskQueued && t.Code != "" && t.BatchID != "" }
