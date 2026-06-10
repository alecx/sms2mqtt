package modem

import "strings"

// urcOnlyPrefixes are line prefixes that the modem only ever emits unsolicited —
// they are never the response to a command we issued. The most important here is
// +CMTI (new SMS stored). Lines like +CREG/+CSQ are deliberately absent: they
// are also valid responses to AT+CREG?/AT+CSQ, so while a command is pending
// they must be treated as response data, not URCs.
var urcOnlyPrefixes = []string{
	"+CMTI:", // new SMS arrived, stored
	"+CMT:",  // new SMS delivered directly (CNMI mode 2)
	"+CDS:",  // SMS status report
	"+CBM:",  // cell broadcast message
	"RING",   // incoming call
	"NO CARRIER",
	"+CLIP:",
}

// isURC reports whether a trimmed, non-empty line should be routed to the URC
// stream rather than treated as part of a command's response.
//
// When no command is pending (idle), any line the modem emits is by definition
// unsolicited. When a command is pending, only the always-unsolicited prefixes
// are diverted; everything else belongs to the in-flight response.
func isURC(line string, pending bool) bool {
	if !pending {
		return true
	}
	for _, p := range urcOnlyPrefixes {
		if strings.HasPrefix(line, p) {
			return true
		}
	}
	return false
}
