package main

import "testing"

func TestSelectClient(t *testing.T) {
	selected, remaining, err := selectClient([]string{"--client", "nethttp", "--base-url", "http://example.test", "list"})
	if err != nil || selected != "nethttp" || len(remaining) != 3 || remaining[2] != "list" {
		t.Fatalf("selection = %q %v %v", selected, remaining, err)
	}
	selected, remaining, err = selectClient([]string{"--client=resty", "list"})
	if err != nil || selected != "resty" || len(remaining) != 1 || remaining[0] != "list" {
		t.Fatalf("resty selection = %q %v %v", selected, remaining, err)
	}
	if exit := run([]string{"--client", "gin", "list"}); exit != 2 {
		t.Fatalf("unsupported client exit = %d", exit)
	}
}
