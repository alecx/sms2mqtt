package modem

import (
	"reflect"
	"testing"
)

func TestAssembler_CompletesOnTerminator(t *testing.T) {
	var a assembler

	if _, done := a.push("+CSQ: 14,99"); done {
		t.Fatal("push of a data line should not complete the response")
	}

	resp, done := a.push("OK")
	if !done {
		t.Fatal("push of OK should complete the response")
	}
	if resp.Status != StatusOK {
		t.Errorf("Status = %v, want StatusOK", resp.Status)
	}
	if want := []string{"+CSQ: 14,99"}; !reflect.DeepEqual(resp.Lines, want) {
		t.Errorf("Lines = %#v, want %#v", resp.Lines, want)
	}
}

func TestAssembler_ResetsBetweenResponses(t *testing.T) {
	var a assembler
	a.push("+CSQ: 14,99")
	a.push("OK")

	// A second, independent response must not carry lines from the first.
	if _, done := a.push("+COPS: 0,0,\"Three\""); done {
		t.Fatal("data line should not complete")
	}
	resp, done := a.push("OK")
	if !done {
		t.Fatal("OK should complete")
	}
	if want := []string{"+COPS: 0,0,\"Three\""}; !reflect.DeepEqual(resp.Lines, want) {
		t.Errorf("Lines = %#v, want %#v", resp.Lines, want)
	}
}
