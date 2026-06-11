package health

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHealth_UnhealthyBeforeAnyBeat(t *testing.T) {
	h := New(time.Minute)
	if code := serve(h); code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503 before any beat", code)
	}
}

func TestHealth_HealthyAfterBeat(t *testing.T) {
	h := New(time.Minute)
	h.Beat()
	if code := serve(h); code != http.StatusOK {
		t.Errorf("status = %d, want 200 after a fresh beat", code)
	}
}

func TestHealth_UnhealthyWhenStale(t *testing.T) {
	h := New(time.Minute)
	h.Beat()
	// Advance the clock past maxAge.
	base := time.Now()
	h.now = func() time.Time { return base.Add(2 * time.Minute) }
	if code := serve(h); code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503 when the last beat is stale", code)
	}
}

func serve(h *Health) int {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	return rec.Code
}
