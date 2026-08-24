package httpapi

import (
	"net/http"
	"strings"
)

func withJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.Header.Get("Accept"), "text/html") {
			w.Header().Set("Vary", "Accept")
		}
		next.ServeHTTP(w, r)
	})
}

func methodGuard(method string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(w, r)
	})
}
