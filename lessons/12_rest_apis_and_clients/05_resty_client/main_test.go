package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRestyWorkflow(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer lesson-token" {
			t.Errorf("Authorization = %q", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type = %q", r.Header.Get("Content-Type"))
		}
		var input map[string]string
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input["name"] != "book" {
			t.Errorf("request body = %v, %v", input, err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(Widget{ID: 1, Name: "book"})
	}))
	t.Cleanup(server.Close)

	widget, err := NewAPIClient(server.URL, "lesson-token").CreateWidget(context.Background(), "book")
	if err != nil || widget != (Widget{ID: 1, Name: "book"}) {
		t.Fatalf("CreateWidget = %+v, %v", widget, err)
	}
}

func TestRestyRejectsUnexpectedStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
	}))
	t.Cleanup(server.Close)

	_, err := NewAPIClient(server.URL, "lesson-token").CreateWidget(context.Background(), "book")
	if err == nil {
		t.Fatal("CreateWidget returned nil error for status 503")
	}
}
