package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func decodeBody(r *http.Request, target any) error {
	if r.Body == nil {
		return fmt.Errorf("request body is required")
	}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	return nil
}

func requestID(r *http.Request) string {
	value := r.Header.Get("X-Request-ID")
	if value == "" {
		return "local-request"
	}
	return value
}

func setCacheHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Request-Trace", "ticketdesk")
}
