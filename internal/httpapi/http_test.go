package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"ticketdesk/internal/batch"
	"ticketdesk/internal/store"
)

func TestHTTPCreateAndSummary(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	handler := New(batch.NewService(db)).Handler()
	body, _ := json.Marshal(map[string]any{"id": "http-1", "source": "gate", "operator": "op", "codes": []string{"ABC123"}})
	req := httptest.NewRequest(http.MethodPost, "/batches", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	get := httptest.NewRequest(http.MethodGet, "/batches/http-1", nil)
	out := httptest.NewRecorder()
	handler.ServeHTTP(out, get)
	if out.Code != http.StatusOK {
		t.Fatalf("summary status=%d", out.Code)
	}
}
