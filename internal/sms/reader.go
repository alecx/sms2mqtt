package sms

import "io"

// reader is a small cursor over a PDU byte slice. The first out-of-bounds access
// latches err and subsequent reads return zero values, so Decode can check once.
type reader struct {
	buf []byte
	pos int
	err error
}

func (r *reader) u8() byte {
	if r.err != nil {
		return 0
	}
	if r.pos >= len(r.buf) {
		r.err = io.ErrUnexpectedEOF
		return 0
	}
	b := r.buf[r.pos]
	r.pos++
	return b
}

func (r *reader) bytes(n int) []byte {
	if r.err != nil {
		return nil
	}
	if n < 0 || r.pos+n > len(r.buf) {
		r.err = io.ErrUnexpectedEOF
		return nil
	}
	b := r.buf[r.pos : r.pos+n]
	r.pos += n
	return b
}

func (r *reader) skip(n int) { r.bytes(n) }

func (r *reader) rest() []byte {
	if r.err != nil {
		return nil
	}
	b := r.buf[r.pos:]
	r.pos = len(r.buf)
	return b
}
