// Package stats parses the GSM network-status AT responses (CSQ, COPS, CREG)
// into a Status suitable for publishing to MQTT / Home Assistant.
package stats

import (
	"strconv"
	"strings"
)

// Status is a snapshot of GSM network state for publishing.
type Status struct {
	Operator     string `json:"operator"`
	AccessTech   string `json:"access_tech"`
	SignalDBM    *int   `json:"signal_dbm"` // nil when unknown
	SignalPct    *int   `json:"signal_pct"`
	Registered   bool   `json:"registered"`
	Roaming      bool   `json:"roaming"`
	Registration string `json:"registration"`
}

// ParseCSQ parses "+CSQ: <rssi>,<ber>" into signal strength. rssi 0–31 maps
// linearly to −113…−51 dBm; 99 means "unknown" and returns ok=false.
func ParseCSQ(line string) (dbm, pct int, ok bool) {
	fields := splitResponse(line, "+CSQ:")
	if len(fields) < 1 {
		return 0, 0, false
	}
	rssi, err := strconv.Atoi(strings.TrimSpace(fields[0]))
	if err != nil || rssi < 0 || rssi > 31 {
		return 0, 0, false
	}
	return -113 + 2*rssi, rssi * 100 / 31, true
}

// ParseCOPS parses `+COPS: <mode>,<format>,"<oper>",<act>` into the operator
// name and a human access-technology label.
func ParseCOPS(line string) (operator, accessTech string) {
	fields := splitResponse(line, "+COPS:")
	if len(fields) >= 3 {
		operator = strings.Trim(strings.TrimSpace(fields[2]), `"`)
	}
	if len(fields) >= 4 {
		accessTech = accessTechName(strings.TrimSpace(fields[3]))
	}
	return operator, accessTech
}

// ParseCREG parses "+CREG: <n>,<stat>" registration state. stat 1 = registered
// home, 5 = registered roaming; everything else is not (usefully) registered.
func ParseCREG(line string) (registered, roaming bool) {
	fields := splitResponse(line, "+CREG:")
	if len(fields) < 2 {
		return false, false
	}
	stat, err := strconv.Atoi(strings.TrimSpace(fields[1]))
	if err != nil {
		return false, false
	}
	return stat == 1 || stat == 5, stat == 5
}

// RegistrationName gives a human label for a CREG stat code.
func RegistrationName(line string) string {
	fields := splitResponse(line, "+CREG:")
	if len(fields) < 2 {
		return "unknown"
	}
	stat, err := strconv.Atoi(strings.TrimSpace(fields[1]))
	if err != nil {
		return "unknown"
	}
	switch stat {
	case 0:
		return "not_registered"
	case 1:
		return "registered_home"
	case 2:
		return "searching"
	case 3:
		return "denied"
	case 5:
		return "registered_roaming"
	default:
		return "unknown"
	}
}

func accessTechName(act string) string {
	switch act {
	case "0", "1", "3":
		return "GSM"
	case "2", "4", "5", "6":
		return "UMTS"
	case "7":
		return "LTE"
	default:
		return act
	}
}

// splitResponse strips an optional "+CMD:" prefix and splits the CSV fields.
func splitResponse(line, prefix string) []string {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, prefix)
	return strings.Split(line, ",")
}
