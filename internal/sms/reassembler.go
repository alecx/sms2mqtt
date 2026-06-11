package sms

import "strings"

// Reassembler joins multipart (concatenated) SMS into a single Message. Parts
// may arrive out of order. Single-part messages pass straight through. Keyed on
// sender+ref so interleaved messages and ref reuse across senders stay separate.
//
// The zero value is ready to use. Not safe for concurrent use; the SMS reader
// drives it from one goroutine.
type Reassembler struct {
	pending map[string]*partial
}

type partial struct {
	total  int
	sender string
	parts  map[int]Message // seq (1-based) -> part
}

// Add ingests one decoded part. It returns the completed Message and true when
// the part finishes a set (or is single-part); otherwise it buffers and returns
// false.
func (r *Reassembler) Add(m Message) (Message, bool) {
	if m.Concat == nil {
		return m, true
	}
	if r.pending == nil {
		r.pending = make(map[string]*partial)
	}

	key := m.Sender + "\x00" + itoa(m.Concat.Ref)
	p := r.pending[key]
	if p == nil {
		p = &partial{total: m.Concat.Total, sender: m.Sender, parts: make(map[int]Message)}
		r.pending[key] = p
	}
	p.parts[m.Concat.Seq] = m

	if len(p.parts) < p.total {
		return Message{}, false
	}

	var b strings.Builder
	var first Message
	for seq := 1; seq <= p.total; seq++ {
		part := p.parts[seq]
		if seq == 1 {
			first = part
		}
		b.WriteString(part.Text)
	}
	delete(r.pending, key)

	return Message{
		Sender:    p.sender,
		Text:      b.String(),
		Timestamp: first.Timestamp,
	}, true
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
