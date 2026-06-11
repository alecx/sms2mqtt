// Package sms decodes received SMS from GSM 03.40 PDUs (SMS-DELIVER): sender,
// timestamp, text (GSM-7 / UCS2), and concatenation headers for multipart
// reassembly. It is independent of the modem transport.
package sms

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"
	"unicode/utf16"
)

// Concat holds the concatenated-SMS reference from a message's UDH, present
// only on multipart messages.
type Concat struct {
	Ref   int // reference shared by all parts
	Total int // number of parts
	Seq   int // 1-based index of this part
}

// Message is a decoded SMS-DELIVER.
type Message struct {
	Sender    string
	Text      string
	Timestamp time.Time
	Concat    *Concat // nil unless multipart
}

// Decode parses a hex-encoded SMS-DELIVER PDU.
func Decode(pduHex string) (Message, error) {
	raw, err := hex.DecodeString(strings.TrimSpace(pduHex))
	if err != nil {
		return Message{}, fmt.Errorf("pdu: bad hex: %w", err)
	}
	r := &reader{buf: raw}

	smscLen := r.u8()
	r.skip(int(smscLen)) // SMSC info — not needed

	firstOctet := r.u8()
	udhi := firstOctet&0x40 != 0

	addrDigits := int(r.u8())
	addrType := r.u8()
	addr := r.bytes((addrDigits + 1) / 2)
	sender := decodeAddress(addrType, addrDigits, addr)

	_ = r.u8() // TP-PID
	dcs := r.u8()
	ts := decodeSCTS(r.bytes(7))
	udl := int(r.u8())
	ud := r.rest()
	if r.err != nil {
		return Message{}, fmt.Errorf("pdu: truncated: %w", r.err)
	}

	var concat *Concat
	text := ""
	switch alphabet(dcs) {
	case alphabetGSM7:
		text, concat = decodeGSM7UD(ud, udl, udhi)
	case alphabetUCS2:
		text, concat = decodeUCS2UD(ud, udhi)
	default:
		return Message{}, fmt.Errorf("pdu: unsupported DCS 0x%02X", dcs)
	}

	return Message{Sender: sender, Text: text, Timestamp: ts, Concat: concat}, nil
}

type alphabetKind int

const (
	alphabetGSM7 alphabetKind = iota
	alphabetUCS2
	alphabet8bit
)

// alphabet extracts the character set from the TP-DCS octet.
func alphabet(dcs byte) alphabetKind {
	switch (dcs >> 2) & 0x03 {
	case 0x01:
		return alphabet8bit
	case 0x02:
		return alphabetUCS2
	default:
		return alphabetGSM7
	}
}

// decodeGSM7UD decodes a GSM-7 user-data field, splitting off the UDH (and its
// 7-bit fill alignment) when present.
func decodeGSM7UD(ud []byte, septets int, udhi bool) (string, *Concat) {
	fill := 0
	var concat *Concat
	if udhi && len(ud) > 0 {
		udhLen := int(ud[0]) // octets following this length byte
		concat = parseConcatUDH(ud[1 : 1+udhLen])
		udhTotal := udhLen + 1 // including the length byte
		// GSM-7 text starts on the next septet boundary after the UDH.
		fill = (7 - (udhTotal*8)%7) % 7
		consumedSeptets := (udhTotal*8 + fill) / 7
		septets -= consumedSeptets
		ud = ud[udhTotal:]
	}
	return unpackGSM7(ud, septets, fill), concat
}

// decodeUCS2UD decodes a UCS2 (UTF-16BE) user-data field, splitting off the UDH
// when present. UCS2 data is byte-aligned, so there are no GSM-7 fill bits.
func decodeUCS2UD(ud []byte, udhi bool) (string, *Concat) {
	var concat *Concat
	if udhi && len(ud) > 0 {
		udhLen := int(ud[0])
		if 1+udhLen <= len(ud) {
			concat = parseConcatUDH(ud[1 : 1+udhLen])
			ud = ud[1+udhLen:]
		}
	}
	u16 := make([]uint16, 0, len(ud)/2)
	for i := 0; i+1 < len(ud); i += 2 {
		u16 = append(u16, uint16(ud[i])<<8|uint16(ud[i+1]))
	}
	return string(utf16.Decode(u16)), concat
}

// decodeAddress decodes a TP address field. International numbers get a leading
// '+'; alphanumeric addresses (type 0xD0) are GSM-7 packed.
func decodeAddress(addrType byte, nDigits int, b []byte) string {
	if addrType == 0xD0 {
		return unpackGSM7(b, nDigits*4/7, 0)
	}
	digits := bcdDigits(b)
	if len(digits) > nDigits {
		digits = digits[:nDigits]
	}
	if addrType == 0x91 {
		return "+" + digits
	}
	return digits
}

func parseConcatUDH(udh []byte) *Concat {
	for i := 0; i+1 < len(udh); {
		iei := udh[i]
		l := int(udh[i+1])
		v := udh[i+2 : i+2+l]
		switch {
		case iei == 0x00 && len(v) == 3: // concatenated, 8-bit ref
			return &Concat{Ref: int(v[0]), Total: int(v[1]), Seq: int(v[2])}
		case iei == 0x08 && len(v) == 4: // concatenated, 16-bit ref
			return &Concat{Ref: int(v[0])<<8 | int(v[1]), Total: int(v[2]), Seq: int(v[3])}
		}
		i += 2 + l
	}
	return nil
}
