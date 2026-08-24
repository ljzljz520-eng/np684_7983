package model

import "fmt"

type LifecycleEvent struct {
	ID       string `json:"id"`
	BatchID  string `json:"batch_id"`
	Code     string `json:"code"`
	From     string `json:"from"`
	To       string `json:"to"`
	Actor    string `json:"actor"`
	Sequence int    `json:"sequence"`
}

func (b TicketBatch) Validate() error {
	if !ValidIdentifier(b.ID) {
		return fmt.Errorf("invalid batch id")
	}
	if b.Total < 0 || b.Processed < 0 {
		return fmt.Errorf("batch counters cannot be negative")
	}
	if b.Processed > b.Total {
		return fmt.Errorf("processed exceeds total")
	}
	if b.Succeeded < 0 || b.Failed < 0 {
		return fmt.Errorf("outcome counters cannot be negative")
	}
	if b.Succeeded+b.Failed > b.Processed {
		return fmt.Errorf("outcomes exceed processed")
	}
	if !ValidBatchStatus(b.Status) {
		return fmt.Errorf("invalid batch status")
	}
	return nil
}

func (c TicketCode) Validate() error {
	if !ValidIdentifier(c.BatchID) {
		return fmt.Errorf("invalid code batch")
	}
	if NormalizeCode(c.Code) != c.Code {
		return fmt.Errorf("code is not canonical")
	}
	if !ValidCodeState(c.State) {
		return fmt.Errorf("invalid code state")
	}
	if c.Attempts < 0 {
		return fmt.Errorf("attempts cannot be negative")
	}
	return nil
}

func (a ValidationAttempt) Validate() error {
	if a.ID == "" || a.BatchID == "" || a.Code == "" {
		return fmt.Errorf("attempt identity is incomplete")
	}
	if a.Outcome != AttemptSuccess && a.Outcome != AttemptFailure && a.Outcome != AttemptDuplicate {
		return fmt.Errorf("invalid attempt outcome")
	}
	if a.Sequence < 0 {
		return fmt.Errorf("attempt sequence cannot be negative")
	}
	return nil
}

func (f FailureDetail) Validate() error {
	if f.ID == "" || f.BatchID == "" || f.Code == "" {
		return fmt.Errorf("failure identity is incomplete")
	}
	if f.Category == "" || f.Message == "" {
		return fmt.Errorf("failure explanation is incomplete")
	}
	return nil
}

func (t WorkerTask) Validate() error {
	if t.ID == "" || t.BatchID == "" || t.Code == "" {
		return fmt.Errorf("task identity is incomplete")
	}
	if t.State != TaskQueued && t.State != TaskWorking && t.State != TaskDone {
		return fmt.Errorf("invalid task state")
	}
	if t.Tries < 0 {
		return fmt.Errorf("task tries cannot be negative")
	}
	return nil
}

func TransitionCode(code TicketCode, target, actor string) (TicketCode, LifecycleEvent, error) {
	if err := code.Validate(); err != nil {
		return code, LifecycleEvent{}, err
	}
	if TerminalCode(code.State) {
		return code, LifecycleEvent{}, fmt.Errorf("terminal code cannot transition")
	}
	if target != CodeConsumed && target != CodeDuplicate && target != CodeInvalid {
		return code, LifecycleEvent{}, fmt.Errorf("unsupported target state")
	}
	event := LifecycleEvent{ID: code.Key() + ":" + target, BatchID: code.BatchID, Code: code.Code, From: code.State, To: target, Actor: CanonicalWorker(actor), Sequence: code.Attempts + 1}
	code.State = target
	code.Attempts++
	code.Holder = CanonicalWorker(actor)
	code.Validated = target == CodeConsumed
	return code, event, nil
}

func CompletionMessage(batch TicketBatch) string {
	if batch.Total == 0 {
		return "no tickets submitted"
	}
	if !batch.IsComplete() {
		return fmt.Sprintf("%d of %d tickets processed", batch.Processed, batch.Total)
	}
	if batch.Failed > 0 {
		return fmt.Sprintf("completed with %d failures", batch.Failed)
	}
	return "all tickets accepted"
}
