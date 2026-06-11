package sms

import "testing"

func TestParseCMGL(t *testing.T) {
	// Real AT+CMGL=4 output shape: a "+CMGL: idx,stat,alpha,len" header line
	// followed by the PDU hex on the next line.
	lines := []string{
		"+CMGL: 0,0,,162",
		"07910427025014F96012D0...",
		"+CMGL: 1,0,,130",
		"07910427025014F96412D0...",
		"+CMGL: 7,1,,49",
		"099194710218163469F004...",
	}
	got := ParseCMGL(lines)
	if len(got) != 3 {
		t.Fatalf("got %d entries, want 3", len(got))
	}
	if got[0].Index != 0 || got[0].PDU != "07910427025014F96012D0..." {
		t.Errorf("entry 0 = %+v", got[0])
	}
	if got[2].Index != 7 || got[2].PDU != "099194710218163469F004..." {
		t.Errorf("entry 2 = %+v", got[2])
	}
}

func TestParseCMGL_Empty(t *testing.T) {
	if got := ParseCMGL([]string{}); len(got) != 0 {
		t.Errorf("empty input gave %d entries, want 0", len(got))
	}
}
