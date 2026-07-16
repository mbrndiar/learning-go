package main

import "testing"

func TestSelectClient(t *testing.T) {
	selected, remaining, err := selectClient([]string{"--client", "nethttp", "--base-url", "http://example.test", "list"})
	if err != nil || selected != "nethttp" || len(remaining) != 3 || remaining[2] != "list" {
		t.Fatalf("selection = %q %v %v", selected, remaining, err)
	}
	if exit := run([]string{"--client", "resty", "list"}); exit != 2 {
		t.Fatalf("resty exit = %d", exit)
	}
}
