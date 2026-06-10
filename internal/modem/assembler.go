package modem

import (
	"strconv"
	"strings"
)

// assembler accumulates the lines of a single AT response as they arrive,
// reporting completion when a final result code (OK/ERROR/+CME/+CMS) is pushed.
// It is the line-driven core that the serial reader feeds.
type assembler struct {
	lines []string
}

// push feeds one already-trimmed, non-empty line. When the line is a final
// result code it returns the completed Response and done=true, and the
// assembler resets for the next response. Otherwise the line is buffered as a
// data line and done=false.
func (a *assembler) push(line string) (Response, bool) {
	if status, code, terminal := classifyLine(line); terminal {
		resp := Response{Lines: a.lines, Status: status, Code: code}
		a.lines = nil
		return resp, true
	}
	a.lines = append(a.lines, line)
	return Response{}, false
}

// classifyLine reports whether a trimmed line is a final result code, and if so
// its Status and numeric error code (0 for OK/ERROR and verbose codes).
func classifyLine(line string) (status Status, code int, terminal bool) {
	switch {
	case line == "OK":
		return StatusOK, 0, true
	case line == "ERROR":
		return StatusError, 0, true
	case strings.HasPrefix(line, "+CME ERROR:"):
		return StatusCMEError, parseErrorCode(line), true
	case strings.HasPrefix(line, "+CMS ERROR:"):
		return StatusCMSError, parseErrorCode(line), true
	}
	return StatusOK, 0, false
}

// parseErrorCode extracts the numeric code from a "+CME ERROR: 10" /
// "+CMS ERROR: 500" line. Non-numeric (verbose) codes yield 0.
func parseErrorCode(line string) int {
	_, after, found := strings.Cut(line, ":")
	if !found {
		return 0
	}
	n, err := strconv.Atoi(strings.TrimSpace(after))
	if err != nil {
		return 0
	}
	return n
}
