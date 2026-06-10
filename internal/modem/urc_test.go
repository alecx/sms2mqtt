package modem

import "testing"

func TestIsURC(t *testing.T) {
	cases := []struct {
		name    string
		line    string
		pending bool // is a command currently awaiting its response?
		want    bool
	}{
		// While a command is pending, only always-unsolicited prefixes are URCs;
		// everything else is response data for the in-flight command.
		{"CMTI while pending is URC", `+CMTI: "SM",3`, true, true},
		{"RING while pending is URC", "RING", true, true},
		{"CSQ while pending is response data", "+CSQ: 14,99", true, false},
		{"CREG while pending is response data (ambiguous)", "+CREG: 1,2", true, false},

		// While idle, every non-empty line is unsolicited by definition.
		{"CMTI while idle is URC", `+CMTI: "SM",3`, false, true},
		{"CREG while idle is URC", "+CREG: 1,2", false, true},
		{"CSQ while idle is URC", "+CSQ: 14,99", false, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isURC(tc.line, tc.pending); got != tc.want {
				t.Errorf("isURC(%q, pending=%v) = %v, want %v", tc.line, tc.pending, got, tc.want)
			}
		})
	}
}
