package sms

import (
	"strconv"
	"strings"
)

// StoredMessage is one entry from an AT+CMGL listing: its storage index (needed
// to delete it with AT+CMGD) and its raw PDU.
type StoredMessage struct {
	Index int
	PDU   string
}

// ParseCMGL parses the lines of an AT+CMGL response into stored messages. Each
// "+CMGL: idx,stat,alpha,len" header is paired with the PDU on the next line.
func ParseCMGL(lines []string) []StoredMessage {
	var out []StoredMessage
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if !strings.HasPrefix(line, "+CMGL:") {
			continue
		}
		fields := strings.Split(strings.TrimPrefix(line, "+CMGL:"), ",")
		idx, err := strconv.Atoi(strings.TrimSpace(fields[0]))
		if err != nil || i+1 >= len(lines) {
			continue
		}
		pdu := strings.TrimSpace(lines[i+1])
		i++
		out = append(out, StoredMessage{Index: idx, PDU: pdu})
	}
	return out
}
