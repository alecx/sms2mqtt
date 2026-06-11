package hass

import "encoding/json"

// DiscoveryMessage is one retained MQTT Discovery config to publish.
type DiscoveryMessage struct {
	Topic   string
	Payload []byte
}

// DiscoveryConfigs builds the HA MQTT Discovery configs that create a single
// "SMS Gateway" device with sensors derived from sms2mqtt/status plus a
// connectivity sensor from sms2mqtt/availability. Publish each retained.
func DiscoveryConfigs(topicPrefix string) []DiscoveryMessage {
	status := topicPrefix + "/status"
	avail := topicPrefix + "/availability"

	device := map[string]any{
		"identifiers":  []string{"sms2mqtt"},
		"name":         "SMS Gateway",
		"manufacturer": "Teltonika",
		"model":        "TRM240 (Quectel EC21)",
	}

	// statusSensor builds a sensor/binary_sensor reading the status JSON, with
	// availability wired so it goes unavailable when the service is down.
	statusSensor := func(component, obj, name string, extra map[string]any) DiscoveryMessage {
		cfg := map[string]any{
			"name":                  name,
			"unique_id":             "sms2mqtt_" + obj,
			"object_id":             "sms2mqtt_" + obj,
			"state_topic":           status,
			"availability_topic":    avail,
			"payload_available":     "online",
			"payload_not_available": "offline",
			"device":                device,
		}
		for k, v := range extra {
			cfg[k] = v
		}
		payload, _ := json.Marshal(cfg)
		return DiscoveryMessage{
			Topic:   "homeassistant/" + component + "/sms2mqtt/" + obj + "/config",
			Payload: payload,
		}
	}

	msgs := []DiscoveryMessage{
		statusSensor("sensor", "operator", "Operator", map[string]any{
			"value_template": "{{ value_json.operator }}", "icon": "mdi:radio-tower",
		}),
		statusSensor("sensor", "signal_dbm", "Signal Strength", map[string]any{
			"value_template": "{{ value_json.signal_dbm }}",
			"device_class":   "signal_strength", "unit_of_measurement": "dBm",
			"state_class": "measurement",
		}),
		statusSensor("sensor", "signal_pct", "Signal", map[string]any{
			"value_template":      "{{ value_json.signal_pct }}",
			"unit_of_measurement": "%", "icon": "mdi:signal", "state_class": "measurement",
		}),
		statusSensor("sensor", "access_tech", "Access Technology", map[string]any{
			"value_template": "{{ value_json.access_tech }}", "icon": "mdi:antenna",
		}),
		statusSensor("sensor", "registration", "Registration", map[string]any{
			"value_template": "{{ value_json.registration }}", "icon": "mdi:sim",
		}),
		statusSensor("sensor", "cell_id", "Cell ID", map[string]any{
			"value_template": "{{ value_json.cell_id }}", "icon": "mdi:radio-tower",
		}),
		statusSensor("binary_sensor", "roaming", "Roaming", map[string]any{
			"value_template": "{{ 'ON' if value_json.roaming else 'OFF' }}",
			"icon":           "mdi:earth",
		}),
		statusSensor("sensor", "last_seen", "Last Seen", map[string]any{
			"value_template": "{{ value_json.ts }}", "device_class": "timestamp",
		}),
	}

	// Connectivity reads the availability topic directly; it must NOT carry an
	// availability_topic of its own.
	conn, _ := json.Marshal(map[string]any{
		"name":         "Connectivity",
		"unique_id":    "sms2mqtt_connectivity",
		"object_id":    "sms2mqtt_connectivity",
		"state_topic":  avail,
		"payload_on":   "online",
		"payload_off":  "offline",
		"device_class": "connectivity",
		"device":       device,
	})
	msgs = append(msgs, DiscoveryMessage{
		Topic:   "homeassistant/binary_sensor/sms2mqtt/connectivity/config",
		Payload: conn,
	})

	return msgs
}
