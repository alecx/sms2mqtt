package sms

// gsm7Basic is the GSM 03.38 basic character set (index 0–127 → Unicode rune).
// 0x1B is the escape to the extension table; it is handled in gsm7Rune.
var gsm7Basic = []rune(
	"@£$¥èéùìòÇ\nØø\rÅå" +
		"Δ_ΦΓΛΩΠΨΣΘΞÆæßÉ" +
		" !\"#¤%&'()*+,-./" +
		"0123456789:;<=>?" +
		"¡ABCDEFGHIJKLMNO" +
		"PQRSTUVWXYZÄÖÑÜ§" +
		"¿abcdefghijklmno" +
		"pqrstuvwxyzäöñüà")

// gsm7Ext maps an extension-table septet (the one following 0x1B) to its rune.
var gsm7Ext = map[byte]rune{
	0x0A: '\n', 0x14: '^', 0x28: '{', 0x29: '}', 0x2F: '\\',
	0x3C: '[', 0x3D: '~', 0x3E: ']', 0x40: '|', 0x65: '€',
}

func gsm7Rune(v byte) rune {
	if int(v) < len(gsm7Basic) {
		return gsm7Basic[v]
	}
	return '�'
}

// unpackGSM7 unpacks septetCount GSM-7 septets from data (LSB-first), discarding
// fill leading bits (UDH alignment), and maps them through the GSM 03.38
// alphabet including the 0x1B escape extension.
func unpackGSM7(data []byte, septetCount, fill int) string {
	septets := make([]byte, 0, septetCount)
	var acc uint32
	var nbits uint
	i := 0
	if fill > 0 && len(data) > 0 {
		acc = uint32(data[0]) >> uint(fill)
		nbits = uint(8 - fill)
		i = 1
	}
	for ; i < len(data); i++ {
		acc |= uint32(data[i]) << nbits
		nbits += 8
		for nbits >= 7 && len(septets) < septetCount {
			septets = append(septets, byte(acc&0x7f))
			acc >>= 7
			nbits -= 7
		}
	}

	out := make([]rune, 0, len(septets))
	for j := 0; j < len(septets); j++ {
		if septets[j] == 0x1B && j+1 < len(septets) {
			j++
			if r, ok := gsm7Ext[septets[j]]; ok {
				out = append(out, r)
			} else {
				out = append(out, gsm7Rune(septets[j]))
			}
			continue
		}
		out = append(out, gsm7Rune(septets[j]))
	}
	return string(out)
}
