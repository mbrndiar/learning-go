package cli_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/mbrndiar/learning-go/projects/tasks/solution/cli"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/client"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/task"
)

type fakeTransport struct {
	operation string
	input     task.UpdateInput
	err       error
	closeErr  error
}

func (transport *fakeTransport) Close() error { return transport.closeErr }

func (transport *fakeTransport) Create(context.Context, task.CreateInput) (task.Task, error) {
	transport.operation = "add"
	return task.Task{ID: 7, Title: "Learn REST"}, transport.err
}
func (transport *fakeTransport) List(context.Context, task.ListFilter) ([]task.Task, error) {
	transport.operation = "list"
	return []task.Task{{ID: 7, Title: "Learn REST"}}, transport.err
}
func (transport *fakeTransport) Get(context.Context, int64) (task.Task, error) {
	transport.operation = "show"
	return task.Task{ID: 7, Title: "Learn REST"}, transport.err
}
func (transport *fakeTransport) Update(_ context.Context, _ int64, input task.UpdateInput) (task.Task, error) {
	transport.operation, transport.input = "update", input
	return task.Task{ID: 7, Title: "Learn REST", Completed: input.Completed != nil && *input.Completed}, transport.err
}
func (transport *fakeTransport) Delete(context.Context, int64) error {
	transport.operation = "remove"
	return transport.err
}

func TestCommandsOutputAndExplicitFalse(t *testing.T) {
	cases := []struct {
		args      []string
		operation string
		output    string
	}{
		{[]string{"add", "Learn REST"}, "add", `{"id":7,"title":"Learn REST","completed":false}` + "\n"},
		{[]string{"list", "--completed", "false"}, "list", `[{"id":7,"title":"Learn REST","completed":false}]` + "\n"},
		{[]string{"show", "7"}, "show", `{"id":7,"title":"Learn REST","completed":false}` + "\n"},
		{[]string{"update", "7", "--completed", "false"}, "update", `{"id":7,"title":"Learn REST","completed":false}` + "\n"},
		{[]string{"complete", "7"}, "update", `{"id":7,"title":"Learn REST","completed":true}` + "\n"},
		{[]string{"remove", "7"}, "remove", `{"deleted":7}` + "\n"},
	}
	for _, test := range cases {
		transport := &fakeTransport{}
		var stdout, stderr bytes.Buffer
		factory := func(config client.Config) (client.Transport, error) {
			if config.BaseURL != "http://127.0.0.1:8000" || config.Timeout != 5*time.Second {
				t.Fatalf("config = %#v", config)
			}
			return transport, nil
		}
		if exit := cli.Run(test.args, factory, &stdout, &stderr); exit != 0 ||
			stdout.String() != test.output || stderr.Len() != 0 || transport.operation != test.operation {
			t.Fatalf("%v => exit=%d stdout=%q stderr=%q operation=%q", test.args, exit, stdout.String(), stderr.String(), transport.operation)
		}
		if test.operation == "update" && test.args[0] == "update" &&
			(transport.input.Completed == nil || *transport.input.Completed) {
			t.Fatal("explicit false was not preserved")
		}
	}
}

func TestUsageAndHelpDoNotConstructTransport(t *testing.T) {
	for _, args := range [][]string{
		{"show", "0"}, {"update", "1"}, {"--timeout", "0", "list"},
		{"--timeout", "nan", "list"}, {"--base-url", "ftp://example.com", "list"},
	} {
		called := false
		var stdout, stderr bytes.Buffer
		exit := cli.Run(args, func(client.Config) (client.Transport, error) {
			called = true
			return &fakeTransport{}, nil
		}, &stdout, &stderr)
		if exit != 2 || called || stdout.Len() != 0 || !strings.HasPrefix(stderr.String(), "usage: ") ||
			strings.Count(stderr.String(), "\n") != 1 {
			t.Fatalf("%v => exit=%d called=%v stdout=%q stderr=%q", args, exit, called, stdout.String(), stderr.String())
		}
	}
	called := false
	var stdout, stderr bytes.Buffer
	exit := cli.Run([]string{"--help"}, func(client.Config) (client.Transport, error) {
		called = true
		return &fakeTransport{}, nil
	}, &stdout, &stderr)
	if exit != 0 || called || !strings.HasPrefix(stdout.String(), "usage: ") || stderr.Len() != 0 {
		t.Fatalf("help => exit=%d called=%v stdout=%q stderr=%q", exit, called, stdout.String(), stderr.String())
	}
}

func TestExitCategories(t *testing.T) {
	cases := []struct {
		err    error
		exit   int
		stderr string
	}{
		{&client.APIError{Status: 404, Code: "not_found", Message: "task 7 was not found"}, 3, "api: 404 not_found: task 7 was not found\n"},
		{&client.ResponseError{Status: 200, Message: "invalid task response"}, 4, "malformed-response: invalid task response\n"},
		{&client.ConnectionError{Err: context.DeadlineExceeded}, 5, "connection: timeout: deadline exceeded\n"},
		{&client.ConnectionError{Err: errors.New("connection refused")}, 5, "connection: connection refused\n"},
		{errors.New("private library detail"), 5, "transport: request failed\n"},
	}

	for _, test := range cases {
		var stdout, stderr bytes.Buffer
		exit := cli.Run([]string{"list"}, func(client.Config) (client.Transport, error) {
			return &fakeTransport{err: test.err}, nil
		}, &stdout, &stderr)
		if exit != test.exit || stdout.Len() != 0 || stderr.String() != test.stderr {
			t.Fatalf("%v => exit=%d stdout=%q stderr=%q", test.err, exit, stdout.String(), stderr.String())
		}
	}
}

func TestFactoryAndCleanupFailures(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exit := cli.Run([]string{"list"}, func(client.Config) (client.Transport, error) {
		return nil, errors.New("private constructor detail")
	}, &stdout, &stderr)
	if exit != 5 || stdout.Len() != 0 || stderr.String() != "transport: request failed\n" {
		t.Fatalf("factory => exit=%d stdout=%q stderr=%q", exit, stdout.String(), stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	exit = cli.Run([]string{"list"}, func(client.Config) (client.Transport, error) {
		return &fakeTransport{closeErr: errors.New("private close detail")}, nil
	}, &stdout, &stderr)
	if exit != 5 || stdout.Len() != 0 || stderr.String() != "transport: request failed\n" {
		t.Fatalf("cleanup => exit=%d stdout=%q stderr=%q", exit, stdout.String(), stderr.String())
	}
}

// TestOperationErrorTakesPrecedenceOverCleanupFailure proves that when both
// the command operation and the subsequent transport Close fail, the
// operation's own error category and message win: a cleanup failure must
// never mask an API, response-validation, or connection error.
func TestOperationErrorTakesPrecedenceOverCleanupFailure(t *testing.T) {
	cases := []struct {
		err    error
		exit   int
		stderr string
	}{
		{&client.APIError{Status: 404, Code: "not_found", Message: "task 7 was not found"}, 3, "api: 404 not_found: task 7 was not found\n"},
		{&client.ResponseError{Status: 200, Message: "invalid task response"}, 4, "malformed-response: invalid task response\n"},
		{&client.ConnectionError{Err: context.DeadlineExceeded}, 5, "connection: timeout: deadline exceeded\n"},
		{&client.ConnectionError{Err: errors.New("connection refused")}, 5, "connection: connection refused\n"},
		{errors.New("private library detail"), 5, "transport: request failed\n"},
	}
	for _, test := range cases {
		var stdout, stderr bytes.Buffer
		exit := cli.Run([]string{"list"}, func(client.Config) (client.Transport, error) {
			return &fakeTransport{err: test.err, closeErr: errors.New("private close detail")}, nil
		}, &stdout, &stderr)
		if exit != test.exit || stdout.Len() != 0 || stderr.String() != test.stderr {
			t.Fatalf("%v => exit=%d stdout=%q stderr=%q", test.err, exit, stdout.String(), stderr.String())
		}
	}
}

// TestCleanupFailureAfterSuccessReportsConnectionWithNoStdout proves that a
// Close failure is only surfaced when the command itself succeeded, and that
// no partial output is ever written to stdout when cleanup fails.
func TestCleanupFailureAfterSuccessReportsConnectionWithNoStdout(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exit := cli.Run([]string{"list"}, func(client.Config) (client.Transport, error) {
		return &fakeTransport{closeErr: errors.New("private close detail")}, nil
	}, &stdout, &stderr)
	if exit != cli.ExitConnection || stdout.Len() != 0 || stderr.String() != "transport: request failed\n" {
		t.Fatalf("success+cleanup => exit=%d stdout=%q stderr=%q", exit, stdout.String(), stderr.String())
	}
}

func TestNormalizedSettingsReachFactory(t *testing.T) {
	var actual client.Config
	var stdout, stderr bytes.Buffer
	exit := cli.Run([]string{"--base-url", "HTTPS://EXAMPLE.COM/api///", "--timeout", "0.25", "list"},
		func(config client.Config) (client.Transport, error) {
			actual = config
			return &fakeTransport{}, nil
		}, &stdout, &stderr)
	if exit != 0 || actual.BaseURL != "https://example.com/api" || actual.Timeout != 250*time.Millisecond {
		t.Fatalf("exit=%d config=%#v stderr=%q", exit, actual, stderr.String())
	}
}

func TestDurationTimeoutAndTitleUpdateReachMain(t *testing.T) {
	var actual client.Config
	transport := &fakeTransport{}
	var stdout, stderr bytes.Buffer
	exit := cli.Main([]string{"--timeout", "250ms", "update", "7", "--title", "Revised"},
		func(config client.Config) (client.Transport, error) {
			actual = config
			return transport, nil
		}, &stdout, &stderr)
	if exit != 0 || actual.Timeout != 250*time.Millisecond || stderr.Len() != 0 {
		t.Fatalf("exit=%d config=%#v stdout=%q stderr=%q", exit, actual, stdout.String(), stderr.String())
	}
	if transport.input.Title == nil || *transport.input.Title != "Revised" ||
		transport.input.Completed != nil {
		t.Fatalf("update input = %#v", transport.input)
	}
}
