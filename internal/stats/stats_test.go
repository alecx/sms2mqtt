package stats

import "testing"

func TestParseCSQ(t *testing.T) {
	dbm, pct, ok := ParseCSQ("+CSQ: 19,99")
	if !ok {
		t.Fatal("ParseCSQ ok = false, want true for rssi 19")
	}
	if dbm != -75 { // -113 + 2*19
		t.Errorf("dBm = %d, want -75", dbm)
	}
	if pct < 55 || pct > 65 { // 19/31 ≈ 61%
		t.Errorf("pct = %d, want ~61", pct)
	}
}

func TestParseCSQ_Unknown(t *testing.T) {
	if _, _, ok := ParseCSQ("+CSQ: 99,99"); ok {
		t.Error("ParseCSQ ok = true for rssi 99 (unknown), want false")
	}
}

func TestParseCOPS(t *testing.T) {
	op, act := ParseCOPS(`+COPS: 0,0,"vodafone IE Vodafone RO",7`)
	if op != "vodafone IE Vodafone RO" {
		t.Errorf("operator = %q, want %q", op, "vodafone IE Vodafone RO")
	}
	if act != "LTE" { // access tech 7 == E-UTRAN
		t.Errorf("act = %q, want LTE", act)
	}
}

func TestParseCREG(t *testing.T) {
	cases := []struct {
		in         string
		registered bool
		roaming    bool
	}{
		{"+CREG: 0,5", true, true},   // registered, roaming
		{"+CREG: 0,1", true, false},  // registered, home
		{"+CREG: 0,2", false, false}, // searching
		{"+CREG: 0,0", false, false}, // not registered
	}
	for _, tc := range cases {
		reg, roam := ParseCREG(tc.in)
		if reg != tc.registered || roam != tc.roaming {
			t.Errorf("ParseCREG(%q) = reg %v roam %v, want reg %v roam %v",
				tc.in, reg, roam, tc.registered, tc.roaming)
		}
	}
}
