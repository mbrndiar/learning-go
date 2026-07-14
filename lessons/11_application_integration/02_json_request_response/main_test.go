package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateTaskHandlerSuccess(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(`{"title":"write tests"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	createTaskHandler(rec, req)

	if got, want := rec.Code, http.StatusCreated; got != want {
		t.Fatalf("status = %d, want %d (body: %s)", got, want, rec.Body.String())
	}
	if got, want := rec.Header().Get("Content-Type"), "application/json; charset=utf-8"; got != want {
		t.Fatalf("Content-Type = %q, want %q", got, want)
	}
	if !strings.Contains(rec.Body.String(), `"write tests"`) {
		t.Fatalf("body = %q, want it to contain the title", rec.Body.String())
	}
}

func TestCreateTaskHandlerRejectsWrongContentType(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(`{"title":"x"}`))
	req.Header.Set("Content-Type", "text/plain")
	rec := httptest.NewRecorder()

	createTaskHandler(rec, req)

	if got, want := rec.Code, http.StatusUnsupportedMediaType; got != want {
		t.Fatalf("status = %d, want %d", got, want)
	}
}

func TestCreateTaskHandlerRejectsUnknownFields(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(`{"title":"x","unexpected":true}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	createTaskHandler(rec, req)

	if got, want := rec.Code, http.StatusBadRequest; got != want {
		t.Fatalf("status = %d, want %d (body: %s)", got, want, rec.Body.String())
	}
}

func TestCreateTaskHandlerRejectsEmptyTitle(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(`{"title":"   "}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	createTaskHandler(rec, req)

	if got, want := rec.Code, http.StatusUnprocessableEntity; got != want {
		t.Fatalf("status = %d, want %d", got, want)
	}
}

func TestCreateTaskHandlerRejectsTrailingData(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(`{"title":"x"}{"title":"y"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	createTaskHandler(rec, req)

	if got, want := rec.Code, http.StatusBadRequest; got != want {
		t.Fatalf("status = %d, want %d (body: %s)", got, want, rec.Body.String())
	}
}

func TestDecodeStrictRejectsUnknownFields(t *testing.T) {
	t.Parallel()

	var dst createTaskRequest
	err := decodeStrict([]byte(`{"title":"x","extra":1}`), &dst)
	if err == nil {
		t.Fatal("decodeStrict() error = nil, want an error for unknown field")
	}
}

func TestDecodeStrictAcceptsValidBody(t *testing.T) {
	t.Parallel()

	var dst createTaskRequest
	if err := decodeStrict([]byte(`{"title":"x"}`), &dst); err != nil {
		t.Fatalf("decodeStrict() error = %v, want nil", err)
	}
	if dst.Title != "x" {
		t.Fatalf("dst.Title = %q, want %q", dst.Title, "x")
	}
}
