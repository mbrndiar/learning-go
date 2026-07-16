package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRoutingAndStrictJSON(t *testing.T) {
	handler := (&API{}).Handler()
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/items", strings.NewReader(`{"name":"book"}`)))
	if rec.Code != http.StatusCreated || rec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("response = %d, %q", rec.Code, rec.Header().Get("Content-Type"))
	}
	var item Item
	if err := json.NewDecoder(rec.Body).Decode(&item); err != nil || item.Name != "book" {
		t.Fatalf("item = %+v, %v", item, err)
	}
	bad := httptest.NewRecorder()
	handler.ServeHTTP(bad, httptest.NewRequest(http.MethodPost, "/items", strings.NewReader(`{"name":"book","extra":true}`)))
	if bad.Code != http.StatusBadRequest {
		t.Fatalf("unknown field status = %d", bad.Code)
	}
	get := httptest.NewRecorder()
	handler.ServeHTTP(get, httptest.NewRequest(http.MethodGet, "/items/7", nil))
	if get.Code != http.StatusOK {
		t.Fatalf("GET status = %d", get.Code)
	}
}
