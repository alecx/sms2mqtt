package modem

import (
	"context"
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// countingRWC wraps a fakeSerial to count Close calls, so tests can assert the
// manager tears down the port between reconnects.
type countingRWC struct {
	*fakeSerial
	closes *int32
}

func (c countingRWC) Close() error {
	atomic.AddInt32(c.closes, 1)
	return c.fakeSerial.Close()
}

func TestManager_RetriesOpenWithBackoffThenConnects(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var openCalls, connects int32
	var mu sync.Mutex
	var sleeps []time.Duration

	open := func() (io.ReadWriteCloser, error) {
		if atomic.AddInt32(&openCalls, 1) <= 2 {
			return nil, errors.New("no device")
		}
		// Succeed, but the device immediately disconnects (EOF) -> reconnect.
		f := newFakeSerial()
		go f.Close()
		return f, nil
	}
	m := &Manager{
		Open:      open,
		OnConnect: func(*Conn) error { atomic.AddInt32(&connects, 1); return nil },
		Backoff:   Backoff{Base: time.Millisecond, Max: 10 * time.Millisecond},
		Sleep: func(d time.Duration) {
			mu.Lock()
			sleeps = append(sleeps, d)
			mu.Unlock()
		},
	}

	go func() {
		for atomic.LoadInt32(&connects) < 2 {
			time.Sleep(time.Millisecond)
		}
		cancel()
	}()
	m.Run(ctx)

	if got := atomic.LoadInt32(&openCalls); got < 4 {
		t.Errorf("openCalls = %d, want >= 4 (2 failures + >= 2 successes)", got)
	}
	if got := atomic.LoadInt32(&connects); got < 2 {
		t.Errorf("connects = %d, want >= 2", got)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(sleeps) < 2 || sleeps[0] != time.Millisecond || sleeps[1] != 2*time.Millisecond {
		t.Errorf("open-failure backoff sleeps = %v, want first two 1ms,2ms", sleeps)
	}
}

func TestManager_TearsDownAndRetriesWhenOnConnectFails(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var attempts, closes int32

	open := func() (io.ReadWriteCloser, error) {
		// Long-lived, idle port (never EOFs on its own).
		return countingRWC{fakeSerial: newFakeSerial(), closes: &closes}, nil
	}
	m := &Manager{
		Open: open,
		OnConnect: func(*Conn) error {
			if atomic.AddInt32(&attempts, 1) == 1 {
				return errors.New("re-init failed")
			}
			return nil
		},
		Backoff: Backoff{Base: time.Millisecond, Max: 10 * time.Millisecond},
		Sleep:   func(time.Duration) {},
	}

	go func() {
		for atomic.LoadInt32(&attempts) < 2 {
			time.Sleep(time.Millisecond)
		}
		cancel()
	}()
	m.Run(ctx)

	if got := atomic.LoadInt32(&attempts); got < 2 {
		t.Errorf("onConnect attempts = %d, want >= 2 (failed then retried)", got)
	}
	if got := atomic.LoadInt32(&closes); got < 1 {
		t.Errorf("port closes = %d, want >= 1 (failed connect torn down)", got)
	}
}
