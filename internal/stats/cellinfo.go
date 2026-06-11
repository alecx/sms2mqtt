package stats

import "strings"

// ParseCellInfo extracts the area code (LAC/TAC) and cell ID from a CREG/CEREG
// response that includes location info (enabled with AT+CREG=2 / AT+CEREG=2):
//
//	+CEREG: <n>,<stat>,<tac>,<ci>[,<AcT>]
//
// Returns empty strings when no location fields are present.
func ParseCellInfo(line string) (area, cell string) {
	line = strings.TrimSpace(line)
	if i := strings.IndexByte(line, ':'); i >= 0 {
		line = line[i+1:]
	}
	fields := strings.Split(line, ",")
	if len(fields) >= 4 {
		area = unquote(fields[2])
		cell = unquote(fields[3])
	}
	return area, cell
}

func unquote(s string) string {
	return strings.Trim(strings.TrimSpace(s), `"`)
}
