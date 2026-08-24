package validate

import "ticketdesk/internal/model"

type Policy struct {
	MaxBatchSize    int
	AllowUnderscore bool
}

func DefaultPolicy() Policy { return Policy{MaxBatchSize: 10000, AllowUnderscore: true} }

func (p Policy) CheckBatch(codes []string) Issue {
	if len(codes) == 0 {
		return Issue{Code: "empty_batch", Message: "batch contains no ticket codes"}
	}
	if p.MaxBatchSize > 0 && len(codes) > p.MaxBatchSize {
		return Issue{Code: "batch_limit", Message: "batch exceeds configured limit"}
	}
	for _, code := range codes {
		if !p.AllowUnderscore && containsUnderscore(code) {
			return Issue{Code: "underscore", Message: "underscores are not allowed"}
		}
	}
	return Issue{}
}

func containsUnderscore(value string) bool {
	for _, r := range value {
		if r == '_' {
			return true
		}
	}
	return false
}

func IsTerminal(code model.TicketCode) bool { return model.TerminalCode(code.State) }
