package validate

import "ticketdesk/internal/model"

type Outcome struct {
	State   string
	Message string
	Issue   Issue
}

func DecideCodeOutcome(code model.TicketCode, issue Issue, consumeErr error) Outcome {
	if issue.Code != "" {
		return Outcome{State: model.CodeInvalid, Message: issue.Message, Issue: issue}
	}
	if consumeErr == nil {
		return Outcome{State: model.CodeConsumed, Message: "ticket accepted"}
	}
	if consumeErr.Error() == "ticket already consumed" {
		return Outcome{State: model.CodeDuplicate, Message: "ticket was already consumed"}
	}
	return Outcome{State: model.CodeInvalid, Message: consumeErr.Error()}
}

func BatchState(processed, succeeded, failed, total int) string {
	if total == 0 {
		return model.BatchPending
	}
	if processed < total {
		return model.BatchRunning
	}
	if failed > 0 {
		return model.BatchFailed
	}
	return model.BatchComplete
}

func IsRetryAllowed(detail model.FailureDetail, retryCount int) bool {
	return detail.Retryable && !detail.Resolved && retryCount < 3
}
