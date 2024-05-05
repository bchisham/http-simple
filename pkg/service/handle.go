package service

import "net/http"

func (s *service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.NotImplemented(w, r)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	// Handle the health check
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}
