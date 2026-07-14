// Command task-client is a thin command-line front end for the task API.
//
// It demonstrates how the reusable taskclient.Client is driven from a program
// with its own flags, context, and exit-code handling.
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

	"github.com/mbrndiar/learning-go/project/taskclient"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := run(ctx, os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "task-client:", err)
		os.Exit(1)
	}
}

// run parses arguments and dispatches to the requested command. It is separate
// from main so tests can exercise it with in-memory writers.
func run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("task-client", flag.ContinueOnError)
	flags.SetOutput(stderr)
	baseURL := flags.String("url", "http://localhost:8080", "base URL of the task API")
	timeout := flags.Duration("timeout", taskclient.DefaultTimeout, "per-request timeout")
	flags.Usage = func() {
		fmt.Fprintln(stderr, "usage: task-client [flags] <command> [args]")
		fmt.Fprintln(stderr, "\ncommands:")
		fmt.Fprintln(stderr, "  list                 list all tasks")
		fmt.Fprintln(stderr, "  add <title>          create a task")
		fmt.Fprintln(stderr, "  get <id>             show a single task")
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

	client, err := taskclient.New(*baseURL, taskclient.WithTimeout(*timeout))
	if err != nil {
		return err
	}

	command, commandArgs := rest[0], rest[1:]
	switch command {
	case "list":
		return runList(ctx, client, stdout, commandArgs)
	case "add":
		return runAdd(ctx, client, stdout, commandArgs)
	case "get":
		return runGet(ctx, client, stdout, commandArgs)
	case "complete":
		return runComplete(ctx, client, stdout, commandArgs)
	case "remove":
		return runRemove(ctx, client, stdout, commandArgs)
	default:
		flags.Usage()
		return fmt.Errorf("unknown command %q", command)
	}
}

func runList(ctx context.Context, client *taskclient.Client, stdout io.Writer, args []string) error {
	if len(args) != 0 {
		return errors.New("list takes no arguments")
	}
	tasks, err := client.List(ctx)
	if err != nil {
		return err
	}
	return writeTasks(stdout, tasks)
}

func runAdd(ctx context.Context, client *taskclient.Client, stdout io.Writer, args []string) error {
	if len(args) != 1 {
		return errors.New("add requires exactly one title argument")
	}
	task, err := client.Add(ctx, args[0])
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "added task %d: %s\n", task.ID, task.Title)
	return nil
}

func runGet(ctx context.Context, client *taskclient.Client, stdout io.Writer, args []string) error {
	id, err := parseID(args)
	if err != nil {
		return err
	}
	task, err := client.Get(ctx, id)
	if err != nil {
		return err
	}
	return writeTasks(stdout, []taskclient.Task{task})
}

func runComplete(ctx context.Context, client *taskclient.Client, stdout io.Writer, args []string) error {
	id, err := parseID(args)
	if err != nil {
		return err
	}
	task, err := client.Complete(ctx, id)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "completed task %d: %s\n", task.ID, task.Title)
	return nil
}

func runRemove(ctx context.Context, client *taskclient.Client, stdout io.Writer, args []string) error {
	id, err := parseID(args)
	if err != nil {
		return err
	}
	if err := client.Remove(ctx, id); err != nil {
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

func writeTasks(stdout io.Writer, tasks []taskclient.Task) error {
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
