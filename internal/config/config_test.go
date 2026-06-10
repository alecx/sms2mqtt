package config

import (
	"testing"
	"time"
)

// envFrom returns a getenv func backed by a map, for deterministic tests.
func envFrom(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestLoad_AppliesDefaults(t *testing.T) {
	cfg, err := Load(envFrom(map[string]string{
		"SMS2MQTT_MQTT_HOST":     "core-mosquitto",
		"SMS2MQTT_SERIAL_DEVICE": "/dev/serial/by-id/usb-Teltonika_TRM240-if01",
	}))
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if cfg.Baud != 115200 {
		t.Errorf("Baud = %d, want default 115200", cfg.Baud)
	}
	if cfg.MQTTPort != 1883 {
		t.Errorf("MQTTPort = %d, want default 1883", cfg.MQTTPort)
	}
	if cfg.StatsInterval != 60*time.Second {
		t.Errorf("StatsInterval = %v, want default 60s", cfg.StatsInterval)
	}
	if cfg.CommandTimeout != 10*time.Second {
		t.Errorf("CommandTimeout = %v, want default 10s", cfg.CommandTimeout)
	}
	if cfg.TopicPrefix != "sms2mqtt" {
		t.Errorf("TopicPrefix = %q, want default sms2mqtt", cfg.TopicPrefix)
	}
	if cfg.HealthAddr != ":8099" {
		t.Errorf("HealthAddr = %q, want default :8099", cfg.HealthAddr)
	}
}

func TestLoad_EnvOverrides(t *testing.T) {
	cfg, err := Load(envFrom(map[string]string{
		"SMS2MQTT_MQTT_HOST":      "broker.local",
		"SMS2MQTT_MQTT_PORT":      "8883",
		"SMS2MQTT_MQTT_USER":      "addons",
		"SMS2MQTT_MQTT_PASS":      "secret",
		"SMS2MQTT_SERIAL_DEVICE":  "/dev/ttyUSB1",
		"SMS2MQTT_BAUD":           "9600",
		"SMS2MQTT_STATS_INTERVAL": "30s",
		"SMS2MQTT_TOPIC_PREFIX":   "gsm",
	}))
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if cfg.MQTTPort != 8883 {
		t.Errorf("MQTTPort = %d, want 8883", cfg.MQTTPort)
	}
	if cfg.MQTTUser != "addons" || cfg.MQTTPass != "secret" {
		t.Errorf("MQTT creds = %q/%q, want addons/secret", cfg.MQTTUser, cfg.MQTTPass)
	}
	if cfg.Baud != 9600 {
		t.Errorf("Baud = %d, want 9600", cfg.Baud)
	}
	if cfg.StatsInterval != 30*time.Second {
		t.Errorf("StatsInterval = %v, want 30s", cfg.StatsInterval)
	}
	if cfg.TopicPrefix != "gsm" {
		t.Errorf("TopicPrefix = %q, want gsm", cfg.TopicPrefix)
	}
}

func TestLoad_RequiresMQTTHost(t *testing.T) {
	_, err := Load(envFrom(map[string]string{
		"SMS2MQTT_SERIAL_DEVICE": "/dev/ttyUSB1",
	}))
	if err == nil {
		t.Fatal("Load should error when SMS2MQTT_MQTT_HOST is unset")
	}
}

func TestLoad_RequiresSerialDevice(t *testing.T) {
	_, err := Load(envFrom(map[string]string{
		"SMS2MQTT_MQTT_HOST": "core-mosquitto",
	}))
	if err == nil {
		t.Fatal("Load should error when SMS2MQTT_SERIAL_DEVICE is unset")
	}
}

func TestLoad_RejectsBadBaud(t *testing.T) {
	_, err := Load(envFrom(map[string]string{
		"SMS2MQTT_MQTT_HOST":     "core-mosquitto",
		"SMS2MQTT_SERIAL_DEVICE": "/dev/ttyUSB1",
		"SMS2MQTT_BAUD":          "not-a-number",
	}))
	if err == nil {
		t.Fatal("Load should error on a non-numeric baud")
	}
}
