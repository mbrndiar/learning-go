package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateAndGetItem(t *testing.T) {
	t.Parallel()

	mux := newMux(newStore())

	createReq := httptest.NewRequest(http.MethodPost, "/items", strings.NewReader(`{"name":"widget"}`))
	createRec := httptest.NewRecorder()
	mux.ServeHTTP(createRec, createReq)

	if got, want := createRec.Code, http.StatusCreated; got != want {
		t.Fatalf("POST /items status = %d, want %d", got, want)
	}
	if !strings.Contains(createRec.Body.String(), `"name":"widget"`) {
		t.Fatalf("POST /items body = %q, want it to contain the created name", createRec.Body.String())
	}

	getReq := httptest.NewRequest(http.MethodGet, "/items/1", nil)
	getRec := httptest.NewRecorder()
	mux.ServeHTTP(getRec, getReq)

	if got, want := getRec.Code, http.StatusOK; got != want {
		t.Fatalf("GET /items/1 status = %d, want %d", got, want)
	}
	if !strings.Contains(getRec.Body.String(), `"widget"`) {
		t.Fatalf("GET /items/1 body = %q, want it to contain widget", getRec.Body.String())
	}
}

func TestGetMissingItemReturnsNotFound(t *testing.T) {
	t.Parallel()

	mux := newMux(newStore())

	req := httptest.NewRequest(http.MethodGet, "/items/99", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if got, want := rec.Code, http.StatusNotFound; got != want {
		t.Fatalf("GET /items/99 status = %d, want %d", got, want)
	}
}

func TestUpdateAndDeleteItem(t *testing.T) {
	t.Parallel()

	s := newStore()
	s.add("original")
	mux := newMux(s)

	putReq := httptest.NewRequest(http.MethodPut, "/items/1", strings.NewReader(`{"name":"renamed"}`))
	putRec := httptest.NewRecorder()
	mux.ServeHTTP(putRec, putReq)

	if got, want := putRec.Code, http.StatusOK; got != want {
		t.Fatalf("PUT /items/1 status = %d, want %d", got, want)
	}
	if !strings.Contains(putRec.Body.String(), "renamed") {
		t.Fatalf("PUT /items/1 body = %q, want it to contain renamed", putRec.Body.String())
	}

	delReq := httptest.NewRequest(http.MethodDelete, "/items/1", nil)
	delRec := httptest.NewRecorder()
	mux.ServeHTTP(delRec, delReq)

	if got, want := delRec.Code, http.StatusNoContent; got != want {
		t.Fatalf("DELETE /items/1 status = %d, want %d", got, want)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/items/1", nil)
	getRec := httptest.NewRecorder()
	mux.ServeHTTP(getRec, getReq)

	if got, want := getRec.Code, http.StatusNotFound; got != want {
		t.Fatalf("GET /items/1 after delete status = %d, want %d", got, want)
	}
}

func TestMethodNotAllowedOnUnmatchedVerb(t *testing.T) {
	t.Parallel()

	mux := newMux(newStore())

	// PATCH has no registered handler for /items, so net/http's mux falls
	// back to 405 because the path matches a different method's pattern.
	req := httptest.NewRequest(http.MethodPatch, "/items", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if got, want := rec.Code, http.StatusMethodNotAllowed; got != want {
		t.Fatalf("PATCH /items status = %d, want %d", got, want)
	}
}

func TestPathIDRejectsNonNumeric(t *testing.T) {
	t.Parallel()

	mux := newMux(newStore())

	req := httptest.NewRequest(http.MethodGet, "/items/not-a-number", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if got, want := rec.Code, http.StatusBadRequest; got != want {
		t.Fatalf("GET /items/not-a-number status = %d, want %d", got, want)
	}
}
