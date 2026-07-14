// Command task-manager is the top-level CLI for the capstone. It selects a
// storage backend—local JSON file or the remote REST API—then runs the
// requested task command through a shared Manager.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/mbrndiar/learning-go/project/taskclient"
	"github.com/mbrndiar/learning-go/project/taskmanager"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := run(ctx, os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "task-manager:", err)
		os.Exit(1)
	}
}

// run parses flags, builds a Manager over the selected backend, and dispatches
// to the requested command. It is separated from main for testing.
func run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("task-manager", flag.ContinueOnError)
	flags.SetOutput(stderr)
	backend := flags.String("backend", "file", "storage backend: file or rest")
	file := flags.String("file", "tasks.json", "path to the JSON task file (file backend)")
	baseURL := flags.String("url", "http://localhost:8080", "base URL of the task API (rest backend)")
	timeout := flags.Duration("timeout", taskclient.DefaultTimeout, "per-operation timeout")
	flags.Usage = func() {
		fmt.Fprintln(stderr, "usage: task-manager [flags] <command> [args]")
		fmt.Fprintln(stderr, "\ncommands:")
		fmt.Fprintln(stderr, "  list                 list all tasks")
		fmt.Fprintln(stderr, "  add <title>          create a task")
		fmt.Fprintln(stderr, "  complete <id>        mark a task as done")
		fmt.Fprintln(stderr, "  remove <id>          delete a task")
		fmt.Fprintln(stderr, "\nflags:")
		flags.PrintDefaults()
	}

	if err := flags.Parse(args); err != nil {
		return err
	}

	rest := flags.Args()
	if len(rest) == 0 {
		flags.Usage()
		return errors.New("a command is required")
	}

	manager, err := buildManager(*backend, *file, *baseURL, *timeout)
	if err != nil {
		return err
	}

	opCtx := ctx
	if *timeout > 0 {
		var cancel context.CancelFunc
		opCtx, cancel = context.WithTimeout(ctx, *timeout)
		defer cancel()
	}

	command, commandArgs := rest[0], rest[1:]
	switch command {
	case "list":
		return runList(opCtx, manager, stdout, commandArgs)
	case "add":
		return runAdd(opCtx, manager, stdout, commandArgs)
	case "complete":
		return runComplete(opCtx, manager, stdout, commandArgs)
	case "remove":
		return runRemove(opCtx, manager, stdout, commandArgs)
	default:
		flags.Usage()
		return fmt.Errorf("unknown command %q", command)
	}
}

// buildManager constructs a Manager over the requested backend.
func buildManager(backend, file, baseURL string, timeout time.Duration) (*taskmanager.Manager, error) {
	var storage taskmanager.Storage

	switch backend {
	case "file":
		fileStorage, err := taskmanager.NewFileStorage(file)
		if err != nil {
			return nil, err
		}
		storage = fileStorage
	case "rest":
		client, err := taskclient.New(baseURL, taskclient.WithTimeout(timeout))
		if err != nil {
			return nil, err
		}
		restStorage, err := taskmanager.NewRESTStorage(client)
		if err != nil {
			return nil, err
		}
		storage = restStorage
	default:
		return nil, fmt.Errorf("unknown backend %q (want file or rest)", backend)
	}

	return taskmanager.NewManager(storage)
}

func runList(ctx context.Context, manager *taskmanager.Manager, stdout io.Writer, args []string) error {
	if len(args) != 0 {
		return errors.New("list takes no arguments")
	}
	tasks, err := manager.List(ctx)
	if err != nil {
		return err
	}
	return writeTasks(stdout, tasks)
}

func runAdd(ctx context.Context, manager *taskmanager.Manager, stdout io.Writer, args []string) error {
	if len(args) != 1 {
		return errors.New("add requires exactly one title argument")
	}
	task, err := manager.Add(ctx, args[0])
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "added task %d: %s\n", task.ID, task.Title)
	return nil
}

func runComplete(ctx context.Context, manager *taskmanager.Manager, stdout io.Writer, args []string) error {
	id, err := parseID(args)
	if err != nil {
		return err
	}
	task, err := manager.Complete(ctx, id)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "completed task %d: %s\n", task.ID, task.Title)
	return nil
}

func runRemove(ctx context.Context, manager *taskmanager.Manager, stdout io.Writer, args []string) error {
	id, err := parseID(args)
	if err != nil {
		return err
	}
	if err := manager.Remove(ctx, id); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "removed task %d\n", id)
	return nil
}

func parseID(args []string) (int, error) {
	if len(args) != 1 {
		return 0, errors.New("exactly one id argument is required")
	}
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return 0, fmt.Errorf("invalid id %q: %w", args[0], err)
	}
	return id, nil
}

func writeTasks(stdout io.Writer, tasks []taskmanager.Task) error {
	if len(tasks) == 0 {
		fmt.Fprintln(stdout, "no tasks")
		return nil
	}

	writer := tabwriter.NewWriter(stdout, 0, 2, 2, ' ', 0)
	fmt.Fprintln(writer, "ID\tDONE\tTITLE")
	for _, task := range tasks {
		fmt.Fprintf(writer, "%d\t%s\t%s\n", task.ID, doneMark(task.Done), task.Title)
	}
	return writer.Flush()
}

func doneMark(done bool) string {
	if done {
		return "x"
	}
	return " "
}
