package store

import (
	"path/filepath"
	"testing"

	"ticketdesk/internal/model"
)

func TestPersistenceSurvivesReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tickets.db")
	db, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	batch := model.TicketBatch{ID: "persist-1", Source: "gate-a", CreatedBy: "op", Status: model.BatchPending, Total: 1}
	if err := db.PutBatch(batch); err != nil {
		t.Fatal(err)
	}
	if err := db.PutCode(model.TicketCode{BatchID: batch.ID, Code: "ABC123", State: model.CodePending}); err != nil {
		t.Fatal(err)
	}
	if err := db.PutTask(model.WorkerTask{ID: "task-1", BatchID: batch.ID, Code: "ABC123", State: model.TaskQueued}); err != nil {
		t.Fatal(err)
	}
	if err := db.PutFailure(model.FailureDetail{ID: "failure-1", BatchID: batch.ID, Code: "ABC123", Category: "format", Message: "bad", Retryable: true}); err != nil {
		t.Fatal(err)
	}
	if err := db.PutAttempt(model.ValidationAttempt{ID: "attempt-1", BatchID: batch.ID, Code: "ABC123", Outcome: model.AttemptFailure}); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	got, err := reopened.GetBatch(batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != batch.ID {
		t.Fatalf("got %q", got.ID)
	}
	codes, _ := reopened.ListCodes(batch.ID)
	if len(codes) != 1 {
		t.Fatalf("codes %d", len(codes))
	}
	tasks, _ := reopened.ListTasks(batch.ID)
	if len(tasks) != 1 {
		t.Fatalf("tasks %d", len(tasks))
	}
	failures, _ := reopened.ListFailures(batch.ID)
	if len(failures) != 1 {
		t.Fatalf("failures %d", len(failures))
	}
	attempts, _ := reopened.ListAttempts(batch.ID)
	if len(attempts) != 1 {
		t.Fatalf("attempts %d", len(attempts))
	}
}
