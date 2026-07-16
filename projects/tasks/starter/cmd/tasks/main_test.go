package main

import "testing"

func TestStarterThinClientCommand(t *testing.T) {
	if exit := run([]string{"--client", "nethttp", "list"}); exit != 1 {
		t.Fatalf("exit = %d", exit)
	}
	if exit := run([]string{"show", "0"}); exit != 2 {
		t.Fatalf("usage exit = %d", exit)
	}
}
