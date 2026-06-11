// Package hass builds the MQTT payloads Home Assistant consumes: the received-
// SMS message (matching the existing automation contract) and the GSM status.
package hass

import (
	"encoding/json"
	"time"

	"github.com/acowan/sms2mqtt/internal/sms"
	"github.com/acowan/sms2mqtt/internal/stats"
)

// ReceivedJSON builds the payload published to sms2mqtt/received. The field
// names (number/datetime/text) match the existing automation.incoming_sms.
func ReceivedJSON(m sms.Message) ([]byte, error) {
	return json.Marshal(map[string]string{
		"number":   m.Sender,
		"datetime": m.Timestamp.Format("2006-01-02 15:04:05"),
		"text":     m.Text,
	})
}

// StatusJSON builds the retained payload for sms2mqtt/status.
func StatusJSON(s stats.Status, ts time.Time) ([]byte, error) {
	type out struct {
		stats.Status
		TS string `json:"ts"`
	}
	return json.Marshal(out{Status: s, TS: ts.Format(time.RFC3339)})
}
