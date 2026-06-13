package main

import "testing"

func TestInitMessage(t *testing.T) {
	expected := "Docker Swarm Utils: initialized"
	if InitMessage != expected {
		t.Errorf("expected %q, got %q", expected, InitMessage)
	}
}
