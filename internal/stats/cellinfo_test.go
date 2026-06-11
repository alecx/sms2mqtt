package stats

import "testing"

func TestParseCellInfo(t *testing.T) {
	cases := []struct {
		name       string
		in         string
		area, cell string
	}{
		// +CEREG: <n>,<stat>,<tac>,<ci>[,<AcT>] with location reporting on (n=2).
		{"CEREG LTE quoted", `+CEREG: 2,5,"1A2D","01A2D3E4",7`, "1A2D", "01A2D3E4"},
		{"CEREG LTE unquoted", `+CEREG: 2,1,1A2D,01A2D3E4,7`, "1A2D", "01A2D3E4"},
		{"CREG 2G/3G", `+CREG: 2,5,"ABCD","1234"`, "ABCD", "1234"},
		// No location fields (n=0) -> empty.
		{"no location", `+CEREG: 0,5`, "", ""},
		{"stat only CREG", `+CREG: 0,1`, "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			area, cell := ParseCellInfo(tc.in)
			if area != tc.area || cell != tc.cell {
				t.Errorf("ParseCellInfo(%q) = area %q cell %q, want area %q cell %q",
					tc.in, area, cell, tc.area, tc.cell)
			}
		})
	}
}
