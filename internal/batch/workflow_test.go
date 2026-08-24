package batch

import (
	"path/filepath"
	"testing"
	"ticketdesk/internal/model"
	"ticketdesk/internal/store"
)

func newTestService(t *testing.T) (*Service, *store.Store) {
	t.Helper()
	db, err := store.Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatal(err)
	}
	return NewService(db), db
}

func TestWorkflowOne(t *testing.T) {
	s, db := newTestService(t)
	defer db.Close()
	batch, issues, err := s.RegisterBatch("wf-one", "gate-a", "worker-a", []string{"ABC123", "XYZ789"})
	if err != nil || len(issues) != 0 {
		t.Fatalf("register err=%v issues=%v", err, issues)
	}
	tasks, err := s.QueueBatch(batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, task := range tasks {
		if _, err := s.ProcessTask(task); err != nil {
			t.Fatal(err)
		}
	}
	summary, err := s.Summary(batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	if summary.Batch.Status != model.BatchComplete || summary.Percent != 100 {
		t.Fatalf("summary=%+v", summary)
	}
}

func TestWorkflowTwo(t *testing.T) {
	s, db := newTestService(t)
	defer db.Close()
	batch, issues, err := s.RegisterBatch("wf-two", "gate-b", "worker-b", []string{"VOID-123"})
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) == 0 {
		t.Fatal("invalid batch accepted")
	}
	if batch.ID != "wf-two" {
		t.Fatal("batch identity lost")
	}
}

func TestWorkflowThree(t *testing.T) {
	s, db := newTestService(t)
	defer db.Close()
	batch, _, err := s.RegisterBatch("wf-three", "gate-c", "worker-c", []string{"ABC123"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.QueueBatch(batch.ID); err != nil {
		t.Fatal(err)
	}
	if got, err := s.Tasks(batch.ID); err != nil || len(got) != 1 {
		t.Fatalf("tasks=%v err=%v", got, err)
	}
}
