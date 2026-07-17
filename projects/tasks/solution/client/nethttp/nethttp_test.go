package nethttp_test

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/mbrndiar/learning-go/projects/tasks/solution/client"
	clientnethttp "github.com/mbrndiar/learning-go/projects/tasks/solution/client/nethttp"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/task"
	"github.com/mbrndiar/learning-go/projects/tasks/tests/m4"
)

func TestMilestone4ClientContract(t *testing.T) {
	m4.AssertClientContract(t, func(config client.Config) (client.Transport, error) {
		return clientnethttp.New(config)
	})
}

func TestBuildURLAndJSONRequest(t *testing.T) {
	built, err := clientnethttp.BuildURL("https://example.com/api", []string{"tasks", "a b", "雪"},
		url.Values{"state": {"not true"}, "tag": {"a/b&c"}})
	if err != nil {
		t.Fatal(err)
	}
	if built != "https://example.com/api/tasks/a%20b/%E9%9B%AA?state=not%20true&tag=a%2Fb%26c" {
		t.Fatalf("URL = %q", built)
	}

	var path, contentType, body string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		path, contentType = request.URL.RequestURI(), request.Header.Get("Content-Type")
		content, _ := io.ReadAll(request.Body)
		body = string(content)
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		writer.WriteHeader(201)
		_, _ = writer.Write([]byte(`{"id":1,"title":"Learn clients 🐍","completed":false}`))
	}))
	defer server.Close()
	transport, err := clientnethttp.New(client.Config{BaseURL: server.URL + "/api", Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	defer transport.Close()
	created, err := transport.Create(context.Background(), task.CreateInput{Title: "Learn clients 🐍"})
	if err != nil || created.ID != 1 {
		t.Fatalf("Create = %#v, %v", created, err)
	}
	if path != "/api/tasks" || !strings.HasPrefix(contentType, "application/json") ||
		body != `{"title":"Learn clients 🐍"}` {
		t.Fatalf("request = path %q type %q body %q", path, contentType, body)
	}
}

func TestResponseValidationAndAPIErrors(t *testing.T) {
	cases := []struct {
		name        string
		status      int
		contentType string
		body        string
		wantAPI     bool
	}{
		{"API error", 404, "application/json", `{"error":{"code":"not_found","message":"task 7 was not found"}}`, true},
		{"wrong content type", 200, "text/plain", `{}`, false},
		{"missing field", 200, "application/json", `{"id":7,"title":"x"}`, false},
		{"unknown field", 200, "application/json", `{"id":7,"title":"x","completed":false,"extra":1}`, false},
		{"duplicate field", 200, "application/json", `{"id":7,"id":8,"title":"x","completed":false}`, false},
		{"invalid UTF-8", 200, "application/json", string([]byte{0xff}), false},
		{"wrong status", 201, "application/json", `{"id":7,"title":"x","completed":false}`, false},
		{"wrong error code", 404, "application/json", `{"error":{"code":"validation_error","message":"wrong"}}`, false},
		{"error details null", 404, "application/json", `{"error":{"code":"not_found","message":"x","details":null}}`, false},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			roundTripper := roundTripFunc(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: test.status,
					Header:     http.Header{"Content-Type": {test.contentType}},
					Body:       io.NopCloser(strings.NewReader(test.body)),
				}, nil
			})
			transport, err := clientnethttp.NewWithHTTPClient(
				client.Config{BaseURL: "http://example.test", Timeout: time.Second},
				&http.Client{Transport: roundTripper, Timeout: time.Second},
			)
			if err != nil {
				t.Fatal(err)
			}
			_, err = transport.Get(context.Background(), 7)
			if test.wantAPI {
				if !errors.Is(err, client.ErrAPI) {
					t.Fatalf("error = %v, want API", err)
				}
			} else if !errors.Is(err, client.ErrUnexpectedResponse) {
				t.Fatalf("error = %v, want unexpected response", err)
			}
		})
	}
}

func TestListShapeOrderAndDelete204(t *testing.T) {
	responses := []*http.Response{
		response(200, "application/json", `{"id":7,"title":"x","completed":false}`),
		response(200, "application/json", `[{"id":2,"title":"second","completed":false},{"id":1,"title":"first","completed":false}]`),
		response(200, "application/json", `[{"id":1,"title":"x","completed":false},{"id":1,"title":"x","completed":false}]`),
		response(204, "", " "),
	}
	index := 0
	transport, err := clientnethttp.NewWithHTTPClient(
		client.Config{BaseURL: "http://example.test", Timeout: time.Second},
		&http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			value := responses[index]
			index++
			return value, nil
		}), Timeout: time.Second},
	)
	if err != nil {
		t.Fatal(err)
	}
	for range 3 {
		if _, err := transport.List(context.Background(), task.ListFilter{}); !errors.Is(err, client.ErrUnexpectedResponse) {
			t.Fatalf("List error = %v", err)
		}
	}
	if err := transport.Delete(context.Background(), 1); !errors.Is(err, client.ErrUnexpectedResponse) {
		t.Fatalf("Delete error = %v", err)
	}
}

func TestConnectionFailureAndFiniteTimeout(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	_ = listener.Close()
	transport, err := clientnethttp.New(client.Config{BaseURL: "http://" + address, Timeout: 100 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := transport.List(context.Background(), task.ListFilter{}); !errors.Is(err, client.ErrConnection) {
		t.Fatalf("connection error = %v", err)
	}

	started := make(chan struct{})
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		close(started)
		<-release
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`[]`))
	}))
	defer server.Close()
	slow, err := clientnethttp.New(client.Config{BaseURL: server.URL, Timeout: 20 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	_, err = slow.List(context.Background(), task.ListFilter{})
	<-started
	close(release)
	if !errors.Is(err, client.ErrConnection) || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("timeout error = %v", err)
	}
}

func TestInjectedClientRedirectPolicyIsOverridden(t *testing.T) {
	followed := false
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/api/tasks/1" {
			http.Redirect(writer, request, "/api/tasks/2", http.StatusFound)
			return
		}
		followed = true
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"id":2,"title":"redirected","completed":false}`))
	}))
	defer server.Close()

	// The zero-value CheckRedirect follows redirects automatically; the
	// caller-owned client below therefore represents a "follows redirects"
	// policy that NewWithHTTPClient must override rather than trust.
	original := &http.Client{Timeout: time.Second}
	transport, err := clientnethttp.NewWithHTTPClient(
		client.Config{BaseURL: server.URL + "/api", Timeout: time.Second}, original)
	if err != nil {
		t.Fatal(err)
	}
	defer transport.Close()

	_, err = transport.Get(context.Background(), 1)
	if followed {
		t.Fatal("client followed a redirect instead of rejecting it as an unexpected response")
	}
	if !errors.Is(err, client.ErrUnexpectedResponse) {
		t.Fatalf("error = %v, want unexpected response for an unfollowed redirect", err)
	}
	if original.CheckRedirect != nil {
		t.Fatal("original caller-owned client's CheckRedirect was mutated")
	}
	if original.Timeout != time.Second {
		t.Fatal("original caller-owned client's Timeout was mutated")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}

func response(status int, contentType, body string) *http.Response {
	header := http.Header{}
	if contentType != "" {
		header.Set("Content-Type", contentType)
	}
	return &http.Response{StatusCode: status, Header: header, Body: io.NopCloser(strings.NewReader(body))}
}
