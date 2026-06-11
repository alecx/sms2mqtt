package sms

import "testing"

func TestReassembler_SinglePartPassesThrough(t *testing.T) {
	var r Reassembler
	msg := Message{Sender: "+100", Text: "hi"}

	got, done := r.Add(msg)
	if !done {
		t.Fatal("single-part message should complete immediately")
	}
	if got.Text != "hi" {
		t.Errorf("Text = %q, want hi", got.Text)
	}
}

func TestReassembler_JoinsMultipartInOrder(t *testing.T) {
	var r Reassembler
	ref := &Concat{Ref: 7, Total: 2, Seq: 2}
	ref1 := &Concat{Ref: 7, Total: 2, Seq: 1}

	// Parts arrive out of order: seq 2 first.
	if _, done := r.Add(Message{Sender: "Bank", Text: "world", Concat: ref}); done {
		t.Fatal("incomplete set should not complete on the first part")
	}
	got, done := r.Add(Message{Sender: "Bank", Text: "hello ", Concat: ref1})
	if !done {
		t.Fatal("set should complete when the last missing part arrives")
	}
	if got.Text != "hello world" {
		t.Errorf("Text = %q, want %q (joined in seq order)", got.Text, "hello world")
	}
	if got.Sender != "Bank" {
		t.Errorf("Sender = %q, want Bank", got.Sender)
	}
	if got.Concat != nil {
		t.Errorf("joined message should have Concat cleared, got %+v", got.Concat)
	}
}

func TestReassembler_DifferentRefsDoNotMix(t *testing.T) {
	var r Reassembler
	// Same sender, two different concat refs interleaved.
	r.Add(Message{Sender: "X", Text: "A1", Concat: &Concat{Ref: 1, Total: 2, Seq: 1}})
	r.Add(Message{Sender: "X", Text: "B1", Concat: &Concat{Ref: 2, Total: 2, Seq: 1}})
	got, done := r.Add(Message{Sender: "X", Text: "A2", Concat: &Concat{Ref: 1, Total: 2, Seq: 2}})
	if !done {
		t.Fatal("ref 1 should complete")
	}
	if got.Text != "A1A2" {
		t.Errorf("Text = %q, want A1A2 (ref 2 must not bleed in)", got.Text)
	}
}
