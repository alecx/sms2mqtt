package hass

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"
)

var topicRe = regexp.MustCompile(`^homeassistant/(sensor|binary_sensor)/sms2mqtt/([^/]+)/config$`)

func byObject(t *testing.T, msgs []DiscoveryMessage, obj string) map[string]any {
	t.Helper()
	for _, m := range msgs {
		if strings.Contains(m.Topic, "/"+obj+"/config") {
			var cfg map[string]any
			if err := json.Unmarshal(m.Payload, &cfg); err != nil {
				t.Fatalf("%s: bad JSON: %v", obj, err)
			}
			return cfg
		}
	}
	t.Fatalf("no discovery message for object %q", obj)
	return nil
}

func TestDiscoveryConfigs_TopicsAndDevice(t *testing.T) {
	msgs := DiscoveryConfigs("sms2mqtt")
	if len(msgs) < 6 {
		t.Fatalf("got %d discovery messages, want >= 6", len(msgs))
	}
	for _, m := range msgs {
		if !topicRe.MatchString(m.Topic) {
			t.Errorf("topic %q does not match the HA discovery pattern", m.Topic)
		}
		var cfg map[string]any
		if err := json.Unmarshal(m.Payload, &cfg); err != nil {
			t.Errorf("%s: bad JSON: %v", m.Topic, err)
			continue
		}
		dev, ok := cfg["device"].(map[string]any)
		if !ok {
			t.Errorf("%s: missing device block", m.Topic)
			continue
		}
		ids, _ := dev["identifiers"].([]any)
		if len(ids) == 0 || ids[0] != "sms2mqtt" {
			t.Errorf("%s: device identifiers = %v, want [sms2mqtt]", m.Topic, dev["identifiers"])
		}
		if _, ok := cfg["unique_id"].(string); !ok {
			t.Errorf("%s: missing unique_id", m.Topic)
		}
	}
}

func TestDiscoveryConfigs_SignalSensor(t *testing.T) {
	cfg := byObject(t, DiscoveryConfigs("sms2mqtt"), "signal_dbm")
	if cfg["state_topic"] != "sms2mqtt/status" {
		t.Errorf("state_topic = %v, want sms2mqtt/status", cfg["state_topic"])
	}
	if cfg["availability_topic"] != "sms2mqtt/availability" {
		t.Errorf("availability_topic = %v, want sms2mqtt/availability", cfg["availability_topic"])
	}
	if cfg["device_class"] != "signal_strength" {
		t.Errorf("device_class = %v, want signal_strength", cfg["device_class"])
	}
	if cfg["unit_of_measurement"] != "dBm" {
		t.Errorf("unit = %v, want dBm", cfg["unit_of_measurement"])
	}
	if vt, _ := cfg["value_template"].(string); !strings.Contains(vt, "signal_dbm") {
		t.Errorf("value_template = %q, want it to reference signal_dbm", vt)
	}
}

func TestDiscoveryConfigs_Connectivity(t *testing.T) {
	cfg := byObject(t, DiscoveryConfigs("sms2mqtt"), "connectivity")
	// The connectivity sensor IS the availability — it reads the availability
	// topic directly and must NOT carry its own availability_topic (or it would
	// go unavailable exactly when it should report "off").
	if cfg["state_topic"] != "sms2mqtt/availability" {
		t.Errorf("state_topic = %v, want sms2mqtt/availability", cfg["state_topic"])
	}
	if cfg["payload_on"] != "online" || cfg["payload_off"] != "offline" {
		t.Errorf("payload_on/off = %v/%v, want online/offline", cfg["payload_on"], cfg["payload_off"])
	}
	if _, has := cfg["availability_topic"]; has {
		t.Errorf("connectivity must not set availability_topic")
	}
	if cfg["device_class"] != "connectivity" {
		t.Errorf("device_class = %v, want connectivity", cfg["device_class"])
	}
}
