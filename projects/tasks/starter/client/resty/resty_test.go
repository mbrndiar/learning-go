package resty_test

import (
	"errors"
	"testing"
	"time"

	"github.com/mbrndiar/learning-go/projects/tasks/starter/client"
	"github.com/mbrndiar/learning-go/projects/tasks/starter/client/resty"
	"github.com/mbrndiar/learning-go/projects/tasks/starter/task"
)

func TestPlaceholderIsExplicit(t *testing.T) {
	value, err := resty.New(client.Config{BaseURL: "http://127.0.0.1:8000", Timeout: time.Second})
	if value != nil || !errors.Is(err, task.ErrNotImplemented) {
		t.Fatalf("New = %#v, %v", value, err)
	}
}
