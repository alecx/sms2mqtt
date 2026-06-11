#!/usr/bin/with-contenv bashio
# Map add-on options + the Supervisor-provided MQTT service into the env the
# sms2mqtt binary reads. Secrets stay out of the command line.
set -e

export SMS2MQTT_SERIAL_DEVICE="$(bashio::config 'serial_device')"
export SMS2MQTT_BAUD="$(bashio::config 'baud')"
export SMS2MQTT_STATS_INTERVAL="$(bashio::config 'stats_interval')"
export SMS2MQTT_LOG_LEVEL="$(bashio::config 'log_level')"
export SMS2MQTT_HEALTH_ADDR=":8099"

if bashio::services.available "mqtt"; then
  export SMS2MQTT_MQTT_HOST="$(bashio::services mqtt 'host')"
  export SMS2MQTT_MQTT_PORT="$(bashio::services mqtt 'port')"
  export SMS2MQTT_MQTT_USER="$(bashio::services mqtt 'username')"
  export SMS2MQTT_MQTT_PASS="$(bashio::services mqtt 'password')"
else
  bashio::exit.nok "No MQTT service available — add the Mosquitto broker add-on."
fi

bashio::log.info "Starting sms2mqtt on ${SMS2MQTT_SERIAL_DEVICE} -> mqtt ${SMS2MQTT_MQTT_HOST}:${SMS2MQTT_MQTT_PORT}"
exec /usr/bin/sms2mqtt
