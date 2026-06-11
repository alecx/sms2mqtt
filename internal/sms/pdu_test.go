package sms

import (
	"strings"
	"testing"
)

// The canonical GSM 03.40 SMS-DELIVER example (Wikipedia "GSM 03.40"):
//
//	SMSC +31 26 04 0000, sender +31 6 41 60 08 96, GSM-7, body "How are you?"
const wikipediaDeliver = "07911326040000F0040B911346610098F60000208062917314080CC8F71D14969741F977FD07"

func TestDecode_SinglePartGSM7(t *testing.T) {
	msg, err := Decode(wikipediaDeliver)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}
	if msg.Sender != "+31641600896" {
		t.Errorf("Sender = %q, want +31641600896", msg.Sender)
	}
	if msg.Text != "How are you?" {
		t.Errorf("Text = %q, want %q", msg.Text, "How are you?")
	}
	if msg.Concat != nil {
		t.Errorf("Concat = %+v, want nil for a single-part message", msg.Concat)
	}
}

// The following are real PDUs captured from the TRM240 (promotional Vodafone
// messages), exercising the decoder against production data.

// GSM-7 with an alphanumeric sender and an 8-bit concatenation UDH.
const realGSM7Concat = "07910427025014F96012D0D637396C7EBBCBB42A000062109011142180A00500030D0201A8EFB6F8CD0E83A675395C9ED697D96550BA2C77A7D3207618340CCBE9657618647D93C3E6B7BB0C9A974169F7185D4E9741F032285603A5C3EE7A589E2E8740D272DA3D0ECBC761D0B80E1A97D920789D9E76836AA0725DFE068DE565729A0E9AA741EE771A94A6A741E1F15B4E0EB741E330BD0C7A83E661F73C0C628741653C5D1E3E97E5"

func TestDecode_AlphanumericSenderAndConcat(t *testing.T) {
	msg, err := Decode(realGSM7Concat)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}
	if msg.Sender != "Vodafone4U" {
		t.Errorf("Sender = %q, want Vodafone4U", msg.Sender)
	}
	if msg.Concat == nil || *msg.Concat != (Concat{Ref: 13, Total: 2, Seq: 1}) {
		t.Errorf("Concat = %+v, want {Ref:13 Total:2 Seq:1}", msg.Concat)
	}
	if !strings.Contains(msg.Text, "Tombola Surprizele") {
		t.Errorf("Text = %q, want it to contain %q", msg.Text, "Tombola Surprizele")
	}
}

// GSM-7 using the escape extension for the euro sign.
const realGSM7Euro = "07910427025014F96012D0D637396C7EBBCBB42A000062401041945321A0050003110201ACE171D84D0F83E86150DA3D2EC3CBA0711DF406C9CB69F7382C1F87E565103B0C1A86E5F4323B0CB2BEC961F3DB5D0E8182E37ABBCE2E87F56190BC9C768FC3F271589E0691CBA076DAED02C5649B32682C2F93D37450DA4D9797413150182E6FC960A076380D9AA741F0373D0D1A87E7F4F4390C7A83ECE171D84D0F83D861503B2C2F83D2"

func TestDecode_GSM7EuroExtension(t *testing.T) {
	msg, err := Decode(realGSM7Euro)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}
	if !strings.Contains(msg.Text, "12€ credit") {
		t.Errorf("Text = %q, want it to contain the euro sign %q", msg.Text, "12€ credit")
	}
}

// UCS2 (UTF-16BE) with Romanian diacritics and a concatenation UDH.
const realUCS2 = "07910427025014F9600ED0D637396C7EBBCB0008623001210364808C05000302060100CE021B00690020006D0075006C021B0075006D0069006D0020006301030020006502190074006900200063006C00690065006E007400200056006F006400610066006F006E00650021002000500065006E0074007200750020006301030020007600720065006D0020007301030020006E0065002000EE006D00620075006E010300740103"

func TestDecode_UCS2Multipart(t *testing.T) {
	msg, err := Decode(realUCS2)
	if err != nil {
		t.Fatalf("Decode error: %v", err)
	}
	if msg.Sender != "Vodafone" {
		t.Errorf("Sender = %q, want Vodafone", msg.Sender)
	}
	if msg.Concat == nil || *msg.Concat != (Concat{Ref: 2, Total: 6, Seq: 1}) {
		t.Errorf("Concat = %+v, want {Ref:2 Total:6 Seq:1}", msg.Concat)
	}
	if !strings.Contains(msg.Text, "Îți mulțumim că ești") {
		t.Errorf("Text = %q, want Romanian UCS2 text", msg.Text)
	}
}
