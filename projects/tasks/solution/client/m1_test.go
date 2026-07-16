package client_test

import (
	"errors"
	"testing"
	"time"

	"github.com/mbrndiar/learning-go/projects/tasks/solution/client"
	"github.com/mbrndiar/learning-go/projects/tasks/tests/m1"
)

func TestMilestone1ClientBoundary(t *testing.T) {
	m1.AssertSolutionClient(t, m1.ClientHarness{
		DefaultTimeout:   client.DefaultTimeout,
		NormalizeBaseURL: client.NormalizeBaseURL,
		ValidateConfig: func(baseURL string, timeout time.Duration) (string, time.Duration, error) {
			config, err := (client.Config{BaseURL: baseURL, Timeout: timeout}).Validate()
			return config.BaseURL, config.Timeout, err
		},
		IsInvalidConfig: func(err error) bool {
			return errors.Is(err, client.ErrInvalidConfiguration)
		},
		ConfigDetails: func(err error) (string, string, bool) {
			var target *client.ConfigError
			if !errors.As(err, &target) {
				return "", "", false
			}
			return target.Field, target.Message, true
		},
		NewAPIError: func() error {
			return &client.APIError{Status: 404, Code: "not_found", Message: "task 1 was not found"}
		},
		IsAPI: func(err error) bool {
			return errors.Is(err, client.ErrAPI)
		},
		NewResponseError: func(err error) error {
			return &client.ResponseError{Status: 200, Message: "invalid task", Err: err}
		},
		IsUnexpectedResponse: func(err error) bool {
			return errors.Is(err, client.ErrUnexpectedResponse)
		},
		NewConnectionError: func(err error) error {
			return &client.ConnectionError{Err: err}
		},
		IsConnection: func(err error) bool {
			return errors.Is(err, client.ErrConnection)
		},
	})
}
