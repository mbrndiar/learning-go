// Package cli owns the testable monitor command boundary.
package cli

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mbrndiar/learning-go/capstones/idiomatic/solution/monitor/api"
	"github.com/mbrndiar/learning-go/capstones/idiomatic/solution/monitor/domain"
	"github.com/mbrndiar/learning-go/capstones/idiomatic/solution/monitor/history"
	"github.com/mbrndiar/learning-go/capstones/idiomatic/solution/monitor/probe"
	"github.com/mbrndiar/learning-go/capstones/idiomatic/solution/monitor/scheduler"
)

const (
	// ExitOK indicates normal command completion.
	ExitOK = 0
	// ExitUsage indicates invalid command-line usage.
	ExitUsage = 2
	// ExitConfig indicates invalid or unsupported configuration.
	ExitConfig = 3
	// ExitConfigIO indicates a configuration file I/O error.
	ExitConfigIO = 4
	// ExitInternal indicates monitor or server startup/internal failure.
	ExitInternal = 5
	// ExitCancelled indicates a cancelled one-shot check.
	ExitCancelled = 130
)

// Dependencies supplies deterministic application and server seams.
type Dependencies struct {
	Client          *http.Client
	Prober          probe.Prober
	Trigger         scheduler.Trigger
	Listen          func(network, address string) (net.Listener, error)
	Now             func() time.Time
	Logger          *slog.Logger
	ShutdownTimeout time.Duration
}

type commandError struct {
	code    string
	message string
	cause   error
}

func (err *commandError) Error() string {
	if err.cause == nil {
		return err.message
	}
	return fmt.Sprintf("%s: %v", err.message, err.cause)
}

func (err *commandError) Unwrap() error {
	return err.cause
}

type stringList []string

func (values *stringList) String() string {
	return strings.Join(*values, ",")
}

func (values *stringList) Set(value string) error {
	if value == "" {
		return errors.New("target name must not be empty")
	}
	*values = append(*values, value)
	return nil
}

// Run is the stable check/serve process boundary.
func Run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	return RunWithDependencies(ctx, args, stdout, stderr, Dependencies{})
}

// RunWithDependencies runs a command with deterministic test seams.
func RunWithDependencies(
	ctx context.Context,
	args []string,
	stdout io.Writer,
	stderr io.Writer,
	dependencies Dependencies,
) int {
	jsonErrors := false
	if len(args) > 0 && args[0] == "--json-errors" {
		jsonErrors = true
		args = args[1:]
	}
	if len(args) == 0 {
		return fail(stderr, jsonErrors, ExitUsage, &commandError{
			code: "usage", message: "expected check or serve subcommand",
		})
	}
	if dependencies.Now == nil {
		dependencies.Now = time.Now
	}
	if dependencies.Listen == nil {
		dependencies.Listen = net.Listen
	}
	if dependencies.ShutdownTimeout <= 0 {
		dependencies.ShutdownTimeout = 5 * time.Second
	}
	if dependencies.Logger == nil {
		dependencies.Logger = slog.New(slog.NewJSONHandler(stderr, nil))
	}

	switch args[0] {
	case "check":
		return runCheck(ctx, args[1:], stdout, stderr, jsonErrors, dependencies)
	case "serve":
		return runServe(ctx, args[1:], stderr, jsonErrors, dependencies)
	default:
		return fail(stderr, jsonErrors, ExitUsage, &commandError{
			code: "usage", message: fmt.Sprintf("unknown subcommand %q", args[0]),
		})
	}
}

func runCheck(
	ctx context.Context,
	args []string,
	stdout io.Writer,
	stderr io.Writer,
	jsonErrors bool,
	dependencies Dependencies,
) int {
	flags := flag.NewFlagSet("check", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	var configPath string
	var requested stringList
	flags.StringVar(&configPath, "config", "", "configuration path")
	flags.Var(&requested, "target", "target name (repeatable)")
	if err := flags.Parse(args); err != nil || configPath == "" || flags.NArg() != 0 {
		if err == nil {
			err = errors.New("--config is required and positional arguments are not accepted")
		}
		return fail(stderr, jsonErrors, ExitUsage, &commandError{code: "usage", message: err.Error()})
	}

	config, commandErr := loadConfig(configPath)
	if commandErr != nil {
		return failConfig(stderr, jsonErrors, commandErr)
	}
	targets, commandErr := selectTargets(config.Targets, requested)
	if commandErr != nil {
		return fail(stderr, jsonErrors, ExitUsage, commandErr)
	}
	cycleStart := dependencies.Now().UTC().Truncate(time.Millisecond)
	store := history.NewMemoryStore(config.HistoryLimit)
	healthProber := dependencies.Prober
	if healthProber == nil {
		healthProber = probe.NewHTTPProber(dependencies.Client)
	}
	monitor := scheduler.New(healthProber, store, targets, config.MaxConcurrency, scheduler.NewManualTrigger())
	results, err := monitor.RunCycle(ctx)
	if err != nil {
		if errors.Is(err, domain.ErrCancelled) || ctx.Err() != nil {
			return fail(stderr, jsonErrors, ExitCancelled, &commandError{
				code: "cancelled", message: "check was cancelled before a complete report", cause: err,
			})
		}
		return fail(stderr, jsonErrors, ExitInternal, &commandError{
			code: "history_error", message: "check could not commit its results", cause: err,
		})
	}
	var checkedAt time.Time
	for _, result := range results {
		if !result.CheckedAt.IsZero() && (checkedAt.IsZero() || result.CheckedAt.Before(checkedAt)) {
			checkedAt = result.CheckedAt
		}
	}
	if checkedAt.IsZero() {
		checkedAt = cycleStart
	}
	report := domain.CheckReport{
		CheckedAt: checkedAt,
		Summary:   domain.Summarize(results),
		Results:   results,
	}
	if err := json.NewEncoder(stdout).Encode(report); err != nil {
		return fail(stderr, jsonErrors, ExitInternal, &commandError{
			code: "internal", message: "write check report", cause: err,
		})
	}
	return ExitOK
}

func runServe(
	ctx context.Context,
	args []string,
	stderr io.Writer,
	jsonErrors bool,
	dependencies Dependencies,
) int {
	flags := flag.NewFlagSet("serve", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	var configPath string
	var listenAddress string
	flags.StringVar(&configPath, "config", "", "configuration path")
	flags.StringVar(&listenAddress, "listen", "127.0.0.1:8080", "loopback listen address")
	if err := flags.Parse(args); err != nil || configPath == "" || flags.NArg() != 0 {
		if err == nil {
			err = errors.New("--config is required and positional arguments are not accepted")
		}
		return fail(stderr, jsonErrors, ExitUsage, &commandError{code: "usage", message: err.Error()})
	}
	if err := validateLoopbackAddress(listenAddress); err != nil {
		return fail(stderr, jsonErrors, ExitUsage, &commandError{
			code: "usage", message: "listen address must use an explicit loopback host", cause: err,
		})
	}
	config, commandErr := loadConfig(configPath)
	if commandErr != nil {
		return failConfig(stderr, jsonErrors, commandErr)
	}
	listener, err := dependencies.Listen("tcp", listenAddress)
	if err != nil {
		return fail(stderr, jsonErrors, ExitInternal, &commandError{
			code: "server_start", message: "listen for monitor API", cause: err,
		})
	}
	defer listener.Close()

	store := history.NewMemoryStore(config.HistoryLimit)
	healthProber := dependencies.Prober
	if healthProber == nil {
		healthProber = probe.NewHTTPProber(dependencies.Client)
	}
	trigger := dependencies.Trigger
	if trigger == nil {
		trigger = scheduler.NewIntervalTrigger()
	}
	monitor := scheduler.New(healthProber, store, config.Targets, config.MaxConcurrency, trigger)
	schedulerContext, cancelScheduler := context.WithCancel(ctx)
	defer cancelScheduler()
	if err := monitor.Start(schedulerContext); err != nil {
		return fail(stderr, jsonErrors, ExitInternal, &commandError{
			code: "server_start", message: "start monitor scheduler", cause: err,
		})
	}

	state := api.NewState()
	server := &http.Server{
		Handler: api.NewHandlerWithOptions(store, config.Targets, api.Options{
			HistoryLimit: config.HistoryLimit,
			State:        state,
			Logger:       dependencies.Logger,
		}),
		ErrorLog: log.New(io.Discard, "", 0),
	}
	serveDone := make(chan error, 1)
	go func() {
		serveDone <- server.Serve(listener)
	}()
	schedulerDone := make(chan error, 1)
	go func() {
		schedulerDone <- monitor.Wait()
	}()
	dependencies.Logger.InfoContext(ctx, "monitor server started", "listen", listener.Addr().String())

	exitCode := ExitOK
	var exitErr *commandError
	serveJoined := false
	schedulerJoined := false
	select {
	case <-ctx.Done():
	case err := <-serveDone:
		serveJoined = true
		if !errors.Is(err, http.ErrServerClosed) {
			exitCode = ExitInternal
			exitErr = &commandError{code: "server_start", message: "monitor API stopped unexpectedly", cause: err}
		}
	case err := <-schedulerDone:
		schedulerJoined = true
		if err != nil {
			exitCode = ExitInternal
			exitErr = &commandError{code: "history_error", message: "monitor scheduler stopped unexpectedly", cause: err}
		}
	}

	// Publish the stopping state before canceling work, then join every
	// scheduler-owned goroutine before draining HTTP handlers. This prevents
	// probe results from mutating history while server shutdown is in progress.
	state.Stop()
	cancelScheduler()
	if !schedulerJoined {
		err := <-schedulerDone
		if err != nil && exitErr == nil {
			exitCode = ExitInternal
			exitErr = &commandError{code: "history_error", message: "wait for monitor scheduler", cause: err}
		}
	}

	shutdownContext, cancelShutdown := context.WithTimeout(context.Background(), dependencies.ShutdownTimeout)
	shutdownErr := server.Shutdown(shutdownContext)
	cancelShutdown()
	if shutdownErr != nil {
		_ = server.Close()
		if exitErr == nil {
			exitCode = ExitInternal
			exitErr = &commandError{code: "server_start", message: "shut down monitor API", cause: shutdownErr}
		}
	}
	if !serveJoined {
		err := <-serveDone
		if err != nil && !errors.Is(err, http.ErrServerClosed) && exitErr == nil {
			exitCode = ExitInternal
			exitErr = &commandError{code: "server_start", message: "serve monitor API", cause: err}
		}
	}
	if exitErr != nil {
		return fail(stderr, jsonErrors, exitCode, exitErr)
	}
	dependencies.Logger.Info("monitor server stopped")
	return ExitOK
}

func loadConfig(path string) (domain.Config, *commandError) {
	file, err := os.Open(path)
	if err != nil {
		return domain.Config{}, &commandError{code: "config_io", message: "open configuration", cause: err}
	}
	defer file.Close()
	config, err := domain.LoadConfig(file)
	if err == nil {
		return config, nil
	}
	switch {
	case errors.Is(err, domain.ErrConfigIO):
		return domain.Config{}, &commandError{code: "config_io", message: "read configuration", cause: err}
	case errors.Is(err, domain.ErrUnsupportedSchema):
		return domain.Config{}, &commandError{code: "unsupported_schema", message: "unsupported configuration", cause: err}
	case errors.Is(err, domain.ErrDuplicateTarget):
		return domain.Config{}, &commandError{code: "duplicate_target", message: "duplicate target", cause: err}
	default:
		return domain.Config{}, &commandError{code: "invalid_config", message: "invalid configuration", cause: err}
	}
}

func selectTargets(configured []domain.Target, requested []string) ([]domain.Target, *commandError) {
	if len(requested) == 0 {
		return append([]domain.Target(nil), configured...), nil
	}
	wanted := make(map[string]struct{}, len(requested))
	for _, name := range requested {
		if _, exists := domain.TargetByName(configured, name); !exists {
			return nil, &commandError{
				code: "target_not_found", message: fmt.Sprintf("target %q was not configured", name),
				cause: domain.ErrTargetNotFound,
			}
		}
		wanted[name] = struct{}{}
	}
	selected := make([]domain.Target, 0, len(wanted))
	for _, target := range configured {
		if _, exists := wanted[target.Name]; exists {
			selected = append(selected, target)
		}
	}
	return selected, nil
}

func validateLoopbackAddress(address string) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return err
	}
	if strings.EqualFold(host, "localhost") {
		return nil
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return errors.New("host is not loopback")
	}
	return nil
}

func failConfig(stderr io.Writer, jsonErrors bool, err *commandError) int {
	if err.code == "config_io" {
		return fail(stderr, jsonErrors, ExitConfigIO, err)
	}
	return fail(stderr, jsonErrors, ExitConfig, err)
}

func fail(stderr io.Writer, jsonErrors bool, exitCode int, err *commandError) int {
	if jsonErrors {
		_ = json.NewEncoder(stderr).Encode(domain.ErrorResponse{
			Error: domain.APIError{Code: err.code, Message: err.message},
		})
		return exitCode
	}
	fmt.Fprintf(stderr, "monitor: %s: %s\n", err.code, err.message)
	return exitCode
}
