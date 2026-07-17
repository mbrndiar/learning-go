package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/mbrndiar/learning-go/projects/tasks/solution/client"
)

// ParseRequest validates global options and parses one Task subcommand.
func ParseRequest(args []string) (Request, error) {
	settings := Settings{BaseURL: "http://127.0.0.1:8000", Timeout: client.DefaultTimeout}
	global := flag.NewFlagSet("tasks", flag.ContinueOnError)
	global.SetOutput(io.Discard)
	global.StringVar(&settings.BaseURL, "base-url", settings.BaseURL, "")
	timeout := "5"
	global.StringVar(&timeout, "timeout", timeout, "")
	if err := global.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return Request{Help: true}, nil
		}
		return Request{}, err
	}
	if err := parseTimeout(timeout, &settings.Timeout); err != nil {
		return Request{}, err
	}
	normalized, err := client.NormalizeBaseURL(settings.BaseURL)
	if err != nil {
		return Request{}, err
	}
	settings.BaseURL = normalized
	remaining := global.Args()
	if len(remaining) == 0 {
		return Request{}, errors.New("a command is required")
	}
	commandArgs := remaining[1:]
	var command any
	switch remaining[0] {
	case "add":
		if len(commandArgs) != 1 {
			return Request{}, errors.New("add requires exactly one title")
		}
		command = AddCommand{Title: commandArgs[0]}
	case "list":
		flags := flag.NewFlagSet("list", flag.ContinueOnError)
		flags.SetOutput(io.Discard)
		var completed optionalBool
		flags.Var(&completed, "completed", "")
		if err := flags.Parse(commandArgs); err != nil || len(flags.Args()) != 0 {
			return Request{}, errors.New("list accepts only --completed true|false")
		}
		command = ListCommand{Completed: completed.pointer()}
	case "show":
		id, parseErr := parseSingleID(commandArgs)
		if parseErr != nil {
			return Request{}, parseErr
		}
		command = ShowCommand{ID: id}
	case "update":
		if len(commandArgs) == 0 {
			return Request{}, errors.New("update requires a task ID")
		}
		id, parseErr := parseID(commandArgs[0])
		if parseErr != nil {
			return Request{}, parseErr
		}
		flags := flag.NewFlagSet("update", flag.ContinueOnError)
		flags.SetOutput(io.Discard)
		var title optionalString
		var completed optionalBool
		flags.Var(&title, "title", "")
		flags.Var(&completed, "completed", "")
		if err := flags.Parse(commandArgs[1:]); err != nil || len(flags.Args()) != 0 {
			return Request{}, errors.New("update accepts --title and --completed")
		}
		if !title.set && !completed.set {
			return Request{}, errors.New("update requires --title or --completed")
		}
		command = UpdateCommand{ID: id, Title: title.pointer(), Completed: completed.pointer()}
	case "complete":
		id, parseErr := parseSingleID(commandArgs)
		if parseErr != nil {
			return Request{}, parseErr
		}
		command = CompleteCommand{ID: id}
	case "remove":
		id, parseErr := parseSingleID(commandArgs)
		if parseErr != nil {
			return Request{}, parseErr
		}
		command = RemoveCommand{ID: id}
	default:
		return Request{}, fmt.Errorf("unknown command: %s", remaining[0])
	}
	return Request{Settings: settings, Command: command}, nil
}

func parseTimeout(raw string, target *time.Duration) error {
	seconds, err := strconv.ParseFloat(raw, 64)
	if err == nil {
		if seconds <= 0 || math.IsNaN(seconds) || math.IsInf(seconds, 0) {
			return errors.New("timeout must be positive and finite")
		}
		duration := time.Duration(seconds * float64(time.Second))
		if duration <= 0 {
			return errors.New("timeout must be positive and finite")
		}
		*target = duration
		return nil
	}
	duration, durationErr := time.ParseDuration(raw)
	if durationErr != nil || duration <= 0 {
		return errors.New("timeout must be positive and finite")
	}
	*target = duration
	return nil
}

func parseSingleID(args []string) (int64, error) {
	if len(args) != 1 {
		return 0, errors.New("command requires exactly one task ID")
	}
	return parseID(args[0])
}

func parseID(raw string) (int64, error) {
	if raw == "" {
		return 0, errors.New("task ID must be a positive integer")
	}
	for _, value := range []byte(raw) {
		if value < '0' || value > '9' {
			return 0, errors.New("task ID must be a positive integer")
		}
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("task ID must be a positive integer")
	}
	return id, nil
}

type optionalBool struct {
	set   bool
	value bool
}

func (value *optionalBool) String() string {
	if !value.set {
		return ""
	}
	return strconv.FormatBool(value.value)
}

func (value *optionalBool) Set(raw string) error {
	if raw != "true" && raw != "false" {
		return errors.New("must be true or false")
	}
	value.set = true
	value.value = raw == "true"
	return nil
}

func (value *optionalBool) pointer() *bool {
	if !value.set {
		return nil
	}
	result := value.value
	return &result
}

type optionalString struct {
	set   bool
	value string
}

func (value *optionalString) String() string {
	return value.value
}

func (value *optionalString) Set(raw string) error {
	value.set = true
	value.value = raw
	return nil
}

func (value *optionalString) pointer() *string {
	if !value.set {
		return nil
	}
	result := strings.Clone(value.value)
	return &result
}
