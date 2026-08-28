package model

const (
	BatchPending     = "pending"
	BatchRunning     = "running"
	BatchComplete    = "complete"
	BatchFailed      = "failed"
	CodePending      = "pending"
	CodeConsumed     = "consumed"
	CodeDuplicate    = "duplicate"
	CodeInvalid      = "invalid"
	AttemptSuccess   = "success"
	AttemptFailure   = "failure"
	AttemptDuplicate = "duplicate"
	TaskQueued       = "queued"
	TaskWorking      = "working"
	TaskDone         = "done"
)

func ValidBatchStatus(status string) bool {
	switch status {
	case BatchPending, BatchRunning, BatchComplete, BatchFailed:
		return true
	default:
		return false
	}
}

func ValidCodeState(state string) bool {
	switch state {
	case CodePending, CodeConsumed, CodeDuplicate, CodeInvalid:
		return true
	default:
		return false
	}
}

func TerminalCode(state string) bool {
	return state == CodeConsumed || state == CodeDuplicate || state == CodeInvalid
}
