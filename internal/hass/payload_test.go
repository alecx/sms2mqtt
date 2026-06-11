package hass

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/acowan/sms2mqtt/internal/sms"
)

func TestReceivedJSON_MatchesExistingContract(t *testing.T) {
	msg := sms.Message{
		Sender:    "Vodafone",
		Text:      "Hello world",
		Timestamp: time.Date(2026, 6, 11, 19, 30, 5, 0, time.UTC),
	}

	raw, err := ReceivedJSON(msg)
	if err != nil {
		t.Fatalf("ReceivedJSON error: %v", err)
	}

	// The existing automation.incoming_sms reads payload_json.number/datetime/text.
	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("not valid JSON: %v", err)
	}
	if got["number"] != "Vodafone" {
		t.Errorf("number = %v, want Vodafone", got["number"])
	}
	if got["text"] != "Hello world" {
		t.Errorf("text = %v, want Hello world", got["text"])
	}
	if got["datetime"] != "2026-06-11 19:30:05" {
		t.Errorf("datetime = %v, want 2026-06-11 19:30:05", got["datetime"])
	}
}
