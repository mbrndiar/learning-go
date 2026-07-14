package httptestlesson

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestHandleGreetRecorder tests the handler directly against an
// httptest.ResponseRecorder. This is the fastest style: no real network
// listener is created, so it is well suited to exhaustive table-driven
// cases.
func TestHandleGreetRecorder(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		wantStatus int
		wantBody   string
	}{
		{name: "with name", query: "name=Ada", wantStatus: http.StatusOK, wantBody: "Hello, Ada!"},
		{name: "missing name", query: "", wantStatus: http.StatusBadRequest},
	}

	mux := NewMux()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/greet?"+test.query, nil)
			recorder := httptest.NewRecorder()

			mux.ServeHTTP(recorder, request)

			if recorder.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", recorder.Code, test.wantStatus, recorder.Body.String())
			}
			if test.wantBody == "" {
				return
			}

			var body greetResponse
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode response body: %v", err)
			}
			if body.Message != test.wantBody {
				t.Errorf("message = %q, want %q", body.Message, test.wantBody)
			}
		})
	}
}

// TestHandleGreetLiveServer spins up a real listening server with
// httptest.NewServer and exercises it through a normal http.Client. This
// style is closer to how a real client would talk to the API, at the cost
// of an actual (loopback) network round trip.
func TestHandleGreetLiveServer(t *testing.T) {
	server := httptest.NewServer(NewMux())
	// t.Cleanup ensures the listener is closed even if a later check fails.
	t.Cleanup(server.Close)

	resp, err := server.Client().Get(server.URL + "/greet?name=Grace")
	if err != nil {
		t.Fatalf("GET /greet: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var body greetResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode response body: %v", err)
	}
	if want := "Hello, Grace!"; body.Message != want {
		t.Errorf("message = %q, want %q", body.Message, want)
	}
}
