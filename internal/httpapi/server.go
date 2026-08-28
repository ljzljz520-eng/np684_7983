package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"ticketdesk/internal/batch"
)

type Server struct {
	service *batch.Service
	mux     *http.ServeMux
}
type createRequest struct {
	ID       string   `json:"id"`
	Source   string   `json:"source"`
	Operator string   `json:"operator"`
	Codes    []string `json:"codes"`
}
type errorResponse struct {
	Error  string `json:"error"`
	Issues any    `json:"issues,omitempty"`
}

func New(service *batch.Service) *Server {
	s := &Server{service: service, mux: http.NewServeMux()}
	s.mux.Handle("/health", withJSON(methodGuard(http.MethodGet, http.HandlerFunc(s.health))))
	s.mux.HandleFunc("/batches", s.batches)
	s.mux.HandleFunc("/batches/", s.batchRoute)
	s.mux.HandleFunc("/failures.csv", s.FailureCSV)
	return s
}

func (s *Server) Handler() http.Handler { return logging(s.mux) }
func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) batches(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		batches, err := s.service.Batches()
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, batches)
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var input createRequest
	if err := decodeBody(r, &input); err != nil {
		writeErrorStatus(w, http.StatusBadRequest, err)
		return
	}
	batch, issues, err := s.service.RegisterBatch(input.ID, input.Source, input.Operator, input.Codes)
	if err != nil {
		writeError(w, err)
		return
	}
	if len(issues) > 0 {
		writeJSON(w, http.StatusUnprocessableEntity, errorResponse{Error: "batch rejected", Issues: issues})
		return
	}
	writeJSON(w, http.StatusCreated, batch)
}

func (s *Server) batchRoute(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/batches/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	id := parts[0]
	if len(parts) == 1 && r.Method == http.MethodGet {
		summary, err := s.service.Summary(id)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, summary)
		return
	}
	if len(parts) != 2 || r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	switch parts[1] {
	case "queue":
		tasks, err := s.service.QueueBatch(id)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusAccepted, tasks)
	case "retry":
		tasks, err := s.service.RetryFailures(id)
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusAccepted, tasks)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	setCacheHeaders(w)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func writeError(w http.ResponseWriter, err error) {
	writeErrorStatus(w, http.StatusInternalServerError, err)
}
func writeErrorStatus(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, errorResponse{Error: err.Error()})
}
func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Request-ID", requestID(r))
		if strings.TrimSpace(r.URL.Path) == "" {
			writeErrorStatus(w, http.StatusNotFound, fmt.Errorf("path is required"))
			return
		}
		next.ServeHTTP(w, r)
	})
}
