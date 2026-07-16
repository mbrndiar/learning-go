package resty_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	restylib "github.com/go-resty/resty/v2"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/client"
	clientresty "github.com/mbrndiar/learning-go/projects/tasks/solution/client/resty"
	"github.com/mbrndiar/learning-go/projects/tasks/tests/m4"
)

func TestBuildURL(t *testing.T) {
	built, err := clientresty.BuildURL("https://example.com/api", []string{"tasks", "a b", "雪"},
		url.Values{"state": {"not true"}, "tag": {"a/b&c"}})
	if err != nil {
		t.Fatal(err)
	}
	if built != "https://example.com/api/tasks/a%20b/%E9%9B%AA?state=not%20true&tag=a%2Fb%26c" {
		t.Fatalf("URL = %q", built)
	}
}

func TestMilestone4ClientContract(t *testing.T) {
	m4.AssertClientContract(t, func(config client.Config) (client.Transport, error) {
		return clientresty.New(config)
	})
}

func TestInjectedClientIsReusedWithFiniteTimeoutAndNoRetries(t *testing.T) {
	underlying := restylib.New().SetRetryCount(4)
	transport, err := clientresty.NewWithRestyClient(
		client.Config{BaseURL: "http://example.test", Timeout: 250 * time.Millisecond},
		underlying,
	)
	if err != nil {
		t.Fatal(err)
	}
	defer transport.Close()
	if underlying.GetClient().Timeout != 250*time.Millisecond {
		t.Fatalf("timeout = %v", underlying.GetClient().Timeout)
	}
	if underlying.RetryCount != 0 {
		t.Fatalf("retry count = %d", underlying.RetryCount)
	}
}

func TestDeleteRejectsNonempty204(t *testing.T) {
	httpClient := &http.Client{
		Timeout: time.Second,
		Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 204,
				Header:     http.Header{},
				Body:       io.NopCloser(strings.NewReader(" ")),
			}, nil
		}),
	}
	transport, err := clientresty.NewWithRestyClient(
		client.Config{BaseURL: "http://example.test", Timeout: time.Second},
		restylib.NewWithClient(httpClient),
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := transport.Delete(context.Background(), 1); !errors.Is(err, client.ErrUnexpectedResponse) {
		t.Fatalf("Delete error = %v", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}
