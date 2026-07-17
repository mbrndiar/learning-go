package cli_test

import (
	"bytes"
	"testing"

	"github.com/mbrndiar/learning-go/projects/tasks/starter/client"
	"github.com/mbrndiar/learning-go/projects/tasks/starter/client/cli"
)

func TestStarterParsesThenStopsExplicitly(t *testing.T) {
	request, err := cli.ParseRequest([]string{"update", "7", "--completed", "false"})
	if err != nil {
		t.Fatal(err)
	}
	command := request.Command.(cli.UpdateCommand)
	if command.Completed == nil || *command.Completed {
		t.Fatal("explicit false was not preserved")
	}
	called := false
	var stdout, stderr bytes.Buffer
	exit := cli.Run([]string{"list"}, func(client.Config) (client.Transport, error) {
		called = true
		return nil, nil
	}, &stdout, &stderr)
	if exit != 1 || called || stdout.Len() != 0 || stderr.String() != "tasks project: not implemented\n" {
		t.Fatalf("exit=%d called=%v stdout=%q stderr=%q", exit, called, stdout.String(), stderr.String())
	}
}
