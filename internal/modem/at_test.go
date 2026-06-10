package modem

import (
	"reflect"
	"testing"
)

func TestParseResponse_OKWithDataLine(t *testing.T) {
	// A typical AT+CSQ response: framing CR/LFs, one data line, terminated by OK.
	raw := "\r\n+CSQ: 14,99\r\n\r\nOK\r\n"

	got := ParseResponse(raw)

	if got.Status != StatusOK {
		t.Errorf("Status = %v, want StatusOK", got.Status)
	}
	want := []string{"+CSQ: 14,99"}
	if !reflect.DeepEqual(got.Lines, want) {
		t.Errorf("Lines = %#v, want %#v", got.Lines, want)
	}
}

func TestParseResponse_BareOK(t *testing.T) {
	got := ParseResponse("\r\nOK\r\n")

	if got.Status != StatusOK {
		t.Errorf("Status = %v, want StatusOK", got.Status)
	}
	if len(got.Lines) != 0 {
		t.Errorf("Lines = %#v, want empty", got.Lines)
	}
}

func TestParseResponse_PlainError(t *testing.T) {
	got := ParseResponse("\r\nERROR\r\n")

	if got.Status != StatusError {
		t.Errorf("Status = %v, want StatusError", got.Status)
	}
}

func TestParseResponse_CMEError(t *testing.T) {
	// +CME ERROR: 10 == "SIM not inserted"
	got := ParseResponse("\r\n+CME ERROR: 10\r\n")

	if got.Status != StatusCMEError {
		t.Errorf("Status = %v, want StatusCMEError", got.Status)
	}
	if got.Code != 10 {
		t.Errorf("Code = %d, want 10", got.Code)
	}
}

func TestParseResponse_CMSError(t *testing.T) {
	// +CMS ERROR: 500 is a generic SMS-layer failure.
	got := ParseResponse("\r\n+CMS ERROR: 500\r\n")

	if got.Status != StatusCMSError {
		t.Errorf("Status = %v, want StatusCMSError", got.Status)
	}
	if got.Code != 500 {
		t.Errorf("Code = %d, want 500", got.Code)
	}
}
