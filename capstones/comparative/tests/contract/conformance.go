package contract

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
	"unicode/utf8"

	_ "modernc.org/sqlite"
)

const (
	specVersion    = "1.0.0"
	processTimeout = 15 * time.Second
)

var commandID atomic.Uint64

// DomainSubject adapts one implementation's domain API to Milestone 1.
type DomainSubject struct {
	ParseKey         func(string) (string, error)
	ParseExpectation func(string, bool) (any, error)
	ParseValue       func(json.RawMessage) (any, error)
	InspectError     func(error) (int, string, map[string]any, bool)
}

// RunMilestone1 validates the domain model directly against shared fixtures.
func RunMilestone1(t *testing.T, subject DomainSubject) {
	t.Helper()
	assertSpecVersions(t)

	constants := readFixture(t, "fixtures/contract.json")
	requireKeys(t, object(t, constants), []string{
		"kind", "spec_id", "spec_version", "key_pattern", "key_max_bytes",
		"safe_integer_min", "safe_integer_max", "value_input_max_utf8_bytes",
		"max_container_depth", "busy_timeout_ms", "commands",
		"set_expectations", "delete_expectations", "exit_codes",
	})
	if stringValue(t, field(t, constants, "kind")) != "constants" {
		t.Fatal("contract fixture has unexpected kind")
	}
	if stringValue(t, field(t, constants, "spec_id")) != "comparative-kv" ||
		integer(t, field(t, constants, "key_max_bytes")) != 128 ||
		integer(t, field(t, constants, "safe_integer_max")) != 9007199254740991 ||
		integer(t, field(t, constants, "value_input_max_utf8_bytes")) != 65536 ||
		integer(t, field(t, constants, "max_container_depth")) != 32 ||
		integer(t, field(t, constants, "busy_timeout_ms")) != 10000 {
		t.Fatal("implementation constants do not match contract.json")
	}
	for key, expected := range map[string]any{
		"commands":            []any{"set", "get", "delete", "list"},
		"set_expectations":    []any{"any", "absent", "exact_revision"},
		"delete_expectations": []any{"any", "exact_revision"},
		"exit_codes": map[string]any{
			"success": json.Number("0"), "validation": json.Number("2"),
			"conflict": json.Number("3"), "not_found": json.Number("4"),
			"storage": json.Number("5"),
		},
	} {
		if !semanticEqual(field(t, constants, key), expected) {
			t.Fatalf("contract constant %s does not match", key)
		}
	}

	keys := readFixture(t, "fixtures/keys.json")
	requireExactKeys(t, object(t, keys), []string{"kind", "spec_version", "accepted", "rejected", "ordering"})
	if stringValue(t, field(t, keys, "kind")) != "key_cases" {
		t.Fatal("keys fixture has unexpected kind")
	}
	for _, caseValue := range array(t, field(t, keys, "accepted")) {
		requireKeys(t, object(t, caseValue), []string{"id", "key", "key_generator"})
		key := generatedKey(t, caseValue)
		actual, err := subject.ParseKey(key)
		if err != nil || actual != key {
			t.Fatalf("accepted key %q: actual=%q err=%v", key, actual, err)
		}
	}
	for _, caseValue := range array(t, field(t, keys, "rejected")) {
		requireKeys(t, object(t, caseValue), []string{"id", "key", "key_generator"})
		key := generatedKey(t, caseValue)
		_, err := subject.ParseKey(key)
		assertDomainError(
			t,
			subject,
			err,
			2,
			"invalid_argument",
			map[string]any{"field": "key", "reason": "format"},
		)
	}

	for _, test := range []struct {
		value       string
		allowAbsent bool
		valid       bool
	}{
		{"any", true, true},
		{"absent", true, true},
		{"1", false, true},
		{"9007199254740991", false, true},
		{"", true, false},
		{"absent", false, false},
		{"0", true, false},
		{"01", true, false},
		{"9007199254740992", true, false},
	} {
		_, err := subject.ParseExpectation(test.value, test.allowAbsent)
		if test.valid && err != nil {
			t.Fatalf("ParseExpectation(%q, %t): %v", test.value, test.allowAbsent, err)
		}
		if !test.valid {
			assertDomainError(
				t,
				subject,
				err,
				2,
				"invalid_argument",
				map[string]any{"field": "expect", "reason": "format"},
			)
		}
	}

	accepted := readFixture(t, "fixtures/values-accepted.json")
	requireExactKeys(t, object(t, accepted), []string{"kind", "spec_version", "cases"})
	if stringValue(t, field(t, accepted, "kind")) != "accepted_value_cases" {
		t.Fatal("accepted-values fixture has unexpected kind")
	}
	for _, caseValue := range array(t, field(t, accepted, "cases")) {
		requireKeys(t, object(t, caseValue), []string{
			"id", "input_json", "input_generator", "normalized", "normalized_generator",
		})
		input := generatedInput(t, caseValue)
		expected := normalizeFixtureNumbers(generatedNormalized(t, caseValue))
		actual, err := subject.ParseValue(json.RawMessage(input))
		if err != nil {
			t.Fatalf("accepted value %s: %v", stringValue(t, field(t, caseValue, "id")), err)
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Fatalf("accepted value %s = %#v, want %#v", stringValue(t, field(t, caseValue, "id")), actual, expected)
		}
	}

	rejected := readFixture(t, "fixtures/values-rejected.json")
	requireExactKeys(t, object(t, rejected), []string{"kind", "spec_version", "cases"})
	if stringValue(t, field(t, rejected, "kind")) != "rejected_value_cases" {
		t.Fatal("rejected-values fixture has unexpected kind")
	}
	for _, caseValue := range array(t, field(t, rejected, "cases")) {
		requireKeys(t, object(t, caseValue), []string{
			"id", "input_json", "input_generator", "exit", "category", "details",
		})
		input := generatedInput(t, caseValue)
		_, err := subject.ParseValue(json.RawMessage(input))
		assertDomainError(
			t,
			subject,
			err,
			int(integer(t, field(t, caseValue, "exit"))),
			stringValue(t, field(t, caseValue, "category")),
			object(t, field(t, caseValue, "details")),
		)
	}

	actual, err := subject.ParseValue(json.RawMessage(`0e-999999999999999999999`))
	if err != nil || !reflect.DeepEqual(actual, int64(0)) {
		t.Fatalf("large-exponent zero = %#v, %v", actual, err)
	}
	_, err = subject.ParseValue(json.RawMessage(`1e-400`))
	assertDomainError(
		t,
		subject,
		err,
		2,
		"invalid_value",
		map[string]any{"reason": "non_integral_number"},
	)
	for _, test := range []struct {
		input  string
		reason string
	}{
		{`[1.5,1e400]`, "non_integral_number"},
		{`{"a":1.5,"a":1,"b":1e400}`, "non_finite_number"},
		{`{"\uD800":1e400}`, "unpaired_surrogate"},
	} {
		_, err = subject.ParseValue(json.RawMessage(test.input))
		assertDomainError(
			t,
			subject,
			err,
			2,
			"invalid_value",
			map[string]any{"reason": test.reason},
		)
	}
	actual, err = subject.ParseValue(json.RawMessage(`{"x":1.5,"x":1}`))
	if err != nil || !reflect.DeepEqual(actual, map[string]any{"x": int64(1)}) {
		t.Fatalf("overwritten invalid member = %#v, %v", actual, err)
	}
	_, err = subject.ParseValue(json.RawMessage(`"\uD800" trailing`))
	assertDomainError(
		t,
		subject,
		err,
		2,
		"invalid_json",
		map[string]any{"reason": "syntax"},
	)
}

// RunMilestone2 validates exact CLI parsing, envelopes, and validation order.
func RunMilestone2(t *testing.T, program string) {
	t.Helper()
	runSequentialFixture(t, program, "fixtures/scenarios/invalid.json")
	assertAdditionalCLIGrammar(t, program)
}

// RunMilestone3 validates initialization, persistence, migration, and schema checks.
func RunMilestone3(t *testing.T, program string) {
	t.Helper()
	runSequentialFixture(t, program, "fixtures/scenarios/normal.json")
	runSequentialFixture(t, program, "fixtures/scenarios/migration.json")
	assertV1StorageInvariants(t, program)
}

// RunMilestone4 validates boundaries, ordering, revisions, and mutation semantics.
func RunMilestone4(t *testing.T, program string) {
	t.Helper()
	runSequentialFixture(t, program, "fixtures/scenarios/boundary.json")
}

// RunMilestone5 validates the complete real-child-process scenario fixture.
// Scenarios spawn real OS processes (the CLI under test and helper
// subprocesses re-invoking this test binary) rather than goroutines, because
// only distinct processes give independent process boundaries, their own
// runtime and connection pools, and real exec/exit behavior, which is what
// is needed to exercise genuine cross-process SQLite locking and
// busy_timeout contention.
func RunMilestone5(t *testing.T, program string) {
	t.Helper()
	fixture := readFixture(t, "fixtures/scenarios/multiprocess.json")
	requireExactKeys(t, object(t, fixture), []string{"kind", "spec_version", "scenarios"})
	if stringValue(t, field(t, fixture, "kind")) != "multiprocess_scenarios" {
		t.Fatal("multiprocess fixture has unexpected kind")
	}
	for _, scenario := range array(t, field(t, fixture, "scenarios")) {
		requireKeys(t, object(t, scenario), []string{"id", "repeat", "database", "setup", "operations"})
		repeat := int(integer(t, field(t, scenario, "repeat")))
		for repetition := 0; repetition < repeat; repetition++ {
			t.Run(
				fmt.Sprintf("%s/%d", stringValue(t, field(t, scenario, "id")), repetition+1),
				func(t *testing.T) {
					runMultiprocessScenario(t, program, scenario)
				},
			)
		}
	}
}

// RunProcessHelper executes the actor or SQLite-lock role in a test subprocess.
// It returns false when the current process was not launched as a helper.
func RunProcessHelper() bool {
	mode := os.Getenv("KV_HELPER_MODE")
	if mode == "" {
		return false
	}
	switch mode {
	case "actor":
		runActorHelper()
	case "lock":
		runLockHelper()
	default:
		fmt.Fprintln(os.Stderr, "unknown comparative helper mode")
		os.Exit(1)
	}
	return true
}

// runActorHelper is the thin wrapper executed by the re-invoked test binary:
// it signals readiness, blocks until the coordinator releases every actor at
// once, and only then replaces itself with the real CLI. This lets
// runParallelGroup start all actors as OS processes ahead of time and still
// have them invoke the CLI simultaneously, so tests can assert on true
// concurrent contention rather than on staggered process-launch timing.
func runActorHelper() {
	ready := requiredEnvironment("KV_HELPER_READY")
	release := requiredEnvironment("KV_HELPER_RELEASE")
	program := requiredEnvironment("KV_HELPER_PROGRAM")
	var args []string
	if err := json.Unmarshal([]byte(requiredEnvironment("KV_HELPER_ARGS")), &args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	signalFile(ready)
	waitForFileProcess(release, 30*time.Second)

	command := exec.Command(program, args...)
	command.Stdin = nil
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			os.Exit(exitError.ExitCode())
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}

// runLockHelper holds a real SQLite write lock in a separate process so the
// CLI under test can be observed contending for it. All statements below run
// on the same pinned *sql.Conn from db.Conn(), so BEGIN IMMEDIATE and the
// later ROLLBACK are guaranteed to execute on the identical connection that
// acquired the lock; SetMaxOpenConns(1) simply caps the pool so this helper
// cannot open another, unneeded connection alongside it.
func runLockHelper() {
	database := requiredEnvironment("KV_HELPER_DATABASE")
	ready := requiredEnvironment("KV_HELPER_READY")
	release := requiredEnvironment("KV_HELPER_RELEASE")
	db, err := sql.Open("sqlite", database)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	connection, err := db.Conn(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer connection.Close()
	if _, err := connection.ExecContext(context.Background(), `PRAGMA busy_timeout = 10000`); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	// BEGIN IMMEDIATE acquires SQLite's write lock immediately (rather than
	// deferring it to the first write), and is deliberately never committed:
	// the transaction is held open until the coordinator sends the release
	// signal, giving the test a deterministic window in which a contending
	// process must observe busy_timeout behavior instead of racing to acquire
	// the lock first.
	if _, err := connection.ExecContext(context.Background(), `BEGIN IMMEDIATE`); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	signalFile(ready)
	waitForFileProcess(release, 30*time.Second)
	// ROLLBACK (not COMMIT) releases the lock without persisting any change,
	// since this transaction only ever existed to hold contention open.
	if _, err := connection.ExecContext(context.Background(), `ROLLBACK`); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func assertSpecVersions(t *testing.T) {
	t.Helper()
	version, err := os.ReadFile(specPath("SPEC_VERSION"))
	if err != nil {
		t.Fatal(err)
	}
	if string(version) != specVersion+"\n" {
		t.Fatalf("SPEC_VERSION = %q", version)
	}
	for _, path := range []string{
		"fixtures/contract.json",
		"fixtures/keys.json",
		"fixtures/values-accepted.json",
		"fixtures/values-rejected.json",
		"fixtures/scenarios/normal.json",
		"fixtures/scenarios/boundary.json",
		"fixtures/scenarios/invalid.json",
		"fixtures/scenarios/migration.json",
		"fixtures/scenarios/multiprocess.json",
	} {
		fixture := readFixture(t, path)
		if stringValue(t, field(t, fixture, "spec_version")) != specVersion {
			t.Fatalf("%s has wrong spec_version", path)
		}
	}
}

func runSequentialFixture(t *testing.T, program, relativePath string) {
	t.Helper()
	fixture := readFixture(t, relativePath)
	requireExactKeys(t, object(t, fixture), []string{"kind", "spec_version", "scenarios"})
	if stringValue(t, field(t, fixture, "kind")) != "sequential_scenarios" {
		t.Fatalf("%s has unexpected kind", relativePath)
	}
	for _, scenario := range array(t, field(t, fixture, "scenarios")) {
		requireKeys(t, object(t, scenario), []string{"id", "database", "setup", "steps"})
		t.Run(stringValue(t, field(t, scenario, "id")), func(t *testing.T) {
			runSequentialScenario(t, program, scenario)
		})
	}
}

func runSequentialScenario(t *testing.T, program string, scenario any) {
	t.Helper()
	directory := scenarioDirectory(t, stringValue(t, field(t, scenario, "id")))
	database := filepath.Join(directory, "store.db")
	missingParent := filepath.Join(directory, "missing-parent", "child")
	defer cleanupScenario(t, database, directory)

	switch stringValue(t, field(t, scenario, "database")) {
	case "fresh":
	case "sqlite_setup":
		setupDatabase(t, database, field(t, scenario, "setup"))
	default:
		t.Fatal("unknown sequential database kind")
	}

	for _, stepValue := range array(t, field(t, scenario, "steps")) {
		step := object(t, stepValue)
		switch {
		case step["fixture_references"] != nil:
			requireExactKeys(t, step, []string{"fixture_references"})
			for _, reference := range array(t, step["fixture_references"]) {
				switch stringValue(t, reference) {
				case "../keys.json":
					runKeyCasesCLI(t, program)
				case "../values-accepted.json":
					runAcceptedValueCasesCLI(t, program)
				case "../values-rejected.json":
					runRejectedValueCasesCLI(t, program)
				default:
					t.Fatalf("unknown fixture reference %v", reference)
				}
			}
		case step["run"] != nil:
			requireExactKeys(t, step, []string{"run", "expect"})
			run := object(t, step["run"])
			requireExactKeys(t, run, []string{"args"})
			args := substitutedArgs(t, field(t, run, "args"), database, missingParent, nil)
			result := runProgram(t, program, args, directory, processTimeout)
			assertExpectation(t, result, field(t, step, "expect"))
		case step["sqlite_assert"] != nil:
			requireExactKeys(t, step, []string{"sqlite_assert"})
			runSQLiteAssertions(t, database, step["sqlite_assert"])
		default:
			t.Fatalf("unknown sequential operation: %#v", step)
		}
	}
	if _, err := os.Stat(database); err == nil {
		assertIntegrity(t, database)
	}
}

func runKeyCasesCLI(t *testing.T, program string) {
	t.Helper()
	fixture := readFixture(t, "fixtures/keys.json")
	requireExactKeys(t, object(t, fixture), []string{"kind", "spec_version", "accepted", "rejected", "ordering"})
	for _, caseValue := range array(t, field(t, fixture, "accepted")) {
		requireKeys(t, object(t, caseValue), []string{"id", "key", "key_generator"})
		directory := scenarioDirectory(t, "accepted-key")
		database := filepath.Join(directory, "store.db")
		key := generatedKey(t, caseValue)
		set := runProgram(t, program, []string{
			"--db", database, "set", key, "--value-json", "null",
		}, directory, processTimeout)
		if set.exitCode != 0 {
			t.Fatalf("accepted key set failed: %#v", set)
		}
		get := runProgram(t, program, []string{
			"--db", database, "get", key,
		}, directory, processTimeout)
		if get.exitCode != 0 ||
			stringValue(t, field(t, field(t, get.envelope, "result"), "key")) != key {
			t.Fatalf("accepted key get failed: %#v", get)
		}
		assertIntegrity(t, database)
		cleanupScenario(t, database, directory)
	}
	for _, caseValue := range array(t, field(t, fixture, "rejected")) {
		requireKeys(t, object(t, caseValue), []string{"id", "key", "key_generator"})
		directory := scenarioDirectory(t, "rejected-key")
		database := filepath.Join(directory, "store.db")
		result := runProgram(t, program, []string{
			"--db", database, "get", generatedKey(t, caseValue),
		}, directory, processTimeout)
		assertStandard(t, result, 2, map[string]any{
			"ok": false,
			"error": map[string]any{
				"category": "invalid_argument",
				"details":  map[string]any{"field": "key", "reason": "format"},
			},
		}, "")
		if _, err := os.Stat(database); !os.IsNotExist(err) {
			t.Fatal("invalid key created storage")
		}
		cleanupScenario(t, database, directory)
	}

	directory := scenarioDirectory(t, "key-ordering")
	database := filepath.Join(directory, "store.db")
	ordering := array(t, field(t, fixture, "ordering"))
	for index := len(ordering) - 1; index >= 0; index-- {
		result := runProgram(t, program, []string{
			"--db", database, "set", stringValue(t, ordering[index]), "--value-json", "null",
		}, directory, processTimeout)
		if result.exitCode != 0 {
			t.Fatalf("ordering set failed: %#v", result)
		}
	}
	listed := runProgram(t, program, []string{"--db", database, "list"}, directory, processTimeout)
	entries := array(t, field(t, field(t, listed.envelope, "result"), "entries"))
	for index, expected := range ordering {
		if stringValue(t, field(t, entries[index], "key")) != stringValue(t, expected) {
			t.Fatalf("binary ordering mismatch at %d", index)
		}
	}
	assertIntegrity(t, database)
	cleanupScenario(t, database, directory)
}

func runAcceptedValueCasesCLI(t *testing.T, program string) {
	t.Helper()
	fixture := readFixture(t, "fixtures/values-accepted.json")
	requireExactKeys(t, object(t, fixture), []string{"kind", "spec_version", "cases"})
	for _, caseValue := range array(t, field(t, fixture, "cases")) {
		requireKeys(t, object(t, caseValue), []string{
			"id", "input_json", "input_generator", "normalized", "normalized_generator",
		})
		directory := scenarioDirectory(t, "accepted-value")
		database := filepath.Join(directory, "store.db")
		input := generatedInput(t, caseValue)
		expected := generatedNormalized(t, caseValue)
		set := runProgram(t, program, []string{
			"--db", database, "set", "value", "--value-json", input, "--expect", "absent",
		}, directory, processTimeout)
		if set.exitCode != 0 ||
			!semanticEqual(field(t, field(t, set.envelope, "result"), "value"), expected) ||
			integer(t, field(t, field(t, set.envelope, "result"), "revision")) != 1 ||
			field(t, field(t, set.envelope, "result"), "created") != true {
			t.Fatalf("accepted value set failed: %#v", set)
		}
		get := runProgram(t, program, []string{
			"--db", database, "get", "value",
		}, directory, processTimeout)
		if !semanticEqual(field(t, field(t, get.envelope, "result"), "value"), expected) {
			t.Fatalf("accepted value get failed: %#v", get)
		}
		assertIntegrity(t, database)
		cleanupScenario(t, database, directory)
	}
}

func runRejectedValueCasesCLI(t *testing.T, program string) {
	t.Helper()
	fixture := readFixture(t, "fixtures/values-rejected.json")
	requireExactKeys(t, object(t, fixture), []string{"kind", "spec_version", "cases"})
	for _, caseValue := range array(t, field(t, fixture, "cases")) {
		requireKeys(t, object(t, caseValue), []string{
			"id", "input_json", "input_generator", "exit", "category", "details",
		})
		directory := scenarioDirectory(t, "rejected-value")
		database := filepath.Join(directory, "store.db")
		result := runProgram(t, program, []string{
			"--db", database, "set", "value", "--value-json", generatedInput(t, caseValue),
		}, directory, processTimeout)
		assertStandard(t, result, int(integer(t, field(t, caseValue, "exit"))), map[string]any{
			"ok": false,
			"error": map[string]any{
				"category": stringValue(t, field(t, caseValue, "category")),
				"details":  field(t, caseValue, "details"),
			},
		}, "")
		if _, err := os.Stat(database); !os.IsNotExist(err) {
			t.Fatal("invalid value created storage")
		}
		cleanupScenario(t, database, directory)
	}
}

func assertAdditionalCLIGrammar(t *testing.T, program string) {
	t.Helper()
	directory := scenarioDirectory(t, "exact-cli")
	database := filepath.Join(directory, "store=with-equals.db")
	defer cleanupScenario(t, database, directory)
	for _, args := range [][]string{
		{"--db=" + database, "list"},
		{"--db", database, "list", "extra"},
		{"--db", database, "set", "key", "--value-json=1"},
		{"--db", database, "set", "key", "--value-json", "1", "--expect=any"},
	} {
		result := runProgram(t, program, args, directory, processTimeout)
		assertStandard(t, result, 2, map[string]any{
			"ok": false,
			"error": map[string]any{
				"category": "usage",
				"details":  map[string]any{"reason": "invalid_cli"},
			},
		}, "")
		if _, err := os.Stat(database); !os.IsNotExist(err) {
			t.Fatal("usage failure created storage")
		}
	}
	for _, test := range []struct {
		args     []string
		category string
		details  map[string]any
	}{
		{
			[]string{"--db", ":memory:", "get", "-bad"},
			"invalid_argument",
			map[string]any{"field": "db", "reason": "unsupported_form"},
		},
		{
			[]string{"--db", filepath.Join(directory, "missing", "store.db"), "set", "-bad", "--value-json", "1.5", "--expect", "01"},
			"invalid_argument",
			map[string]any{"field": "key", "reason": "format"},
		},
		{
			[]string{"--db", filepath.Join(directory, "missing", "store.db"), "set", "valid", "--value-json", "1.5", "--expect", ""},
			"invalid_argument",
			map[string]any{"field": "expect", "reason": "format"},
		},
	} {
		result := runProgram(t, program, test.args, directory, processTimeout)
		assertStandard(t, result, 2, map[string]any{
			"ok": false,
			"error": map[string]any{
				"category": test.category,
				"details":  test.details,
			},
		}, "")
	}
	valid := runProgram(t, program, []string{
		"--db", database, "set", "equals", "--value-json", `"a=b"`,
	}, directory, processTimeout)
	if valid.exitCode != 0 ||
		stringValue(t, field(t, field(t, valid.envelope, "result"), "value")) != "a=b" {
		t.Fatalf("literal equals handling failed: %#v", valid)
	}
}

func assertV1StorageInvariants(t *testing.T, program string) {
	t.Helper()
	cases := []struct {
		id         string
		statements []string
		details    map[string]any
	}{
		{
			"invalid-key",
			[]string{
				`INSERT INTO store_metadata(singleton, schema_version, global_revision) VALUES (1, 1, 1)`,
				`INSERT INTO entries(key, value_json, revision) VALUES ('-bad', 'null', 1)`,
			},
			map[string]any{"reason": "invalid_key", "key": "-bad"},
		},
		{
			"invalid-value",
			[]string{
				`INSERT INTO store_metadata(singleton, schema_version, global_revision) VALUES (1, 1, 1)`,
				`INSERT INTO entries(key, value_json, revision) VALUES ('good', '1.5', 1)`,
			},
			map[string]any{"reason": "invalid_value", "key": "good"},
		},
		{
			"non-normalized-whitespace",
			[]string{
				`INSERT INTO store_metadata(singleton, schema_version, global_revision) VALUES (1, 1, 1)`,
				`INSERT INTO entries(key, value_json, revision) VALUES ('good', ' null ', 1)`,
			},
			map[string]any{"reason": "invalid_value", "key": "good"},
		},
		{
			"non-normalized-number",
			[]string{
				`INSERT INTO store_metadata(singleton, schema_version, global_revision) VALUES (1, 1, 1)`,
				`INSERT INTO entries(key, value_json, revision) VALUES ('good', '1.0', 1)`,
			},
			map[string]any{"reason": "invalid_value", "key": "good"},
		},
		{
			"non-normalized-duplicate",
			[]string{
				`INSERT INTO store_metadata(singleton, schema_version, global_revision) VALUES (1, 1, 1)`,
				`INSERT INTO entries(key, value_json, revision) VALUES ('good', '{"a":1,"a":2}', 1)`,
			},
			map[string]any{"reason": "invalid_value", "key": "good"},
		},
		{
			"duplicate-revision",
			[]string{
				`INSERT INTO store_metadata(singleton, schema_version, global_revision) VALUES (1, 1, 1)`,
				`INSERT INTO entries(key, value_json, revision) VALUES ('a', 'null', 1)`,
				`INSERT INTO entries(key, value_json, revision) VALUES ('b', 'true', 1)`,
			},
			map[string]any{"reason": "revision_invariant"},
		},
		{
			"revision-ahead",
			[]string{
				`INSERT INTO store_metadata(singleton, schema_version, global_revision) VALUES (1, 1, 1)`,
				`INSERT INTO entries(key, value_json, revision) VALUES ('a', 'null', 2)`,
			},
			map[string]any{"reason": "revision_invariant"},
		},
	}
	for _, test := range cases {
		t.Run(test.id, func(t *testing.T) {
			directory := scenarioDirectory(t, test.id)
			database := filepath.Join(directory, "store.db")
			defer cleanupScenario(t, database, directory)
			db := openSQLite(t, database)
			createV1Tables(t, db)
			for _, statement := range test.statements {
				if _, err := db.Exec(statement); err != nil {
					t.Fatal(err)
				}
			}
			if err := db.Close(); err != nil {
				t.Fatal(err)
			}
			result := runProgram(t, program, []string{"--db", database, "list"}, directory, processTimeout)
			assertStandard(t, result, 5, map[string]any{
				"ok": false,
				"error": map[string]any{
					"category": "invalid_storage",
					"details":  test.details,
				},
			}, "")
			assertIntegrity(t, database)
		})
	}

	t.Run("nondefault-pragmas", func(t *testing.T) {
		directory := scenarioDirectory(t, "nondefault-pragmas")
		database := filepath.Join(directory, "store.db")
		defer cleanupScenario(t, database, directory)
		db := openSQLite(t, database)
		if _, err := db.Exec(`PRAGMA user_version = 7`); err != nil {
			t.Fatal(err)
		}
		createV1Tables(t, db)
		if _, err := db.Exec(
			`INSERT INTO store_metadata(singleton, schema_version, global_revision) VALUES (1, 1, 0)`,
		); err != nil {
			t.Fatal(err)
		}
		if err := db.Close(); err != nil {
			t.Fatal(err)
		}
		result := runProgram(t, program, []string{"--db", database, "list"}, directory, processTimeout)
		assertStandard(t, result, 5, map[string]any{
			"ok": false,
			"error": map[string]any{
				"category": "invalid_storage",
				"details":  map[string]any{"reason": "malformed_schema"},
			},
		}, "")
	})
}

func createV1Tables(t *testing.T, db *sql.DB) {
	t.Helper()
	for _, statement := range []string{
		`CREATE TABLE store_metadata (
			singleton INTEGER PRIMARY KEY CHECK (singleton = 1),
			schema_version INTEGER NOT NULL CHECK (schema_version = 1),
			global_revision INTEGER NOT NULL CHECK (global_revision BETWEEN 0 AND 9007199254740991)
		)`,
		`CREATE TABLE entries (
			key TEXT PRIMARY KEY COLLATE BINARY,
			value_json TEXT NOT NULL CHECK (json_valid(value_json)),
			revision INTEGER NOT NULL CHECK (revision BETWEEN 1 AND 9007199254740991)
		)`,
	} {
		if _, err := db.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}
}

func runMultiprocessScenario(t *testing.T, program string, scenario any) {
	t.Helper()
	directory := scenarioDirectory(t, stringValue(t, field(t, scenario, "id")))
	database := filepath.Join(directory, "store.db")
	missingParent := filepath.Join(directory, "missing-parent", "child")
	defer cleanupScenario(t, database, directory)

	switch stringValue(t, field(t, scenario, "database")) {
	case "fresh":
	case "sqlite_setup":
		setupDatabase(t, database, field(t, scenario, "setup"))
	default:
		t.Fatal("unknown multiprocess database kind")
	}

	captures := make(map[string]any)
	// locks and running track scenario-declared subprocesses by their fixture
	// id so later operations (release_lock_helper, await_cli) can address a
	// specific still-live process. The deferred terminate loop is a safety
	// net for scenarios that fail mid-way; well-formed scenarios drain both
	// maps themselves, which is asserted below.
	locks := make(map[string]*runningProcess)
	running := make(map[string]*runningProcess)
	defer func() {
		for _, process := range running {
			process.terminate()
		}
		for _, process := range locks {
			process.terminate()
		}
	}()

	for _, operationValue := range array(t, field(t, scenario, "operations")) {
		operation := object(t, operationValue)
		if len(operation) != 1 {
			t.Fatal("multiprocess operation must have one kind")
		}
		switch {
		case operation["parallel"] != nil:
			runParallelGroup(t, program, database, directory, operation["parallel"])
		case operation["run_assert"] != nil:
			runSingleAssert(
				t,
				program,
				database,
				missingParent,
				directory,
				operation["run_assert"],
				captures,
			)
		case operation["start_lock_helper"] != nil:
			start := object(t, operation["start_lock_helper"])
			requireExactKeys(t, start, []string{"id"})
			id := stringValue(t, field(t, start, "id"))
			locks[id] = startLockHelper(t, database, directory, id)
		case operation["start_cli"] != nil:
			start := object(t, operation["start_cli"])
			requireExactKeys(t, start, []string{"id", "args"})
			id := stringValue(t, field(t, start, "id"))
			args := substitutedArgs(t, field(t, start, "args"), database, missingParent, nil)
			running[id] = startProgram(t, program, args, directory, "running-"+id, nil)
		case operation["sleep_ms"] != nil:
			time.Sleep(time.Duration(integer(t, operation["sleep_ms"])) * time.Millisecond)
		case operation["release_lock_helper"] != nil:
			release := object(t, operation["release_lock_helper"])
			requireExactKeys(t, release, []string{"id"})
			id := stringValue(t, field(t, release, "id"))
			process := locks[id]
			if process == nil {
				t.Fatalf("unknown lock helper %q", id)
			}
			signalFile(process.releasePath)
			result := process.finish(t, 15*time.Second)
			if result.exitCode != 0 {
				t.Fatalf("lock helper failed: %#v", result)
			}
			removeFile(t, process.readyPath)
			removeFile(t, process.releasePath)
			delete(locks, id)
		case operation["await_cli"] != nil:
			await := object(t, operation["await_cli"])
			requireKeys(t, await, []string{"id", "expect", "assert"})
			id := stringValue(t, field(t, await, "id"))
			process := running[id]
			if process == nil {
				t.Fatalf("unknown running CLI %q", id)
			}
			result := process.finish(t, 20*time.Second)
			assertExpectation(t, result, field(t, await, "expect"))
			if assertion := await["assert"]; assertion != nil {
				assertDuration(t, result, assertion)
			}
			delete(running, id)
		default:
			t.Fatalf("unknown multiprocess operation: %#v", operation)
		}
	}
	if len(running) != 0 || len(locks) != 0 {
		t.Fatal("all child processes must be awaited")
	}
	if _, err := os.Stat(database); err == nil {
		assertIntegrity(t, database)
	}
}

func runParallelGroup(t *testing.T, program, database, directory string, parallel any) {
	t.Helper()
	group := object(t, parallel)
	requireExactKeys(t, group, []string{"actors_generator", "assert"})
	generator := object(t, field(t, group, "actors_generator"))
	requireKeys(t, generator, []string{"kind", "count", "pad_width", "args"})
	if stringValue(t, field(t, generator, "kind")) != "indexed_commands" {
		t.Fatal("unknown actor generator")
	}
	count := int(integer(t, field(t, generator, "count")))
	padWidth := 0
	if value := generator["pad_width"]; value != nil {
		padWidth = int(integer(t, value))
	}
	release := filepath.Join(directory, fmt.Sprintf("parallel-%d.release", commandID.Add(1)))
	actors := make([]*runningProcess, 0, count)
	defer func() {
		for _, actor := range actors {
			actor.terminate()
		}
	}()
	for index := 0; index < count; index++ {
		number := index + 1
		replacements := map[string]string{
			"i":        strconv.Itoa(index),
			"n":        strconv.Itoa(number),
			"padded_n": fmt.Sprintf("%0*d", padWidth, number),
		}
		args := substitutedArgs(
			t,
			field(t, generator, "args"),
			database,
			filepath.Join(directory, "missing"),
			replacements,
		)
		ready := filepath.Join(directory, fmt.Sprintf("parallel-%d-%d.ready", commandID.Add(1), index))
		encodedArgs, err := json.Marshal(args)
		if err != nil {
			t.Fatal(err)
		}
		environment := []string{
			"KV_HELPER_MODE=actor",
			"KV_HELPER_READY=" + ready,
			"KV_HELPER_RELEASE=" + release,
			"KV_HELPER_PROGRAM=" + program,
			"KV_HELPER_ARGS=" + string(encodedArgs),
		}
		process := startProgram(
			t,
			os.Args[0],
			[]string{"-test.run=^TestComparativeProcessHelper$"},
			directory,
			fmt.Sprintf("actor-%d", index),
			environment,
		)
		process.readyPath = ready
		process.releasePath = release
		process.args = args
		actors = append(actors, process)
	}
	// Wait for every actor to report readiness before releasing any of them:
	// this is a barrier that guarantees all actor processes have already
	// been spawned and are blocked immediately before their real exec, so
	// signaling the single shared release file below causes them to race
	// against each other for real rather than starting at staggered times.
	for _, actor := range actors {
		waitForFile(t, actor.readyPath, 15*time.Second)
	}
	releasedAt := time.Now()
	// started is reset to the moment of release (not process spawn), so
	// duration assertions measure the actual contended operation instead of
	// including barrier wait time.
	for _, actor := range actors {
		actor.started = releasedAt
	}
	signalFile(release)
	results := make([]actorResult, 0, len(actors))
	for _, actor := range actors {
		result := actor.finish(t, 30*time.Second)
		results = append(results, actorResult{args: actor.args, result: result})
		removeFile(t, actor.readyPath)
	}
	removeFile(t, release)
	assertParallel(t, program, database, directory, results, field(t, group, "assert"))
}

type actorResult struct {
	args   []string
	result runResult
}

func assertParallel(
	t *testing.T,
	program, database, directory string,
	actors []actorResult,
	assertionValue any,
) {
	t.Helper()
	assertion := object(t, assertionValue)
	requireKeys(t, assertion, []string{
		"all_exit", "all_ok", "stdout_semantic_all", "success_count",
		"category_counts", "result_revision_set", "success_revision",
		"conflict_actual", "not_found_count", "winner_value_matches_final",
		"duration_less_than_ms", "duration_at_least_ms",
	})
	for _, actor := range actors {
		if actor.result.stderr != "" {
			t.Fatalf("actor stderr = %q", actor.result.stderr)
		}
	}
	if value := assertion["all_exit"]; value != nil {
		expected := int(integer(t, value))
		for _, actor := range actors {
			if actor.result.exitCode != expected {
				t.Fatalf("actor result = %#v, want exit %d", actor.result, expected)
			}
		}
	}
	if value := assertion["all_ok"]; value != nil {
		for _, actor := range actors {
			if field(t, actor.result.envelope, "ok") != value {
				t.Fatalf("actor ok mismatch: %#v", actor.result)
			}
		}
	}
	if value := assertion["stdout_semantic_all"]; value != nil {
		for _, actor := range actors {
			if !semanticEqual(actor.result.envelope, value) {
				t.Fatalf("actor envelope = %#v, want %#v", actor.result.envelope, value)
			}
		}
	}

	successes := make([]actorResult, 0)
	categoryCounts := make(map[string]int)
	for _, actor := range actors {
		if ok, _ := field(t, actor.result.envelope, "ok").(bool); ok {
			successes = append(successes, actor)
		} else {
			category := stringValue(t, field(t, field(t, actor.result.envelope, "error"), "category"))
			categoryCounts[category]++
		}
	}
	if value := assertion["success_count"]; value != nil &&
		len(successes) != int(integer(t, value)) {
		t.Fatalf("success count = %d", len(successes))
	}
	if value := assertion["category_counts"]; value != nil {
		expected := object(t, value)
		if len(categoryCounts) != len(expected) {
			t.Fatalf("category counts = %#v, want %#v", categoryCounts, expected)
		}
		for category, count := range expected {
			if categoryCounts[category] != int(integer(t, count)) {
				t.Fatalf("category %s count = %d", category, categoryCounts[category])
			}
		}
	}
	if value := assertion["result_revision_set"]; value != nil {
		actual := make([]int64, 0, len(successes))
		for _, actor := range successes {
			actual = append(actual, integer(t, field(t, field(t, actor.result.envelope, "result"), "revision")))
		}
		assertIntegerRange(t, actual, value)
	}
	if value := assertion["success_revision"]; value != nil {
		expected := integer(t, value)
		for _, actor := range successes {
			if integer(t, field(t, field(t, actor.result.envelope, "result"), "revision")) != expected {
				t.Fatal("successful actor has unexpected revision")
			}
		}
	}
	if value := assertion["conflict_actual"]; value != nil {
		for _, actor := range actors {
			if errorCategory(t, actor.result.envelope) == "conflict" &&
				!semanticEqual(field(t, field(t, field(t, actor.result.envelope, "error"), "details"), "actual"), value) {
				t.Fatal("conflict actual mismatch")
			}
		}
	}
	if value := assertion["not_found_count"]; value != nil {
		count := 0
		for _, actor := range actors {
			if errorCategory(t, actor.result.envelope) == "not_found" {
				count++
			}
		}
		if count != int(integer(t, value)) {
			t.Fatalf("not_found count = %d", count)
		}
	}
	if assertion["winner_value_matches_final"] == true {
		if len(successes) != 1 {
			t.Fatal("winner assertion requires one success")
		}
		winner := successes[0]
		keyIndex := stringIndex(winner.args, "set") + 1
		valueIndex := stringIndex(winner.args, "--value-json") + 1
		final := runProgram(t, program, []string{
			"--db", database, "get", winner.args[keyIndex],
		}, directory, processTimeout)
		expected := decodeJSON(t, []byte(winner.args[valueIndex]))
		if !semanticEqual(field(t, field(t, final.envelope, "result"), "value"), expected) {
			t.Fatal("winning value does not match final state")
		}
	}
	for _, actor := range actors {
		assertDuration(t, actor.result, assertionValue)
	}
}

func runSingleAssert(
	t *testing.T,
	program, database, missingParent, directory string,
	runAssertValue any,
	captures map[string]any,
) {
	t.Helper()
	runAssert := object(t, runAssertValue)
	requireKeys(t, runAssert, []string{"args", "expect", "assert", "capture"})
	args := substitutedArgs(t, field(t, runAssert, "args"), database, missingParent, nil)
	result := runProgram(t, program, args, directory, processTimeout)
	assertExpectation(t, result, field(t, runAssert, "expect"))
	if assertion := runAssert["assert"]; assertion != nil {
		assertStructural(t, result, assertion, captures)
	}
	if capture := runAssert["capture"]; capture != nil {
		captures[stringValue(t, capture)] = field(t, result.envelope, "result")
	}
}

func assertStructural(t *testing.T, result runResult, assertionValue any, captures map[string]any) {
	t.Helper()
	assertion := object(t, assertionValue)
	requireKeys(t, assertion, []string{
		"keys_in_order", "global_revision", "entry_count", "entry_revision_set",
		"values_by_key", "revision_by_key", "state_unchanged_from",
		"duration_less_than_ms", "duration_at_least_ms",
	})
	resultValue := field(t, result.envelope, "result")
	var entries []any
	if value := object(t, resultValue)["entries"]; value != nil {
		entries = array(t, value)
	}
	if value := assertion["keys_in_order"]; value != nil {
		actual := make([]any, 0, len(entries))
		for _, entry := range entries {
			actual = append(actual, field(t, entry, "key"))
		}
		if !semanticEqual(actual, value) {
			t.Fatalf("keys = %#v, want %#v", actual, value)
		}
	}
	if value := assertion["global_revision"]; value != nil &&
		!semanticEqual(field(t, resultValue, "global_revision"), value) {
		t.Fatal("global revision mismatch")
	}
	if value := assertion["entry_count"]; value != nil && len(entries) != int(integer(t, value)) {
		t.Fatal("entry count mismatch")
	}
	if value := assertion["entry_revision_set"]; value != nil {
		revisions := make([]int64, 0, len(entries))
		for _, entry := range entries {
			revisions = append(revisions, integer(t, field(t, entry, "revision")))
		}
		assertIntegerRange(t, revisions, value)
	}
	if value := assertion["values_by_key"]; value != nil {
		byKey := entriesByKey(t, entries)
		for key, expected := range object(t, value) {
			if !semanticEqual(field(t, byKey[key], "value"), expected) {
				t.Fatalf("value for %s mismatch", key)
			}
		}
	}
	if value := assertion["revision_by_key"]; value != nil {
		byKey := entriesByKey(t, entries)
		for key, expected := range object(t, value) {
			if !semanticEqual(field(t, byKey[key], "revision"), expected) {
				t.Fatalf("revision for %s mismatch", key)
			}
		}
	}
	if value := assertion["state_unchanged_from"]; value != nil {
		expected := captures[stringValue(t, value)]
		if !semanticEqual(resultValue, expected) {
			t.Fatal("captured state changed")
		}
	}
	assertDuration(t, result, assertionValue)
}

func entriesByKey(t *testing.T, entries []any) map[string]any {
	t.Helper()
	result := make(map[string]any, len(entries))
	for _, entry := range entries {
		result[stringValue(t, field(t, entry, "key"))] = entry
	}
	return result
}

func assertIntegerRange(t *testing.T, actual []int64, rangeValue any) {
	t.Helper()
	sort.Slice(actual, func(left, right int) bool { return actual[left] < actual[right] })
	rangeObject := object(t, rangeValue)
	from := integer(t, field(t, rangeObject, "from"))
	to := integer(t, field(t, rangeObject, "to"))
	if len(actual) != int(to-from+1) {
		t.Fatalf("revision set = %v, want %d..%d", actual, from, to)
	}
	for index := range actual {
		if actual[index] != from+int64(index) {
			t.Fatalf("revision set = %v, want %d..%d", actual, from, to)
		}
	}
}

func assertDuration(t *testing.T, result runResult, assertionValue any) {
	t.Helper()
	assertion := object(t, assertionValue)
	if value := assertion["duration_less_than_ms"]; value != nil {
		limit := time.Duration(integer(t, value)) * time.Millisecond
		if result.duration >= limit {
			t.Fatalf("duration %v is not less than %v", result.duration, limit)
		}
	}
	if value := assertion["duration_at_least_ms"]; value != nil {
		limit := time.Duration(integer(t, value)) * time.Millisecond
		if result.duration < limit {
			t.Fatalf("duration %v is less than %v", result.duration, limit)
		}
	}
}

func startLockHelper(t *testing.T, database, directory, id string) *runningProcess {
	t.Helper()
	ready := filepath.Join(directory, fmt.Sprintf("lock-%s-%d.ready", id, commandID.Add(1)))
	release := filepath.Join(directory, fmt.Sprintf("lock-%s-%d.release", id, commandID.Add(1)))
	process := startProgram(
		t,
		os.Args[0],
		[]string{"-test.run=^TestComparativeProcessHelper$"},
		directory,
		"lock-"+id,
		[]string{
			"KV_HELPER_MODE=lock",
			"KV_HELPER_DATABASE=" + database,
			"KV_HELPER_READY=" + ready,
			"KV_HELPER_RELEASE=" + release,
		},
	)
	process.readyPath = ready
	process.releasePath = release
	process.parseEnvelope = false
	readyObserved := false
	// If waitForFile below fails, t.Fatal unwinds this goroutine via
	// Goexit, running this deferred cleanup while readyObserved is still
	// false; once the ready file is actually seen we flip the flag so the
	// helper (and its held lock) survives for the caller to use.
	defer func() {
		if !readyObserved {
			process.terminate()
		}
	}()
	// Blocking here until the ready file appears is what guarantees that,
	// once this function returns, the helper's BEGIN IMMEDIATE has already
	// succeeded and the lock is genuinely held: callers can rely on real
	// contention rather than racing the helper to acquire the lock first.
	waitForFile(t, ready, 15*time.Second)
	readyObserved = true
	return process
}

type runResult struct {
	exitCode int
	stdout   string
	stderr   string
	envelope any
	duration time.Duration
}

type runningProcess struct {
	command    *exec.Cmd
	stdoutPath string
	stderrPath string
	started    time.Time
	// finished guards against awaiting (finish) or reaping (terminate) the
	// same OS process twice; both check/set it, and since startProgram
	// registers terminate via t.Cleanup, an already-finished process is
	// safely ignored by that deferred safety net.
	finished bool
	// readyPath/releasePath are only populated for helper subprocesses
	// (actor/lock); a plain CLI run under runProgram/startProgram leaves
	// them empty, and terminate tolerates that by discarding os.Remove's
	// error on a nonexistent ("") path.
	readyPath     string
	releasePath   string
	args          []string
	parseEnvelope bool
}

func runProgram(
	t *testing.T,
	program string,
	args []string,
	directory string,
	timeout time.Duration,
) runResult {
	t.Helper()
	process := startProgram(t, program, args, directory, "command", nil)
	return process.finish(t, timeout)
}

func startProgram(
	t *testing.T,
	program string,
	args []string,
	directory, label string,
	environment []string,
) *runningProcess {
	t.Helper()
	id := commandID.Add(1)
	stdoutPath := filepath.Join(directory, fmt.Sprintf("%s-%d.stdout", label, id))
	stderrPath := filepath.Join(directory, fmt.Sprintf("%s-%d.stderr", label, id))
	stdout, err := os.Create(stdoutPath)
	if err != nil {
		t.Fatal(err)
	}
	stderr, err := os.Create(stderrPath)
	if err != nil {
		stdout.Close()
		t.Fatal(err)
	}
	command := exec.Command(program, args...)
	command.Stdin = nil
	command.Stdout = stdout
	command.Stderr = stderr
	commandEnvironment := append(os.Environ(), environment...)
	if coverageRoot := os.Getenv("GOCOVERDIR"); coverageRoot != "" {
		coverageDirectory := filepath.Join(coverageRoot, fmt.Sprintf("process-%d", id))
		if err := os.MkdirAll(coverageDirectory, 0o700); err != nil {
			stdout.Close()
			stderr.Close()
			t.Fatal(err)
		}
		commandEnvironment = append(commandEnvironment, "GOCOVERDIR="+coverageDirectory)
	}
	command.Env = commandEnvironment
	configureProcessGroup(command)
	if err := command.Start(); err != nil {
		stdout.Close()
		stderr.Close()
		t.Fatal(err)
	}
	process := &runningProcess{
		command:       command,
		stdoutPath:    stdoutPath,
		stderrPath:    stderrPath,
		started:       time.Now(),
		args:          append([]string(nil), args...),
		parseEnvelope: true,
	}
	t.Cleanup(process.terminate)
	if err := stdout.Close(); err != nil {
		t.Fatal(err)
	}
	if err := stderr.Close(); err != nil {
		t.Fatal(err)
	}
	return process
}

func (process *runningProcess) finish(t *testing.T, timeout time.Duration) runResult {
	t.Helper()
	if process.finished {
		t.Fatal("process awaited twice")
	}
	wait := make(chan error, 1)
	go func() {
		wait <- process.command.Wait()
	}()
	var waitErr error
	select {
	case waitErr = <-wait:
	case <-time.After(timeout):
		terminateProcessGroup(process.command.Process.Pid)
		<-wait
		process.finished = true
		t.Fatalf("child %d exceeded %v", process.command.Process.Pid, timeout)
	}
	process.finished = true
	duration := time.Since(process.started)
	stdoutBytes, err := os.ReadFile(process.stdoutPath)
	if err != nil {
		t.Fatal(err)
	}
	stderrBytes, err := os.ReadFile(process.stderrPath)
	if err != nil {
		t.Fatal(err)
	}
	removeFile(t, process.stdoutPath)
	removeFile(t, process.stderrPath)

	exitCode := 0
	if waitErr != nil {
		var exitError *exec.ExitError
		if !errors.As(waitErr, &exitError) {
			t.Fatal(waitErr)
		}
		exitCode = exitError.ExitCode()
	}
	var envelope any
	if process.parseEnvelope {
		envelope = parseStdout(t, stdoutBytes)
		assertCommandResultShape(t, process.args, envelope)
	}
	return runResult{
		exitCode: exitCode,
		stdout:   string(stdoutBytes),
		stderr:   string(stderrBytes),
		envelope: envelope,
		duration: duration,
	}
}

func (process *runningProcess) terminate() {
	if process == nil || process.finished || process.command.Process == nil {
		return
	}
	terminateProcessGroup(process.command.Process.Pid)
	_ = process.command.Wait()
	process.finished = true
	_ = os.Remove(process.stdoutPath)
	_ = os.Remove(process.stderrPath)
	_ = os.Remove(process.readyPath)
	_ = os.Remove(process.releasePath)
}

func parseStdout(t *testing.T, stdout []byte) any {
	t.Helper()
	if !utf8.Valid(stdout) {
		t.Fatal("stdout is not UTF-8")
	}
	if bytes.HasPrefix(stdout, []byte{0xef, 0xbb, 0xbf}) {
		t.Fatal("stdout contains BOM")
	}
	if len(stdout) == 0 || stdout[len(stdout)-1] != '\n' {
		t.Fatalf("stdout lacks final LF: %q", stdout)
	}
	body := stdout[:len(stdout)-1]
	if bytes.ContainsAny(body, "\r\n") {
		t.Fatalf("stdout is not exactly one line: %q", stdout)
	}
	assertCompactJSON(t, body)
	value := decodeJSON(t, body)
	root := object(t, value)
	ok, valid := root["ok"].(bool)
	if !valid {
		t.Fatal("stdout ok member is not boolean")
	}
	if ok {
		requireExactKeys(t, root, []string{"ok", "result"})
	} else {
		requireExactKeys(t, root, []string{"ok", "error"})
		errorValue := object(t, root["error"])
		requireExactKeys(t, errorValue, []string{"category", "details"})
		assertErrorDetailsShape(t, errorValue)
	}
	assertNormalizedNumbers(t, value)
	return value
}

func assertCompactJSON(t *testing.T, body []byte) {
	t.Helper()
	inString := false
	escaped := false
	for _, character := range body {
		if inString {
			if escaped {
				escaped = false
			} else if character == '\\' {
				escaped = true
			} else if character == '"' {
				inString = false
			}
			continue
		}
		if character == '"' {
			inString = true
		} else if character == ' ' || character == '\t' || character == '\r' || character == '\n' {
			t.Fatal("stdout JSON contains insignificant whitespace")
		}
	}
}

func assertNormalizedNumbers(t *testing.T, value any) {
	t.Helper()
	switch value := value.(type) {
	case json.Number:
		text := value.String()
		if text != "0" {
			unsigned := strings.TrimPrefix(text, "-")
			if unsigned == "" || unsigned[0] == '0' {
				t.Fatalf("noncanonical output number %q", text)
			}
			for _, character := range []byte(unsigned) {
				if character < '0' || character > '9' {
					t.Fatalf("non-integral output number %q", text)
				}
			}
		}
		integerValue, err := strconv.ParseInt(text, 10, 64)
		if err != nil || integerValue < -9007199254740991 || integerValue > 9007199254740991 {
			t.Fatalf("unsafe output number %q", text)
		}
	case []any:
		for _, item := range value {
			assertNormalizedNumbers(t, item)
		}
	case map[string]any:
		for _, item := range value {
			assertNormalizedNumbers(t, item)
		}
	}
}

func assertCommandResultShape(t *testing.T, args []string, envelope any) {
	t.Helper()
	if ok, _ := field(t, envelope, "ok").(bool); !ok {
		return
	}
	result := object(t, field(t, envelope, "result"))
	if len(args) < 3 {
		t.Fatal("successful command lacks command argument")
	}
	switch args[2] {
	case "set":
		requireExactKeys(t, result, []string{"key", "value", "revision", "created"})
	case "get":
		requireExactKeys(t, result, []string{"key", "value", "revision"})
	case "delete":
		requireExactKeys(t, result, []string{"key", "deleted_revision", "revision"})
	case "list":
		requireExactKeys(t, result, []string{"entries", "global_revision"})
		for _, entry := range array(t, field(t, result, "entries")) {
			requireExactKeys(t, object(t, entry), []string{"key", "value", "revision"})
		}
	default:
		t.Fatalf("unexpected successful command %q", args[2])
	}
}

func assertErrorDetailsShape(t *testing.T, errorValue map[string]any) {
	t.Helper()
	category := stringValue(t, field(t, errorValue, "category"))
	details := object(t, field(t, errorValue, "details"))
	switch category {
	case "usage", "invalid_json", "invalid_value":
		requireExactKeys(t, details, []string{"reason"})
	case "invalid_argument":
		requireExactKeys(t, details, []string{"field", "reason"})
	case "conflict":
		requireExactKeys(t, details, []string{"key", "expected", "actual"})
	case "not_found":
		requireExactKeys(t, details, []string{"key"})
	case "busy":
		requireExactKeys(t, details, []string{"timeout_ms"})
	case "unsupported_schema":
		requireExactKeys(t, details, []string{"found", "supported"})
	case "invalid_storage":
		if details["key"] == nil {
			requireExactKeys(t, details, []string{"reason"})
		} else {
			requireExactKeys(t, details, []string{"reason", "key"})
		}
	case "revision_exhausted":
		requireExactKeys(t, details, []string{"maximum"})
	case "storage_error":
		requireExactKeys(t, details, []string{"operation", "reason"})
	default:
		t.Fatalf("unknown error category %q", category)
	}
}

func assertExpectation(t *testing.T, result runResult, expectationValue any) {
	t.Helper()
	expectation := object(t, expectationValue)
	requireKeys(t, expectation, []string{"exit", "stdout", "stderr"})
	if result.exitCode != int(integer(t, field(t, expectation, "exit"))) {
		t.Fatalf("exit = %d, want %d; stdout=%s stderr=%s", result.exitCode, integer(t, field(t, expectation, "exit")), result.stdout, result.stderr)
	}
	if result.stderr != stringValue(t, field(t, expectation, "stderr")) {
		t.Fatalf("stderr = %q", result.stderr)
	}
	if expected := expectation["stdout"]; expected != nil && !semanticEqual(result.envelope, expected) {
		t.Fatalf("stdout = %#v, want %#v", result.envelope, expected)
	}
}

func assertStandard(t *testing.T, result runResult, exit int, stdout any, stderr string) {
	t.Helper()
	if result.exitCode != exit || result.stderr != stderr || !semanticEqual(result.envelope, stdout) {
		t.Fatalf("result = %#v, want exit=%d stdout=%#v stderr=%q", result, exit, stdout, stderr)
	}
}

func assertDomainError(
	t *testing.T,
	subject DomainSubject,
	err error,
	exit int,
	category string,
	details map[string]any,
) {
	t.Helper()
	actualExit, actualCategory, actualDetails, ok := subject.InspectError(err)
	if !ok {
		t.Fatalf("error %v is not a structured domain error", err)
	}
	if actualExit != exit ||
		actualCategory != category ||
		!semanticEqual(actualDetails, details) {
		t.Fatalf(
			"error = exit %d category %q details %#v, want %d %q %#v",
			actualExit,
			actualCategory,
			actualDetails,
			exit,
			category,
			details,
		)
	}
}

func setupDatabase(t *testing.T, database string, setupValue any) {
	t.Helper()
	setup := object(t, setupValue)
	requireExactKeys(t, setup, []string{"statements"})
	db := openSQLite(t, database)
	for _, statement := range array(t, field(t, setup, "statements")) {
		if _, err := db.Exec(stringValue(t, statement)); err != nil {
			db.Close()
			t.Fatal(err)
		}
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
}

func runSQLiteAssertions(t *testing.T, database string, assertionValue any) {
	t.Helper()
	assertion := object(t, assertionValue)
	requireExactKeys(t, assertion, []string{"queries"})
	db := openSQLite(t, database)
	defer db.Close()
	for _, queryValue := range array(t, field(t, assertion, "queries")) {
		query := object(t, queryValue)
		requireExactKeys(t, query, []string{"sql", "rows"})
		rows, err := db.Query(stringValue(t, field(t, query, "sql")))
		if err != nil {
			t.Fatal(err)
		}
		columns, err := rows.Columns()
		if err != nil {
			rows.Close()
			t.Fatal(err)
		}
		actual := make([]any, 0)
		for rows.Next() {
			values := make([]any, len(columns))
			destinations := make([]any, len(columns))
			for index := range values {
				destinations[index] = &values[index]
			}
			if err := rows.Scan(destinations...); err != nil {
				rows.Close()
				t.Fatal(err)
			}
			for index := range values {
				values[index] = sqliteFixtureValue(values[index])
			}
			actual = append(actual, values)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			t.Fatal(err)
		}
		if err := rows.Close(); err != nil {
			t.Fatal(err)
		}
		if !semanticEqual(actual, field(t, query, "rows")) {
			t.Fatalf("SQLite rows = %#v, want %#v", actual, field(t, query, "rows"))
		}
	}
}

func sqliteFixtureValue(value any) any {
	switch value := value.(type) {
	case int64:
		return json.Number(strconv.FormatInt(value, 10))
	case float64:
		return json.Number(strconv.FormatFloat(value, 'g', -1, 64))
	case []byte:
		bytesValue := make([]any, len(value))
		for index, item := range value {
			bytesValue[index] = json.Number(strconv.Itoa(int(item)))
		}
		return bytesValue
	default:
		return value
	}
}

func assertIntegrity(t *testing.T, database string) {
	t.Helper()
	db := openSQLite(t, database)
	var result string
	if err := db.QueryRow(`PRAGMA integrity_check`).Scan(&result); err != nil {
		db.Close()
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if result != "ok" {
		t.Fatalf("integrity_check = %q", result)
	}
}

func openSQLite(t *testing.T, database string) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", database)
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	if err := db.Ping(); err != nil {
		db.Close()
		t.Fatal(err)
	}
	return db
}

func cleanupScenario(t *testing.T, database, directory string) {
	t.Helper()
	for _, suffix := range []string{"", "-wal", "-shm", "-journal"} {
		path := database + suffix
		err := os.Remove(path)
		if err != nil && !os.IsNotExist(err) {
			t.Errorf("remove %s: %v", path, err)
		}
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("%s remained after cleanup", path)
		}
	}
	entries, err := os.ReadDir(directory)
	if err == nil {
		for _, entry := range entries {
			path := filepath.Join(directory, entry.Name())
			if removeErr := os.Remove(path); removeErr != nil {
				t.Errorf("remove artifact %s: %v", path, removeErr)
			}
		}
	}
	if err := os.Remove(directory); err != nil && !os.IsNotExist(err) {
		t.Errorf("remove scenario directory %s: %v", directory, err)
	}
}

func scenarioDirectory(t *testing.T, label string) string {
	t.Helper()
	root := filepath.Join(filepath.Dir(filepath.Dir(specPath("SPEC_VERSION"))), ".conformance")
	if err := os.MkdirAll(root, 0o700); err != nil {
		t.Fatal(err)
	}
	directory, err := os.MkdirTemp(root, label+"-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(directory)
		entries, readErr := os.ReadDir(root)
		if readErr == nil && len(entries) == 0 {
			_ = os.Remove(root)
		}
	})
	return directory
}

func generatedKey(t *testing.T, caseValue any) string {
	t.Helper()
	caseObject := object(t, caseValue)
	if key := caseObject["key"]; key != nil {
		return stringValue(t, key)
	}
	generator := object(t, field(t, caseObject, "key_generator"))
	requireExactKeys(t, generator, []string{"kind", "prefix", "character", "count"})
	if stringValue(t, field(t, generator, "kind")) != "repeat_suffix" {
		t.Fatal("unknown key generator")
	}
	return stringValue(t, field(t, generator, "prefix")) +
		strings.Repeat(
			stringValue(t, field(t, generator, "character")),
			int(integer(t, field(t, generator, "count"))),
		)
}

func generatedInput(t *testing.T, caseValue any) string {
	t.Helper()
	caseObject := object(t, caseValue)
	if input := caseObject["input_json"]; input != nil {
		return stringValue(t, input)
	}
	return generateInput(t, field(t, caseObject, "input_generator"))
}

func generateInput(t *testing.T, generatorValue any) string {
	t.Helper()
	generator := object(t, generatorValue)
	switch stringValue(t, field(t, generator, "kind")) {
	case "nested_arrays":
		requireExactKeys(t, generator, []string{"kind", "depth", "leaf"})
		value := nestedValue(
			int(integer(t, field(t, generator, "depth"))),
			field(t, generator, "leaf"),
		)
		encoded, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		return string(encoded)
	case "ascii_string_total_bytes":
		requireExactKeys(t, generator, []string{"kind", "character", "total_bytes"})
		total := int(integer(t, field(t, generator, "total_bytes")))
		character := stringValue(t, field(t, generator, "character"))
		if len(character) != 1 {
			t.Fatal("ASCII generator character is not one byte")
		}
		result := `"` + strings.Repeat(character, total-2) + `"`
		if len(result) != total {
			t.Fatal("generated input has wrong byte length")
		}
		return result
	default:
		t.Fatal("unknown input generator")
		return ""
	}
}

func generatedNormalized(t *testing.T, caseValue any) any {
	t.Helper()
	caseObject := object(t, caseValue)
	if value, exists := caseObject["normalized"]; exists {
		return value
	}
	generator := object(t, field(t, caseObject, "normalized_generator"))
	switch stringValue(t, field(t, generator, "kind")) {
	case "nested_arrays":
		requireExactKeys(t, generator, []string{"kind", "depth", "leaf"})
		return nestedValue(
			int(integer(t, field(t, generator, "depth"))),
			field(t, generator, "leaf"),
		)
	case "ascii_string_total_bytes":
		requireExactKeys(t, generator, []string{"kind", "character", "total_bytes"})
		total := int(integer(t, field(t, generator, "total_bytes")))
		return strings.Repeat(stringValue(t, field(t, generator, "character")), total-2)
	default:
		t.Fatal("unknown normalized generator")
		return nil
	}
}

func nestedValue(depth int, leaf any) any {
	value := leaf
	for range depth {
		value = []any{value}
	}
	return value
}

func substitutedArgs(
	t *testing.T,
	value any,
	database, missingParent string,
	replacements map[string]string,
) []string {
	t.Helper()
	values := array(t, value)
	result := make([]string, 0, len(values))
	for _, argumentValue := range values {
		argument := strings.ReplaceAll(stringValue(t, argumentValue), "${DB}", database)
		argument = strings.ReplaceAll(argument, "${MISSING_PARENT}", missingParent)
		for key, replacement := range replacements {
			argument = strings.ReplaceAll(argument, "${"+key+"}", replacement)
		}
		result = append(result, argument)
	}
	return result
}

func readFixture(t *testing.T, relativePath string) any {
	t.Helper()
	content, err := os.ReadFile(specPath(relativePath))
	if err != nil {
		t.Fatal(err)
	}
	return decodeJSON(t, content)
}

func decodeJSON(t *testing.T, content []byte) any {
	t.Helper()
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		t.Fatal(err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		t.Fatal("JSON contains trailing data")
	}
	return value
}

func specPath(relativePath string) string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("locate contract package")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", "spec", relativePath))
}

func normalizeFixtureNumbers(value any) any {
	switch value := value.(type) {
	case json.Number:
		integerValue, err := value.Int64()
		if err != nil {
			panic(err)
		}
		return integerValue
	case []any:
		result := make([]any, len(value))
		for index, item := range value {
			result[index] = normalizeFixtureNumbers(item)
		}
		return result
	case map[string]any:
		result := make(map[string]any, len(value))
		for key, item := range value {
			result[key] = normalizeFixtureNumbers(item)
		}
		return result
	default:
		return value
	}
}

func semanticEqual(left, right any) bool {
	leftJSON, leftErr := json.Marshal(left)
	rightJSON, rightErr := json.Marshal(right)
	if leftErr != nil || rightErr != nil {
		return false
	}
	var leftValue, rightValue any
	leftDecoder := json.NewDecoder(bytes.NewReader(leftJSON))
	leftDecoder.UseNumber()
	rightDecoder := json.NewDecoder(bytes.NewReader(rightJSON))
	rightDecoder.UseNumber()
	if leftDecoder.Decode(&leftValue) != nil || rightDecoder.Decode(&rightValue) != nil {
		return false
	}
	return reflect.DeepEqual(leftValue, rightValue)
}

func errorCategory(t *testing.T, envelope any) string {
	t.Helper()
	root := object(t, envelope)
	errorValue := root["error"]
	if errorValue == nil {
		return ""
	}
	return stringValue(t, field(t, errorValue, "category"))
}

func object(t *testing.T, value any) map[string]any {
	t.Helper()
	result, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("value %#v is not an object", value)
	}
	return result
}

func array(t *testing.T, value any) []any {
	t.Helper()
	result, ok := value.([]any)
	if !ok {
		t.Fatalf("value %#v is not an array", value)
	}
	return result
}

func field(t *testing.T, value any, name string) any {
	t.Helper()
	result, exists := object(t, value)[name]
	if !exists {
		t.Fatalf("missing field %q in %#v", name, value)
	}
	return result
}

func stringValue(t *testing.T, value any) string {
	t.Helper()
	result, ok := value.(string)
	if !ok {
		t.Fatalf("value %#v is not a string", value)
	}
	return result
}

func integer(t *testing.T, value any) int64 {
	t.Helper()
	switch value := value.(type) {
	case json.Number:
		result, err := value.Int64()
		if err != nil {
			t.Fatal(err)
		}
		return result
	case int:
		return int64(value)
	case int64:
		return value
	default:
		t.Fatalf("value %#v is not an integer", value)
		return 0
	}
}

func requireKeys(t *testing.T, values map[string]any, allowed []string) {
	t.Helper()
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, key := range allowed {
		allowedSet[key] = struct{}{}
	}
	for key := range values {
		if _, ok := allowedSet[key]; !ok {
			t.Fatalf("unknown fixture key %q", key)
		}
	}
}

func requireExactKeys(t *testing.T, values map[string]any, expected []string) {
	t.Helper()
	if len(values) != len(expected) {
		t.Fatalf("object keys = %v, want %v", sortedKeys(values), expected)
	}
	requireKeys(t, values, expected)
}

func sortedKeys(values map[string]any) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func stringIndex(values []string, target string) int {
	for index, value := range values {
		if value == target {
			return index
		}
	}
	return -1
}

// waitForFile, waitForFileProcess, and signalFile implement the ready/release
// handshake used to coordinate helper subprocesses. Because helpers run as
// separate OS processes with no shared memory or channels, synchronization
// has to be visible externally: a ready file's existence is the helper
// proving (via the filesystem) that it has reached its blocking point —
// e.g., the SQLite lock is actually held, or the actor is parked just before
// its real exec — and a release file is the coordinator's one-shot signal to
// proceed. Coordinators always wait for every helper's ready file before
// creating a release file, which is what prevents start-order races: without
// this handshake a helper could still be initializing (or not yet have
// acquired its lock) when the "concurrent" operation it is supposed to race
// against begins.
func waitForFile(t *testing.T, path string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		if _, err := os.Stat(path); err == nil {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", path)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// waitForFileProcess is the helper-subprocess counterpart to waitForFile: it
// runs outside of any *testing.T, so on timeout it must report failure via
// stderr/os.Exit rather than t.Fatalf.
func waitForFileProcess(path string, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for {
		if _, err := os.Stat(path); err == nil {
			return
		}
		if time.Now().After(deadline) {
			fmt.Fprintln(os.Stderr, "timed out waiting for release")
			os.Exit(1)
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// signalFile creates path exactly once (O_EXCL), so a helper or coordinator
// accidentally signaling the same handshake step twice fails loudly instead
// of silently succeeding.
func signalFile(path string) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := file.Close(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func removeFile(t *testing.T, path string) {
	t.Helper()
	if path == "" {
		return
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
}

func requiredEnvironment(name string) string {
	value, ok := os.LookupEnv(name)
	if !ok {
		fmt.Fprintf(os.Stderr, "%s is required\n", name)
		os.Exit(1)
	}
	return value
}
