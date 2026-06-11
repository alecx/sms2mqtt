package modem

import (
	"strings"
	"sync"
	"testing"
)

// autoResponder wraps a fakeSerial and replies to each written command with a
// scripted response, so command/response round-trips complete without manual
// injection timing. reply receives the command (CR-trimmed).
type autoResponder struct {
	*fakeSerial
	reply func(cmd string) string

	mu       sync.Mutex
	commands []string
}

func newAutoResponder(reply func(string) string) *autoResponder {
	return &autoResponder{fakeSerial: newFakeSerial(), reply: reply}
}

func (a *autoResponder) Write(p []byte) (int, error) {
	n, err := a.fakeSerial.Write(p)
	cmd := strings.TrimRight(string(p), "\r")
	a.mu.Lock()
	a.commands = append(a.commands, cmd)
	a.mu.Unlock()
	go a.fakeSerial.inject(a.reply(cmd))
	return n, err
}

func (a *autoResponder) sent() []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make([]string, len(a.commands))
	copy(out, a.commands)
	return out
}

func TestInitModem_RunsExpectedSequence(t *testing.T) {
	ar := newAutoResponder(func(string) string { return "\r\nOK\r\n" })
	c := NewConn(ar)
	defer c.Close()

	if err := InitModem(c); err != nil {
		t.Fatalf("InitModem error: %v", err)
	}

	got := ar.sent()
	// Must set PDU mode and enable new-message indications. Order: echo off first.
	wantContains := []string{"ATE0", "AT+CMGF=0", "AT+CNMI"}
	for _, w := range wantContains {
		found := false
		for _, c := range got {
			if strings.HasPrefix(c, w) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("init sequence %v missing a command starting with %q", got, w)
		}
	}
	if got[0] != "ATE0" {
		t.Errorf("first init command = %q, want ATE0 (echo off before anything else)", got[0])
	}
}

func TestInitModem_ErrorsIfModemRejectsACommand(t *testing.T) {
	ar := newAutoResponder(func(cmd string) string {
		if strings.HasPrefix(cmd, "AT+CMGF") {
			return "\r\nERROR\r\n"
		}
		return "\r\nOK\r\n"
	})
	c := NewConn(ar)
	defer c.Close()

	err := InitModem(c)
	if err == nil {
		t.Fatal("InitModem should error when the modem rejects a command")
	}
	if !strings.Contains(err.Error(), "CMGF") {
		t.Errorf("error %q should name the failing command (CMGF)", err)
	}
}
