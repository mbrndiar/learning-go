package kvstore_test

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/mbrndiar/learning-go/capstones/comparative/solution/kvstore/domain"
	"github.com/mbrndiar/learning-go/capstones/comparative/tests/contract"
	"github.com/mbrndiar/learning-go/capstones/comparative/tests/m1"
	"github.com/mbrndiar/learning-go/capstones/comparative/tests/m2"
	"github.com/mbrndiar/learning-go/capstones/comparative/tests/m3"
	"github.com/mbrndiar/learning-go/capstones/comparative/tests/m4"
	"github.com/mbrndiar/learning-go/capstones/comparative/tests/m5"
)

func TestComparativeProcessHelper(t *testing.T) {
	if !contract.RunProcessHelper() {
		t.Skip("only executed in a conformance helper subprocess")
	}
}

func TestMilestones(t *testing.T) {
	if !domain.Implemented {
		t.Fatal("solution must advertise complete behavior")
	}
	program := buildProgram(t)
	subject := contract.DomainSubject{
		ParseKey: domain.ParseKey,
		ParseExpectation: func(value string, allowAbsent bool) (any, error) {
			return domain.ParseExpectation(value, allowAbsent)
		},
		ParseValue: func(value json.RawMessage) (any, error) {
			return domain.ParseValue(value)
		},
		InspectError: func(err error) (int, string, map[string]any, bool) {
			var contractError *domain.Error
			if !errors.As(err, &contractError) {
				return 0, "", nil, false
			}
			return contractError.ExitCode(), contractError.Category, contractError.Details, true
		},
	}

	t.Run("m1-domain", func(t *testing.T) {
		m1.Run(t, subject)
	})
	t.Run("m2-application", func(t *testing.T) {
		m2.Run(t, program)
	})
	t.Run("m3-storage", func(t *testing.T) {
		m3.Run(t, program)
	})
	t.Run("m4-mutations", func(t *testing.T) {
		m4.Run(t, program)
	})
	t.Run("m5-processes", func(t *testing.T) {
		m5.Run(t, program)
	})
}

func buildProgram(t *testing.T) string {
	t.Helper()
	if program := os.Getenv("COMPARATIVE_KV_PROGRAM"); program != "" {
		if info, err := os.Stat(program); err != nil || info.IsDir() {
			t.Fatalf("COMPARATIVE_KV_PROGRAM is not an executable file: %q", program)
		}
		return program
	}
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate comparative solution")
	}
	repositoryRoot := filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", "..", ".."))
	artifactRoot := filepath.Join(repositoryRoot, "capstones", "comparative", ".conformance")
	if err := os.MkdirAll(artifactRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	directory, err := os.MkdirTemp(artifactRoot, "binary-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(directory)
		entries, readErr := os.ReadDir(artifactRoot)
		if readErr == nil && len(entries) == 0 {
			_ = os.Remove(artifactRoot)
		}
	})
	program := filepath.Join(directory, "kvstore")
	command := exec.Command(
		"go",
		"build",
		"-o",
		program,
		"./capstones/comparative/solution/kvstore/cmd/kvstore",
	)
	command.Dir = repositoryRoot
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("build kvstore: %v\n%s", err, output)
	}
	return program
}
