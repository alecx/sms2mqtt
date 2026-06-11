#!/usr/bin/with-contenv bashio
# Map add-on options + the Supervisor-provided MQTT service into the env the
# sms2mqtt binary reads. Secrets stay out of the command line.
set -e

export SMS2MQTT_SERIAL_DEVICE="$(bashio::config 'serial_device')"
export SMS2MQTT_BAUD="$(bashio::config 'baud')"
export SMS2MQTT_STATS_INTERVAL="$(bashio::config 'stats_interval')"
export SMS2MQTT_LOG_LEVEL="$(bashio::config 'log_level')"
export SMS2MQTT_HEALTH_ADDR=":8099"

# An explicit mqtt_host wins (e.g. point at another HA instance's broker so its
# automations react). Otherwise use the Supervisor's Mosquitto.
if bashio::config.has_value 'mqtt_host'; then
  export SMS2MQTT_MQTT_HOST="$(bashio::config 'mqtt_host')"
  export SMS2MQTT_MQTT_PORT="$(bashio::config 'mqtt_port')"
  export SMS2MQTT_MQTT_USER="$(bashio::config 'mqtt_user')"
  export SMS2MQTT_MQTT_PASS="$(bashio::config 'mqtt_pass')"
elif bashio::services.available "mqtt"; then
  export SMS2MQTT_MQTT_HOST="$(bashio::services mqtt 'host')"
  export SMS2MQTT_MQTT_PORT="$(bashio::services mqtt 'port')"
  export SMS2MQTT_MQTT_USER="$(bashio::services mqtt 'username')"
  export SMS2MQTT_MQTT_PASS="$(bashio::services mqtt 'password')"
else
  bashio::exit.nok "No MQTT broker — set the mqtt_host option or install the Mosquitto broker add-on."
fi

bashio::log.info "Starting sms2mqtt on ${SMS2MQTT_SERIAL_DEVICE} -> mqtt ${SMS2MQTT_MQTT_HOST}:${SMS2MQTT_MQTT_PORT}"
exec /usr/bin/sms2mqtt
