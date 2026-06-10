// Package config loads the sms2mqtt service configuration from the environment.
// In the Home Assistant add-on, run.sh derives these vars from the add-on
// options and the Supervisor-provided MQTT service; in dev they are set directly.
package config

import (
	"fmt"
	"strconv"
	"time"
)

// Config is the fully-resolved service configuration.
type Config struct {
	SerialDevice   string        // stable /dev/serial/by-id path to the AT port
	Baud           int           // serial baud rate
	CommandTimeout time.Duration // per-AT-command timeout
	StatsInterval  time.Duration // how often to poll + publish GSM stats
	MQTTHost       string
	MQTTPort       int
	MQTTUser       string
	MQTTPass       string
	TopicPrefix    string // MQTT topic root, e.g. "sms2mqtt"
	HealthAddr     string // listen address for the /health watchdog endpoint
	LogLevel       string
}

// Load resolves configuration using getenv to read variables, applying defaults
// and validating required fields. getenv is injected so callers (and tests) can
// supply os.Getenv or a map.
func Load(getenv func(string) string) (Config, error) {
	cfg := Config{
		SerialDevice:   getenv("SMS2MQTT_SERIAL_DEVICE"),
		Baud:           115200,
		CommandTimeout: 10 * time.Second,
		StatsInterval:  60 * time.Second,
		MQTTHost:       getenv("SMS2MQTT_MQTT_HOST"),
		MQTTPort:       1883,
		MQTTUser:       getenv("SMS2MQTT_MQTT_USER"),
		MQTTPass:       getenv("SMS2MQTT_MQTT_PASS"),
		TopicPrefix:    "sms2mqtt",
		HealthAddr:     ":8099",
		LogLevel:       "info",
	}

	if v := getenv("SMS2MQTT_TOPIC_PREFIX"); v != "" {
		cfg.TopicPrefix = v
	}
	if v := getenv("SMS2MQTT_HEALTH_ADDR"); v != "" {
		cfg.HealthAddr = v
	}
	if v := getenv("SMS2MQTT_LOG_LEVEL"); v != "" {
		cfg.LogLevel = v
	}

	if err := setInt(getenv("SMS2MQTT_BAUD"), &cfg.Baud, "SMS2MQTT_BAUD"); err != nil {
		return Config{}, err
	}
	if err := setInt(getenv("SMS2MQTT_MQTT_PORT"), &cfg.MQTTPort, "SMS2MQTT_MQTT_PORT"); err != nil {
		return Config{}, err
	}
	if err := setDuration(getenv("SMS2MQTT_COMMAND_TIMEOUT"), &cfg.CommandTimeout, "SMS2MQTT_COMMAND_TIMEOUT"); err != nil {
		return Config{}, err
	}
	if err := setDuration(getenv("SMS2MQTT_STATS_INTERVAL"), &cfg.StatsInterval, "SMS2MQTT_STATS_INTERVAL"); err != nil {
		return Config{}, err
	}

	if cfg.MQTTHost == "" {
		return Config{}, fmt.Errorf("SMS2MQTT_MQTT_HOST is required")
	}
	if cfg.SerialDevice == "" {
		return Config{}, fmt.Errorf("SMS2MQTT_SERIAL_DEVICE is required")
	}
	if cfg.Baud <= 0 {
		return Config{}, fmt.Errorf("SMS2MQTT_BAUD must be positive, got %d", cfg.Baud)
	}

	return cfg, nil
}

func setInt(v string, dst *int, name string) error {
	if v == "" {
		return nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	*dst = n
	return nil
}

func setDuration(v string, dst *time.Duration, name string) error {
	if v == "" {
		return nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	*dst = d
	return nil
}
