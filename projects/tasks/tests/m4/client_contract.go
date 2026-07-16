package m4

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mbrndiar/learning-go/projects/tasks/solution/client"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/task"
)

type ClientFactory func(client.Config) (client.Transport, error)

func AssertClientContract(t *testing.T, factory ClientFactory) {
	t.Helper()
	t.Run("request construction and reusable transport", func(t *testing.T) {
		var requests atomic.Int32
		var path, contentType, body string
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			requests.Add(1)
			path = request.URL.RequestURI()
			contentType = request.Header.Get("Content-Type")
			content, _ := io.ReadAll(request.Body)
			body = string(content)
			writer.Header().Set("Content-Type", "application/json; charset=utf-8")
			switch requests.Load() {
			case 1:
				writer.WriteHeader(http.StatusCreated)
				_, _ = writer.Write([]byte(`{"id":1,"title":"Learn clients 🐍","completed":false}`))
			default:
				_, _ = writer.Write([]byte(`[{"id":1,"title":"Learn clients 🐍","completed":false}]`))
			}
		}))
		defer server.Close()
		transport := newTransport(t, factory, server.URL+"/api", time.Second)
		defer closeTransport(t, transport)
		created, err := transport.Create(context.Background(), task.CreateInput{Title: "Learn clients 🐍"})
		if err != nil || created.ID != 1 {
			t.Fatalf("Create = %#v, %v", created, err)
		}
		if path != "/api/tasks" || contentType != "application/json; charset=utf-8" ||
			body != `{"title":"Learn clients 🐍"}` {
			t.Fatalf("request = path %q type %q body %q", path, contentType, body)
		}
		completed := false
		values, err := transport.List(context.Background(), task.ListFilter{Completed: &completed})
		if err != nil || len(values) != 1 || requests.Load() != 2 ||
			path != "/api/tasks?completed=false" || contentType != "" || body != "" {
			t.Fatalf("List = %#v, %v; request=%q %q %q count=%d",
				values, err, path, contentType, body, requests.Load())
		}
	})

	t.Run("strict response validation", func(t *testing.T) {
		cases := []struct {
			name, contentType, body string
			status                  int
		}{
			{"missing content type", "", `{}`, 200},
			{"wrong content type", "text/plain", `{}`, 200},
			{"wrong charset", "application/json; charset=iso-8859-1", validTask, 200},
			{"invalid UTF-8", "application/json", string([]byte{0xff}), 200},
			{"malformed JSON", "application/json", `{`, 200},
			{"duplicate field", "application/json", `{"id":7,"id":8,"title":"x","completed":false}`, 200},
			{"trailing value", "application/json", validTask + `{}`, 200},
			{"unknown field", "application/json", `{"id":7,"title":"x","completed":false,"extra":1}`, 200},
			{"missing field", "application/json", `{"id":7,"title":"x"}`, 200},
			{"null field", "application/json", `{"id":7,"title":null,"completed":false}`, 200},
			{"wrong id type", "application/json", `{"id":true,"title":"x","completed":false}`, 200},
			{"wrong completed type", "application/json", `{"id":7,"title":"x","completed":0}`, 200},
			{"invalid id", "application/json", `{"id":0,"title":"x","completed":false}`, 200},
			{"unnormalized title", "application/json", `{"id":7,"title":" padded ","completed":false}`, 200},
			{"wrong status", "application/json", validTask, 201},
		}
		for _, test := range cases {
			t.Run(test.name, func(t *testing.T) {
				server := responseServer(test.status, test.contentType, test.body)
				defer server.Close()
				transport := newTransport(t, factory, server.URL, time.Second)
				defer closeTransport(t, transport)
				if _, err := transport.Get(context.Background(), 7); !errors.Is(err, client.ErrUnexpectedResponse) {
					t.Fatalf("error = %v, want unexpected response", err)
				}
			})
		}
	})

	t.Run("API error validation", func(t *testing.T) {
		cases := []struct {
			name, body string
			wantAPI    bool
		}{
			{"valid", `{"error":{"code":"not_found","message":"task 7 was not found"}}`, true},
			{"wrong envelope", `{"unexpected":{}}`, false},
			{"error not object", `{"error":"x"}`, false},
			{"missing code", `{"error":{"message":"x"}}`, false},
			{"unknown field", `{"error":{"code":"not_found","message":"x","extra":true}}`, false},
			{"wrong code", `{"error":{"code":"validation_error","message":"x"}}`, false},
			{"empty message", `{"error":{"code":"not_found","message":""}}`, false},
			{"wrong message type", `{"error":{"code":"not_found","message":7}}`, false},
			{"details array", `{"error":{"code":"not_found","message":"x","details":[]}}`, false},
			{"details null", `{"error":{"code":"not_found","message":"x","details":null}}`, false},
		}
		for _, test := range cases {
			t.Run(test.name, func(t *testing.T) {
				server := responseServer(404, "application/json", test.body)
				defer server.Close()
				transport := newTransport(t, factory, server.URL, time.Second)
				defer closeTransport(t, transport)
				_, err := transport.Get(context.Background(), 7)
				if test.wantAPI && !errors.Is(err, client.ErrAPI) {
					t.Fatalf("error = %v, want API error", err)
				}
				if !test.wantAPI && !errors.Is(err, client.ErrUnexpectedResponse) {
					t.Fatalf("error = %v, want unexpected response", err)
				}
			})
		}
	})

	t.Run("list order shape and exact 204", func(t *testing.T) {
		listBodies := []string{
			validTask,
			`[{"id":2,"title":"second","completed":false},{"id":1,"title":"first","completed":false}]`,
			`[{"id":1,"title":"x","completed":false},{"id":1,"title":"x","completed":false}]`,
			`null`,
		}
		for _, body := range listBodies {
			server := responseServer(200, "application/json", body)
			transport := newTransport(t, factory, server.URL, time.Second)
			_, err := transport.List(context.Background(), task.ListFilter{})
			closeTransport(t, transport)
			server.Close()
			if !errors.Is(err, client.ErrUnexpectedResponse) {
				t.Fatalf("List body %q error = %v", body, err)
			}
		}
		for _, response := range []struct{ contentType, body string }{
			{"application/json", ""},
		} {
			server := responseServer(204, response.contentType, response.body)
			transport := newTransport(t, factory, server.URL, time.Second)
			err := transport.Delete(context.Background(), 1)
			closeTransport(t, transport)
			server.Close()
			if !errors.Is(err, client.ErrUnexpectedResponse) {
				t.Fatalf("Delete response %#v error = %v", response, err)
			}
		}
	})

	t.Run("response limit and retries disabled", func(t *testing.T) {
		server := responseServer(200, "application/json", strings.Repeat(" ", (1<<20)+1))
		transport := newTransport(t, factory, server.URL, time.Second)
		_, err := transport.List(context.Background(), task.ListFilter{})
		closeTransport(t, transport)
		server.Close()
		if !errors.Is(err, client.ErrUnexpectedResponse) {
			t.Fatalf("oversized response error = %v", err)
		}

		var requests atomic.Int32
		server = httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
			requests.Add(1)
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(500)
			_, _ = writer.Write([]byte(`{"error":{"code":"internal_error","message":"failure"}}`))
		}))
		transport = newTransport(t, factory, server.URL, time.Second)
		_, err = transport.List(context.Background(), task.ListFilter{})
		closeTransport(t, transport)
		server.Close()
		if !errors.Is(err, client.ErrAPI) || requests.Load() != 1 {
			t.Fatalf("error = %v, request count = %d", err, requests.Load())
		}
	})

	t.Run("connection context and finite timeout", func(t *testing.T) {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		address := listener.Addr().String()
		_ = listener.Close()
		transport := newTransport(t, factory, "http://"+address, 100*time.Millisecond)
		_, err = transport.List(context.Background(), task.ListFilter{})
		closeTransport(t, transport)
		if !errors.Is(err, client.ErrConnection) {
			t.Fatalf("connection error = %v", err)
		}

		started := make(chan struct{})
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			close(started)
			<-request.Context().Done()
			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write([]byte(`[]`))
		}))
		transport = newTransport(t, factory, server.URL, 20*time.Millisecond)
		_, err = transport.List(context.Background(), task.ListFilter{})
		<-started
		closeTransport(t, transport)
		server.Close()
		if !errors.Is(err, client.ErrConnection) || !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("timeout error = %v", err)
		}

		server = httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			<-request.Context().Done()
		}))
		transport = newTransport(t, factory, server.URL, time.Second)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err = transport.List(ctx, task.ListFilter{})
		closeTransport(t, transport)
		server.Close()
		if !errors.Is(err, client.ErrConnection) {
			t.Fatalf("context cancellation error = %v", err)
		}
	})
}

const validTask = `{"id":7,"title":"Learn REST","completed":false}`

func newTransport(t *testing.T, factory ClientFactory, baseURL string, timeout time.Duration) client.Transport {
	t.Helper()
	transport, err := factory(client.Config{BaseURL: baseURL, Timeout: timeout})
	if err != nil {
		t.Fatal(err)
	}
	return transport
}

func closeTransport(t *testing.T, transport client.Transport) {
	t.Helper()
	if closer, ok := transport.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func responseServer(status int, contentType, body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		if contentType != "" {
			writer.Header().Set("Content-Type", contentType)
		}
		writer.WriteHeader(status)
		_, _ = writer.Write([]byte(body))
	}))
}
