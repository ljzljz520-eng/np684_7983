package validate

import (
	"testing"
	"ticketdesk/internal/model"
)

func TestValidatorClassifiesCodes(t *testing.T) {
	v := NewCodeValidator()
	if v.ValidateFormat("ABC123").Code != "" {
		t.Fatal("valid code rejected")
	}
	if v.ValidateFormat("VOID-123").Code != "blocked" {
		t.Fatal("blocked code accepted")
	}
	if v.ValidateFormat("a").Code != "short" {
		t.Fatal("short code accepted")
	}
	issues := v.ValidateBatch(model.TicketBatch{ID: "b", Source: "s", Total: 2}, []model.TicketCode{{BatchID: "b", Code: "ABC123", State: model.CodePending}, {BatchID: "b", Code: "ABC123", State: model.CodePending}})
	if len(issues) == 0 {
		t.Fatal("duplicate not detected")
	}
}
