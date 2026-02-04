// Package health provides health check endpoints for the Statigo framework.
package health

import (
	"encoding/json"
	"net/http"
	"time"
)

// Handler handles health check HTTP requests.
type Handler struct {
	checker *Checker
}

// NewHandler creates a new health handler.
func NewHandler(checkTimeout time.Duration) *Handler {
	return &Handler{
		checker: NewChecker(checkTimeout),
	}
}

// Liveness is a simple liveness probe that returns OK if the app is running.
// Use for Kubernetes liveness probes or simple uptime checks.
func (h *Handler) Liveness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// Readiness checks external dependencies and returns detailed status.
// Use for Kubernetes readiness probes or monitoring.
func (h *Handler) Readiness(w http.ResponseWriter, r *http.Request) {
	status := h.checker.CheckAll(r.Context())

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// Return 503 if any check is down, 200 if all are up or degraded
	if status.Status == "down" {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(status)
}
