package modem

import (
	"testing"
	"time"
)

func TestBackoff_ExponentialWithCap(t *testing.T) {
	b := Backoff{Base: 100 * time.Millisecond, Max: time.Second}

	want := []time.Duration{
		100 * time.Millisecond, // base
		200 * time.Millisecond,
		400 * time.Millisecond,
		800 * time.Millisecond,
		time.Second, // 1600ms capped at Max
		time.Second, // stays capped
	}
	for i, w := range want {
		if got := b.Next(); got != w {
			t.Errorf("Next() #%d = %v, want %v", i+1, got, w)
		}
	}
}

func TestBackoff_ResetReturnsToBase(t *testing.T) {
	b := Backoff{Base: 100 * time.Millisecond, Max: time.Second}
	b.Next()
	b.Next()
	b.Reset()
	if got := b.Next(); got != 100*time.Millisecond {
		t.Errorf("after Reset, Next() = %v, want 100ms", got)
	}
}
