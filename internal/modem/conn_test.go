package modem

import (
	"io"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

// fakeSerial is an in-memory io.ReadWriter standing in for the modem's serial
// port. The test injects modem→host bytes via inject(); commands written by the
// Conn (host→modem) are captured in writes.
type fakeSerial struct {
	out  *io.PipeReader
	outW *io.PipeWriter

	mu     sync.Mutex
	writes []string
}

func newFakeSerial() *fakeSerial {
	r, w := io.Pipe()
	return &fakeSerial{out: r, outW: w}
}

func (f *fakeSerial) Read(p []byte) (int, error) { return f.out.Read(p) }

// Close ends the stream (simulates the device going away).
func (f *fakeSerial) Close() error {
	_ = f.outW.Close()
	return f.out.Close()
}

func (f *fakeSerial) Write(p []byte) (int, error) {
	f.mu.Lock()
	f.writes = append(f.writes, string(p))
	f.mu.Unlock()
	return len(p), nil
}

// inject simulates the modem emitting bytes. It blocks until the Conn's reader
// consumes them (io.Pipe is synchronous), so call it from a goroutine.
func (f *fakeSerial) inject(s string) { _, _ = io.WriteString(f.outW, s) }

func (f *fakeSerial) lastWrite() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.writes) == 0 {
		return ""
	}
	return f.writes[len(f.writes)-1]
}

func TestConn_CommandReturnsResponse(t *testing.T) {
	f := newFakeSerial()
	c := NewConn(f)
	defer c.Close()

	go f.inject("\r\n+CSQ: 14,99\r\n\r\nOK\r\n")

	resp, err := c.Command("AT+CSQ")
	if err != nil {
		t.Fatalf("Command error: %v", err)
	}
	if resp.Status != StatusOK {
		t.Errorf("Status = %v, want StatusOK", resp.Status)
	}
	if want := []string{"+CSQ: 14,99"}; !reflect.DeepEqual(resp.Lines, want) {
		t.Errorf("Lines = %#v, want %#v", resp.Lines, want)
	}
	if w := f.lastWrite(); !strings.HasPrefix(w, "AT+CSQ") || !strings.HasSuffix(w, "\r") {
		t.Errorf("command written = %q, want AT+CSQ terminated by CR", w)
	}
}

func TestConn_DeliversIdleURC(t *testing.T) {
	f := newFakeSerial()
	c := NewConn(f)
	defer c.Close()

	go f.inject("\r\n+CMTI: \"SM\",3\r\n")

	select {
	case urc := <-c.URCs():
		if urc != `+CMTI: "SM",3` {
			t.Errorf("URC = %q, want +CMTI: \"SM\",3", urc)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for idle URC")
	}
}

func TestConn_RoutesURCDuringCommand(t *testing.T) {
	f := newFakeSerial()
	c := NewConn(f)
	defer c.Close()

	// A new-SMS URC arrives interleaved before the command's terminator. It must
	// be routed to URCs() and must NOT pollute the command's response.
	go f.inject("\r\n+CMTI: \"SM\",3\r\n+CSQ: 14,99\r\n\r\nOK\r\n")

	resp, err := c.Command("AT+CSQ")
	if err != nil {
		t.Fatalf("Command error: %v", err)
	}
	if want := []string{"+CSQ: 14,99"}; !reflect.DeepEqual(resp.Lines, want) {
		t.Errorf("Lines = %#v, want just the CSQ line", resp.Lines)
	}
	select {
	case urc := <-c.URCs():
		if urc != `+CMTI: "SM",3` {
			t.Errorf("URC = %q, want the CMTI", urc)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("CMTI URC was not delivered")
	}
}

func TestConn_CommandErrsAfterDisconnect(t *testing.T) {
	f := newFakeSerial()
	c := NewConn(f)
	defer c.Close()

	// Modem vanishes mid-conversation: close the read side so the line stream EOFs.
	go func() { _ = f.outW.Close() }()

	if _, err := c.Command("AT+CSQ"); err == nil {
		t.Fatal("Command should return an error after disconnect, got nil")
	}
}
