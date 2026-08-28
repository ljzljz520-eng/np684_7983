package httpapi

import (
	"net/http"

	"ticketdesk/internal/report"
)

func (s *Server) FailureCSV(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	id := r.URL.Query().Get("batch")
	summary, err := s.service.Summary(id)
	if err != nil {
		writeError(w, err)
		return
	}
	w.Header().Set("Content-Type", "text/csv")
	_ = report.WriteFailureCSV(w, summary.Failures)
}
