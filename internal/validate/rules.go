package validate

import (
	"fmt"
	"strings"
	"unicode"

	"ticketdesk/internal/model"
)

type Issue struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Retryable bool   `json:"retryable"`
}

type CodeValidator struct{ forbiddenPrefixes []string }

func NewCodeValidator() *CodeValidator {
	return &CodeValidator{forbiddenPrefixes: []string{"VOID-", "TEST-", "FAKE-"}}
}

func (v *CodeValidator) ValidateFormat(code string) Issue {
	normal := model.NormalizeCode(code)
	if normal == "" {
		return Issue{Code: "empty", Message: "ticket code is empty"}
	}
	if len(normal) < 6 {
		return Issue{Code: "short", Message: "ticket code is too short"}
	}
	if len(normal) > 32 {
		return Issue{Code: "long", Message: "ticket code is too long"}
	}
	for _, r := range normal {
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_') {
			return Issue{Code: "characters", Message: "ticket code contains unsupported characters"}
		}
	}
	for _, prefix := range v.forbiddenPrefixes {
		if strings.HasPrefix(normal, prefix) {
			return Issue{Code: "blocked", Message: "ticket code is blocked", Retryable: false}
		}
	}
	return Issue{}
}

func (v *CodeValidator) ValidateBatch(batch model.TicketBatch, codes []model.TicketCode) []Issue {
	issues := make([]Issue, 0)
	if strings.TrimSpace(batch.ID) == "" {
		issues = append(issues, Issue{Code: "batch_id", Message: "batch id is required"})
	}
	if strings.TrimSpace(batch.Source) == "" {
		issues = append(issues, Issue{Code: "source", Message: "source is required"})
	}
	if len(codes) == 0 {
		issues = append(issues, Issue{Code: "codes", Message: "at least one code is required"})
	}
	if batch.Total != len(codes) {
		issues = append(issues, Issue{Code: "total", Message: fmt.Sprintf("total %d does not match codes %d", batch.Total, len(codes))})
	}
	seen := map[string]bool{}
	for _, code := range codes {
		normal := model.NormalizeCode(code.Code)
		if seen[normal] {
			issues = append(issues, Issue{Code: "duplicate", Message: "duplicate code in batch"})
		}
		seen[normal] = true
		if issue := v.ValidateFormat(normal); issue.Code != "" {
			issues = append(issues, issue)
		}
	}
	return issues
}

func (v *CodeValidator) Classify(issue Issue) string {
	if issue.Code == "blocked" || issue.Code == "characters" {
		return "policy"
	}
	if issue.Code == "duplicate" {
		return "duplicate_input"
	}
	return "format"
}

func (v *CodeValidator) Retryable(issue Issue) bool {
	return issue.Retryable && issue.Code != "blocked" && issue.Code != "characters"
}
