// Package health exposes a liveness endpoint for the Home Assistant Supervisor
// watchdog. The service "beats" after each successful stats publish (which
// requires both a responsive modem and a live MQTT connection); /health reports
// unhealthy if no beat has landed within maxAge, so the watchdog restarts a
// wedged add-on.
package health

import (
	"net/http"
	"sync"
	"time"
)

// Health tracks the last successful heartbeat and serves it over HTTP.
type Health struct {
	mu     sync.Mutex
	last   time.Time
	maxAge time.Duration
	now    func() time.Time
}

// New returns a Health that reports unhealthy once a beat is older than maxAge.
func New(maxAge time.Duration) *Health {
	return &Health{maxAge: maxAge, now: time.Now}
}

// Beat records a successful heartbeat.
func (h *Health) Beat() {
	h.mu.Lock()
	h.last = h.now()
	h.mu.Unlock()
}

// ServeHTTP returns 200 when the last beat is within maxAge, else 503.
func (h *Health) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	h.mu.Lock()
	last := h.last
	h.mu.Unlock()

	if last.IsZero() || h.now().Sub(last) > h.maxAge {
		http.Error(w, "unhealthy", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
