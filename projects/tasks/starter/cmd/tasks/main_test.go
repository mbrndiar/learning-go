package main

import "testing"

func TestStarterThinClientCommand(t *testing.T) {
	if exit := run([]string{"--client", "nethttp", "list"}); exit != 1 {
		t.Fatalf("exit = %d", exit)
	}
	if exit := run([]string{"--client", "resty", "list"}); exit != 1 {
		t.Fatalf("resty exit = %d", exit)
	}
	if exit := run([]string{"--client", "gin", "list"}); exit != 2 {
		t.Fatalf("unsupported exit = %d", exit)
	}
	if exit := run([]string{"show", "0"}); exit != 2 {
		t.Fatalf("usage exit = %d", exit)
	}
}
