package store

import (
	"path/filepath"
	"testing"
	"ticketdesk/internal/model"
)

func TestStoreValidationResultAndCounts(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.PutBatch(model.TicketBatch{ID: "b", Source: "s", Status: model.BatchPending, Total: 1}); err != nil {
		t.Fatal(err)
	}
	failure := model.FailureDetail{ID: "f", BatchID: "b", Code: "BAD123", Category: "format", Message: "bad", Retryable: false}
	if err := db.SaveValidationResult(model.ValidationAttempt{ID: "a", BatchID: "b", Code: "BAD123", Outcome: model.AttemptFailure}, &failure); err != nil {
		t.Fatal(err)
	}
	counts, err := db.SnapshotCounts()
	if err != nil {
		t.Fatal(err)
	}
	if counts["attempts"] != 1 || counts["failures"] != 1 {
		t.Fatalf("counts %#v", counts)
	}
}
