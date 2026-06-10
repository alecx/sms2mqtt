// Package modem owns the AT serial conversation with the GSM modem:
// framing raw byte streams into responses, and (later) the reconnect state machine.
package modem

import "strings"

// Status is the final result code that terminates an AT command response.
type Status int

const (
	StatusOK Status = iota
	StatusError
	StatusCMEError
	StatusCMSError
)

// Response is a parsed AT command response: the intermediate data lines plus
// the terminating status (and an error code for +CME/+CMS errors).
type Response struct {
	Lines  []string
	Status Status
	Code   int
}

// ParseResponse turns the raw text of a complete AT response into a Response.
// It strips CR/LF framing and blank lines, separating the intermediate data
// lines from the final result code. It is a convenience wrapper over the
// line-driven assembler for callers that already hold a full response buffer.
func ParseResponse(raw string) Response {
	var a assembler
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(strings.TrimRight(line, "\r"))
		if line == "" {
			continue
		}
		if resp, done := a.push(line); done {
			return resp
		}
	}
	// No terminator seen: return whatever data lines accumulated.
	return Response{Lines: a.lines}
}
