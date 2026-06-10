package modem

import (
	"bufio"
	"errors"
	"io"
	"strings"
)

// ErrDisconnected is returned by Command when the serial stream ended (EOF or
// read error) before the command's response completed — e.g. the modem reset or
// was unplugged. The reconnect layer turns this into a re-open.
var ErrDisconnected = errors.New("modem: serial stream disconnected")

// Conn is the single owner of an AT serial port. One goroutine reads the port
// and either completes the in-flight command's response or routes unsolicited
// result codes (URCs) to URCs(). All command/URC traffic is serialized through
// this goroutine, so nothing else may read or write the underlying port.
//
// URCs() MUST be drained continuously: a URC that arrives mid-command blocks the
// owner goroutine until it is consumed. The SMS reader is the designated drainer.
type Conn struct {
	rw       io.ReadWriter
	requests chan request
	urcs     chan string
	done     chan struct{}
}

type request struct {
	cmd   string
	reply chan result
}

type result struct {
	resp Response
	err  error
}

// NewConn starts the owner goroutine reading rw and returns a ready Conn.
func NewConn(rw io.ReadWriter) *Conn {
	c := &Conn{
		rw:       rw,
		requests: make(chan request),
		urcs:     make(chan string, 32),
		done:     make(chan struct{}),
	}
	go c.run()
	return c
}

// URCs returns the stream of unsolicited result codes. Drain it continuously.
func (c *Conn) URCs() <-chan string { return c.urcs }

// Done is closed when the connection ends — either the serial stream
// disconnected (EOF/read error) or Close was called. The reconnect Manager
// waits on it to detect a dead modem.
func (c *Conn) Done() <-chan struct{} { return c.done }

// Command writes cmd to the modem (appending the AT carriage return) and returns
// the assembled response. It returns ErrDisconnected if the stream ends first.
func (c *Conn) Command(cmd string) (Response, error) {
	req := request{cmd: cmd, reply: make(chan result, 1)}
	select {
	case c.requests <- req:
	case <-c.done:
		return Response{}, ErrDisconnected
	}
	res := <-req.reply
	return res.resp, res.err
}

// Close stops the owner goroutine. The underlying port is the caller's to close.
func (c *Conn) Close() {
	select {
	case <-c.done:
	default:
		close(c.done)
	}
}

func (c *Conn) run() {
	lines := make(chan string)
	go scanLines(c.rw, lines)

	for {
		select {
		case <-c.done:
			return
		case req := <-c.requests:
			c.serveCommand(req, lines)
		case line, ok := <-lines:
			if !ok {
				c.Close()
				return
			}
			c.emit(line) // idle: every line is unsolicited
		}
	}
}

// serveCommand writes the command and consumes lines until the response
// completes, diverting any URCs that arrive in the meantime.
func (c *Conn) serveCommand(req request, lines <-chan string) {
	if _, err := io.WriteString(c.rw, req.cmd+"\r"); err != nil {
		req.reply <- result{err: err}
		return
	}
	var a assembler
	for {
		select {
		case <-c.done:
			req.reply <- result{err: ErrDisconnected}
			return
		case line, ok := <-lines:
			if !ok {
				c.Close()
				req.reply <- result{err: ErrDisconnected}
				return
			}
			if isURC(line, true) {
				c.emit(line)
				continue
			}
			if resp, done := a.push(line); done {
				req.reply <- result{resp: resp}
				return
			}
		}
	}
}

func (c *Conn) emit(line string) {
	select {
	case c.urcs <- line:
	case <-c.done:
	}
}

// scanLines reads r line by line, emitting trimmed non-empty lines on out and
// closing out when the stream ends.
func scanLines(r io.Reader, out chan<- string) {
	defer close(out)
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 4096), 64*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		out <- line
	}
}
