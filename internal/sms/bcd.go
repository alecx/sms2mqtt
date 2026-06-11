package sms

import (
	"strings"
	"time"
)

// bcdDigits decodes nibble-swapped BCD telephone digits, dropping the 0xF
// half-octet filler used to pad odd-length numbers.
func bcdDigits(b []byte) string {
	var sb strings.Builder
	for _, o := range b {
		lo, hi := o&0x0f, o>>4
		if lo != 0x0f {
			sb.WriteByte('0' + lo)
		}
		if hi != 0x0f {
			sb.WriteByte('0' + hi)
		}
	}
	return sb.String()
}

// decodeSCTS decodes the 7-octet TP-Service-Centre-Time-Stamp (nibble-swapped
// BCD: YY MM DD hh mm ss TZ). TZ is in quarter-hours; bit 0x08 of its first
// swapped nibble is the sign.
func decodeSCTS(b []byte) time.Time {
	if len(b) < 7 {
		return time.Time{}
	}
	d := func(o byte) int { return int(o&0x0f)*10 + int(o>>4) }

	tzLo, tzHi := b[6]&0x0f, b[6]>>4
	quarters := int(tzLo&0x07)*10 + int(tzHi)
	offset := quarters * 15 * 60
	if tzLo&0x08 != 0 {
		offset = -offset
	}

	return time.Date(2000+d(b[0]), time.Month(d(b[1])), d(b[2]),
		d(b[3]), d(b[4]), d(b[5]), 0, time.FixedZone("", offset))
}
