package modem

import (
	"context"
	"io"
	"time"
)

// Manager keeps a modem Conn alive across resets and unplugs. It opens the
// serial port, hands the Conn to OnConnect for re-initialization, then waits for
// the connection to die and reconnects — pacing failed attempts with Backoff.
//
// All side effects are injected so the loop is testable without hardware or real
// time: Open opens the port, OnConnect runs the AT init sequence, and Sleep
// stands in for time.Sleep.
type Manager struct {
	// Open opens (or re-opens) the serial port, e.g. the /dev/serial/by-id path.
	Open func() (io.ReadWriteCloser, error)
	// OnConnect runs once per successful open to re-init the modem (CMGF=0,
	// CNMI, storage). A non-nil error tears the connection down and reconnects.
	OnConnect func(*Conn) error
	// Backoff paces reconnect attempts; it is reset after each healthy connect.
	Backoff Backoff
	// Sleep waits for d, returning early if unblocked. Defaults to time.Sleep.
	Sleep func(d time.Duration)
}

// Run supervises the modem until ctx is canceled.
func (m *Manager) Run(ctx context.Context) {
	sleep := m.Sleep
	if sleep == nil {
		sleep = time.Sleep
	}

	for ctx.Err() == nil {
		rwc, err := m.Open()
		if err != nil {
			sleep(m.Backoff.Next())
			continue
		}

		conn := NewConn(rwc)
		if m.OnConnect != nil {
			if err := m.OnConnect(conn); err != nil {
				teardown(conn, rwc)
				sleep(m.Backoff.Next())
				continue
			}
		}
		// Healthy connection: reset the backoff and wait for it to die.
		m.Backoff.Reset()
		select {
		case <-ctx.Done():
		case <-conn.Done():
		}
		teardown(conn, rwc)

		if ctx.Err() == nil {
			sleep(m.Backoff.Next())
		}
	}
}

func teardown(conn *Conn, rwc io.Closer) {
	conn.Close()
	_ = rwc.Close()
}
