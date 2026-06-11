# SMS2MQTT (TRM240) — Home Assistant add-on

Bridges a Teltonika TRM240 (Quectel EC21) USB GSM modem to MQTT: publishes
received SMS to `sms2mqtt/received` (the `automation.incoming_sms` contract) and
GSM stats to `sms2mqtt/status`, with availability via MQTT LWT and a `/health`
watchdog.

## Install

The service binary is pre-built per arch (`bin/sms2mqtt.{aarch64,amd64}`).
Regenerate with `./build.sh` from the repo root after code changes.

### Option A — local add-on (fastest, for the first deploy)
1. On the Yellow, enable the **Samba** or **SSH** add-on.
2. Copy this `sms2mqtt/` folder into `/addons/` (so it's `/addons/sms2mqtt/`).
3. Settings → Add-ons → Add-on Store → ⋮ → **Check for updates**; the add-on
   appears under **Local add-ons**. Install it.

### Option B — GitHub custom repository (for ongoing updates)
1. Settings → Add-ons → Add-on Store → ⋮ → **Repositories** → add
   `https://github.com/alecx/sms2mqtt`.
2. Install **SMS2MQTT (TRM240)** from the new repository entry.

## Configuration

| Option | Default | Notes |
|---|---|---|
| `serial_device` | `/dev/serial/by-id/usb-Android_Android-if02-port0` | The EC21 **AT control port** is interface **if02**. Confirm with `ls /dev/serial/by-id/` on the host. |
| `baud` | `115200` | |
| `stats_interval` | `60s` | Doubles as the heartbeat; `/health` fails after 3× this with no publish. |
| `log_level` | `info` | |

MQTT host/credentials come from the Supervisor (`mqtt:need`) — install the
**Mosquitto broker** add-on first.

## ⚠️ ModemManager on Home Assistant OS

If the add-on logs `EBUSY` / cannot open the serial port, ModemManager on the
host has claimed the modem (same issue documented for the NUC). Verify with the
SSH add-on:

```sh
mmcli -L            # if it lists a Quectel/EC21 modem, MM has it
```

If so, add a host udev rule so MM ignores `2c7c:0121` (see the vault page
`wiki/troubleshooting/modemmanager-grabs-usb-modem.md`) and reload. On a stock
HA-OS install ModemManager may be absent — check before assuming.

## Verifying

- `Settings → Devices & Services → MQTT` → listen to `sms2mqtt/#`, or use the
  Mosquitto add-on logs.
- Expect `sms2mqtt/availability = online`, periodic `sms2mqtt/status`, and a
  `sms2mqtt/received` JSON when an SMS arrives.
