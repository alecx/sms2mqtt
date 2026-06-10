package modem

import (
	"errors"
	"testing"
	"time"
)

func TestConn_CommandTimesOutOnWedgedModem(t *testing.T) {
	f := newFakeSerial()
	c := NewConn(f, WithCommandTimeout(50*time.Millisecond))
	defer c.Close()

	// Never inject a response: the modem is wedged (responsive USB, dead AT).
	_, err := c.Command("AT")
	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("Command err = %v, want ErrTimeout", err)
	}

	// A command timeout is treated as a dead connection so the Manager reconnects.
	select {
	case <-c.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("Done not closed after a command timeout")
	}
}

func TestConn_NoTimeoutByDefault(t *testing.T) {
	f := newFakeSerial()
	c := NewConn(f) // no WithCommandTimeout
	defer c.Close()

	go f.inject("\r\nOK\r\n")
	if _, err := c.Command("AT"); err != nil {
		t.Fatalf("Command err = %v, want nil with no timeout configured", err)
	}
}
