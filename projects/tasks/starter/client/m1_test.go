package client_test

import (
	"errors"
	"testing"
	"time"

	"github.com/mbrndiar/learning-go/projects/tasks/starter/client"
	"github.com/mbrndiar/learning-go/projects/tasks/starter/task"
	"github.com/mbrndiar/learning-go/projects/tasks/tests/m1"
)

func TestMilestone1StarterClientIsExplicitlyIncomplete(t *testing.T) {
	m1.AssertStarterExplicit(t, task.ErrNotImplemented, func() int {
		return 0
	}, []func() error{
		func() error {
			_, err := client.NormalizeBaseURL("http://example.test")
			return err
		},
		func() error {
			_, err := (client.Config{
				BaseURL: "http://example.test",
				Timeout: time.Second,
			}).Validate()
			return err
		},
	})

	if !errors.Is((&client.APIError{}), client.ErrAPI) {
		t.Fatal("APIError classification is unavailable in the starter")
	}
}
