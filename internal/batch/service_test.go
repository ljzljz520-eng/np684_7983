package batch

import (
	"path/filepath"
	"testing"
	"ticketdesk/internal/store"
)

func TestRetryFailuresCreatesTasks(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	s := NewService(db)
	batch, _, err := s.RegisterBatch("retry", "gate", "op", []string{"ABC123"})
	if err != nil {
		t.Fatal(err)
	}
	tasks, err := s.QueueBatch(batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	db.SetConsumeBarrier(1)
	if _, err := s.ProcessTask(tasks[0]); err != nil {
		t.Fatal(err)
	}
	retries, err := s.RetryFailures(batch.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(retries) != 0 {
		t.Fatalf("unexpected retry tasks: %v", retries)
	}
}
