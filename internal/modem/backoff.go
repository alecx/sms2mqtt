package modem

import "time"

// Backoff produces an exponentially increasing delay (Base, 2·Base, 4·Base, …)
// capped at Max, used to pace serial reconnect attempts. Reset returns it to
// Base after a successful connection. The zero value is not useful — set Base
// and Max. Not safe for concurrent use; the reconnect loop drives it serially.
type Backoff struct {
	Base    time.Duration
	Max     time.Duration
	attempt int
}

// Next returns the delay for the current attempt and advances the counter.
func (b *Backoff) Next() time.Duration {
	d := b.Base << b.attempt
	if d <= 0 || d > b.Max {
		d = b.Max
	} else {
		b.attempt++
	}
	return d
}

// Reset returns the delay sequence to Base (call after a successful connect).
func (b *Backoff) Reset() { b.attempt = 0 }
