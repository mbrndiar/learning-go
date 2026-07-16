package m1

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

type Task struct {
	ID        int64
	Title     string
	Completed bool
}

type CreateInput struct {
	Title string
}

type UpdateInput struct {
	Title     *string
	Completed *bool
}

type ListFilter struct {
	Completed *bool
}

type Repository interface {
	Create(context.Context, CreateInput) (Task, error)
	List(context.Context, ListFilter) ([]Task, error)
	Get(context.Context, int64) (Task, error)
	Update(context.Context, int64, UpdateInput) (Task, error)
	Delete(context.Context, int64) error
}

type Service interface {
	Create(context.Context, CreateInput) (Task, error)
	List(context.Context, ListFilter) ([]Task, error)
	Get(context.Context, int64) (Task, error)
	Update(context.Context, int64, UpdateInput) (Task, error)
	Delete(context.Context, int64) error
}

type TaskHarness struct {
	MaxTitleLength      int
	NormalizeTitle      func(string) (string, error)
	ValidateTitle       func(string) error
	ValidateID          func(int64) error
	NormalizeUpdate     func(UpdateInput) (UpdateInput, error)
	NormalizeListFilter func(ListFilter) (ListFilter, error)
	ValidateTask        func(Task) error
	NewService          func(Repository) Service
	IsValidation        func(error) bool
	ValidationDetails   func(error) (string, string, bool)
	NewNotFound         func(int64) error
	IsNotFound          func(error) bool
	NotFoundID          func(error) (int64, bool)
	WrapStorage         func(string, error) error
	IsStorage           func(error) bool
}

type Recorder struct {
	Calls       []string
	Context     context.Context
	CreateInput CreateInput
	ListFilter  ListFilter
	ID          int64
	UpdateInput UpdateInput
	TaskResult  Task
	ListResult  []Task
	Err         error
}

func (r *Recorder) Create(ctx context.Context, input CreateInput) (Task, error) {
	r.Calls = append(r.Calls, "create")
	r.Context = ctx
	r.CreateInput = input
	return r.TaskResult, r.Err
}

func (r *Recorder) List(ctx context.Context, filter ListFilter) ([]Task, error) {
	r.Calls = append(r.Calls, "list")
	r.Context = ctx
	r.ListFilter = filter
	return r.ListResult, r.Err
}

func (r *Recorder) Get(ctx context.Context, id int64) (Task, error) {
	r.Calls = append(r.Calls, "get")
	r.Context = ctx
	r.ID = id
	return r.TaskResult, r.Err
}

func (r *Recorder) Update(ctx context.Context, id int64, input UpdateInput) (Task, error) {
	r.Calls = append(r.Calls, "update")
	r.Context = ctx
	r.ID = id
	r.UpdateInput = input
	return r.TaskResult, r.Err
}

func (r *Recorder) Delete(ctx context.Context, id int64) error {
	r.Calls = append(r.Calls, "delete")
	r.Context = ctx
	r.ID = id
	return r.Err
}

func AssertSolutionTask(t *testing.T, harness TaskHarness) {
	t.Helper()
	assertTitleRules(t, harness)
	assertValidationRules(t, harness)
	assertErrorCategories(t, harness)
	assertService(t, harness)
}

func assertTitleRules(t *testing.T, harness TaskHarness) {
	t.Helper()
	maxASCII := strings.Repeat("a", harness.MaxTitleLength)
	maxUnicode := strings.Repeat("界", harness.MaxTitleLength)

	tests := []struct {
		name    string
		input   string
		want    string
		message string
	}{
		{name: "trim", input: " \u2003 Learn REST \u00a0", want: "Learn REST"},
		{name: "trim surrounding tab", input: "\tLearn REST\t", want: "Learn REST"},
		{name: "trim surrounding carriage return", input: "\rLearn REST\r", want: "Learn REST"},
		{name: "trim surrounding line feed", input: "\nLearn REST\n", want: "Learn REST"},
		{name: "trim surrounding line separator", input: "\u2028Learn REST\u2028", want: "Learn REST"},
		{name: "trim surrounding paragraph separator", input: "\u2029Learn REST\u2029", want: "Learn REST"},
		{name: "unicode", input: "Καλημέρα 世界 🚀", want: "Καλημέρα 世界 🚀"},
		{name: "maximum ASCII", input: maxASCII, want: maxASCII},
		{name: "maximum Unicode", input: maxUnicode, want: maxUnicode},
		{name: "empty", input: " \u2003 ", message: "title must contain between 1 and 120 characters"},
		{name: "too long", input: maxUnicode + "界", message: "title must contain between 1 and 120 characters"},
		{name: "line feed", input: "one\ntwo", message: "title must occupy one physical line"},
		{name: "carriage return", input: "one\rtwo", message: "title must occupy one physical line"},
		{name: "vertical tab line break", input: "one\vtwo", message: "title must occupy one physical line"},
		{name: "form feed line break", input: "one\ftwo", message: "title must occupy one physical line"},
		{name: "file separator line break", input: "one\u001ctwo", message: "title must occupy one physical line"},
		{name: "group separator line break", input: "one\u001dtwo", message: "title must occupy one physical line"},
		{name: "record separator line break", input: "one\u001etwo", message: "title must occupy one physical line"},
		{name: "next line break", input: "one\u0085two", message: "title must occupy one physical line"},
		{name: "line separator", input: "one\u2028two", message: "title must occupy one physical line"},
		{name: "paragraph separator", input: "one\u2029two", message: "title must occupy one physical line"},
		{name: "tab", input: "one\ttwo", message: "title must not contain control characters"},
		{name: "nul", input: "one\x00two", message: "title must not contain control characters"},
		{name: "invalid UTF-8", input: string([]byte{0xff}), message: "title must contain valid UTF-8"},
	}

	for _, test := range tests {
		t.Run("title/"+test.name, func(t *testing.T) {
			got, err := harness.NormalizeTitle(test.input)
			if test.message == "" {
				if err != nil {
					t.Fatalf("NormalizeTitle() error = %v", err)
				}
				if got != test.want {
					t.Fatalf("NormalizeTitle() = %q, want %q", got, test.want)
				}
				return
			}
			assertValidation(t, harness, err, "title", test.message)
		})
	}

	if err := harness.ValidateTitle(" padded "); err == nil {
		t.Fatal("ValidateTitle() accepted surrounding whitespace")
	}
	if err := harness.ValidateTitle("valid"); err != nil {
		t.Fatalf("ValidateTitle(valid) error = %v", err)
	}
}

func assertValidationRules(t *testing.T, harness TaskHarness) {
	t.Helper()
	for _, id := range []int64{-1, 0} {
		t.Run(fmt.Sprintf("id/%d", id), func(t *testing.T) {
			assertValidation(t, harness, harness.ValidateID(id), "id", "task ID must be a positive integer")
		})
	}
	if err := harness.ValidateID(1); err != nil {
		t.Fatalf("ValidateID(1) error = %v", err)
	}

	t.Run("both update fields absent", func(t *testing.T) {
		_, err := harness.NormalizeUpdate(UpdateInput{})
		assertValidation(t, harness, err, "update", "update must include title or completed")
	})

	t.Run("update normalization and false", func(t *testing.T) {
		title := "  updated  "
		completed := false
		got, err := harness.NormalizeUpdate(UpdateInput{Title: &title, Completed: &completed})
		if err != nil {
			t.Fatalf("NormalizeUpdate() error = %v", err)
		}
		if got.Title == nil || *got.Title != "updated" {
			t.Fatalf("normalized title = %#v", got.Title)
		}
		if got.Completed == nil || *got.Completed {
			t.Fatalf("normalized completed = %#v, want explicit false", got.Completed)
		}
	})

	t.Run("optional filters", func(t *testing.T) {
		got, err := harness.NormalizeListFilter(ListFilter{})
		if err != nil || got.Completed != nil {
			t.Fatalf("nil filter = %#v, %v", got, err)
		}
		completed := false
		got, err = harness.NormalizeListFilter(ListFilter{Completed: &completed})
		if err != nil || got.Completed == nil || *got.Completed {
			t.Fatalf("false filter = %#v, %v", got, err)
		}
	})

	t.Run("task validation", func(t *testing.T) {
		if err := harness.ValidateTask(Task{ID: 1, Title: "valid"}); err != nil {
			t.Fatalf("ValidateTask(valid) error = %v", err)
		}
		assertValidation(t, harness, harness.ValidateTask(Task{ID: 0, Title: "valid"}), "id", "task ID must be a positive integer")
		if err := harness.ValidateTask(Task{ID: 1, Title: " padded "}); err == nil {
			t.Fatal("ValidateTask() accepted an untrimmed title")
		}
	})
}

func assertErrorCategories(t *testing.T, harness TaskHarness) {
	t.Helper()
	t.Run("not found", func(t *testing.T) {
		err := harness.NewNotFound(42)
		if !harness.IsNotFound(fmt.Errorf("get: %w", err)) {
			t.Fatal("wrapped not-found error was not classified")
		}
		id, ok := harness.NotFoundID(err)
		if !ok || id != 42 {
			t.Fatalf("not-found ID = %d, %v", id, ok)
		}
		if err.Error() != "task 42 was not found" {
			t.Fatalf("not-found message = %q", err)
		}
	})

	t.Run("storage", func(t *testing.T) {
		cause := errors.New("disk unavailable")
		err := harness.WrapStorage("list", cause)
		if !harness.IsStorage(fmt.Errorf("service: %w", err)) {
			t.Fatal("wrapped storage error was not classified")
		}
		if !errors.Is(err, cause) {
			t.Fatal("storage error did not preserve its cause")
		}
		if harness.WrapStorage("list", nil) != nil {
			t.Fatal("WrapStorage(nil) did not return nil")
		}
	})
}

func assertService(t *testing.T, harness TaskHarness) {
	t.Helper()
	result := Task{ID: 7, Title: "stored", Completed: true}
	listResult := []Task{result}

	t.Run("create", func(t *testing.T) {
		recorder := &Recorder{TaskResult: result}
		service := harness.NewService(recorder)
		ctx := context.WithValue(context.Background(), contextKey{}, "create")
		got, err := service.Create(ctx, CreateInput{Title: "  stored  "})
		if err != nil || got != result {
			t.Fatalf("Create() = %#v, %v", got, err)
		}
		if recorder.CreateInput.Title != "stored" || recorder.Context != ctx {
			t.Fatalf("Create repository arguments = %#v, %v", recorder.CreateInput, recorder.Context)
		}
	})

	t.Run("list explicit false", func(t *testing.T) {
		recorder := &Recorder{ListResult: listResult}
		service := harness.NewService(recorder)
		completed := false
		got, err := service.List(context.Background(), ListFilter{Completed: &completed})
		if err != nil || len(got) != 1 || got[0] != result {
			t.Fatalf("List() = %#v, %v", got, err)
		}
		if recorder.ListFilter.Completed == nil || *recorder.ListFilter.Completed {
			t.Fatalf("repository filter = %#v", recorder.ListFilter)
		}
	})

	t.Run("get update delete", func(t *testing.T) {
		recorder := &Recorder{TaskResult: result}
		service := harness.NewService(recorder)
		if got, err := service.Get(context.Background(), 7); err != nil || got != result || recorder.ID != 7 {
			t.Fatalf("Get() = %#v, %v; ID %d", got, err, recorder.ID)
		}
		title := " next "
		completed := false
		if got, err := service.Update(context.Background(), 7, UpdateInput{Title: &title, Completed: &completed}); err != nil || got != result {
			t.Fatalf("Update() = %#v, %v", got, err)
		}
		if recorder.UpdateInput.Title == nil || *recorder.UpdateInput.Title != "next" ||
			recorder.UpdateInput.Completed == nil || *recorder.UpdateInput.Completed {
			t.Fatalf("Update repository arguments = %#v", recorder.UpdateInput)
		}
		if err := service.Delete(context.Background(), 7); err != nil || recorder.ID != 7 {
			t.Fatalf("Delete() error = %v; ID %d", err, recorder.ID)
		}
	})

	t.Run("validation prevents repository calls", func(t *testing.T) {
		tests := []struct {
			name string
			call func(Service) error
		}{
			{name: "create", call: func(service Service) error {
				_, err := service.Create(context.Background(), CreateInput{})
				return err
			}},
			{name: "get", call: func(service Service) error {
				_, err := service.Get(context.Background(), 0)
				return err
			}},
			{name: "update ID", call: func(service Service) error {
				completed := true
				_, err := service.Update(context.Background(), 0, UpdateInput{Completed: &completed})
				return err
			}},
			{name: "both update fields absent", call: func(service Service) error {
				_, err := service.Update(context.Background(), 1, UpdateInput{})
				return err
			}},
			{name: "delete", call: func(service Service) error {
				return service.Delete(context.Background(), 0)
			}},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				recorder := &Recorder{}
				err := test.call(harness.NewService(recorder))
				if !harness.IsValidation(err) {
					t.Fatalf("error = %v, want validation error", err)
				}
				if len(recorder.Calls) != 0 {
					t.Fatalf("repository calls = %v", recorder.Calls)
				}
			})
		}
	})

	t.Run("repository errors and canceled context", func(t *testing.T) {
		recorder := &Recorder{Err: context.Canceled}
		service := harness.NewService(recorder)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := service.Get(ctx, 1)
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Get() error = %v, want context.Canceled", err)
		}
		if recorder.Context != ctx || !errors.Is(recorder.Context.Err(), context.Canceled) {
			t.Fatal("canceled context was not forwarded unchanged")
		}
	})

	t.Run("all repository errors are unchanged", func(t *testing.T) {
		repositoryErr := errors.New("repository error")
		tests := []struct {
			name string
			call func(Service) error
		}{
			{name: "create", call: func(service Service) error {
				_, err := service.Create(context.Background(), CreateInput{Title: "valid"})
				return err
			}},
			{name: "list", call: func(service Service) error {
				_, err := service.List(context.Background(), ListFilter{})
				return err
			}},
			{name: "get", call: func(service Service) error {
				_, err := service.Get(context.Background(), 1)
				return err
			}},
			{name: "update", call: func(service Service) error {
				completed := true
				_, err := service.Update(context.Background(), 1, UpdateInput{Completed: &completed})
				return err
			}},
			{name: "delete", call: func(service Service) error {
				return service.Delete(context.Background(), 1)
			}},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				recorder := &Recorder{Err: repositoryErr}
				if err := test.call(harness.NewService(recorder)); err != repositoryErr {
					t.Fatalf("error = %v, want original repository error", err)
				}
				if len(recorder.Calls) != 1 {
					t.Fatalf("repository calls = %v", recorder.Calls)
				}
			})
		}
	})
}

func assertValidation(t *testing.T, harness TaskHarness, err error, field, message string) {
	t.Helper()
	if err == nil {
		t.Fatalf("error = nil, want validation error for %s", field)
	}
	if !harness.IsValidation(fmt.Errorf("wrapped: %w", err)) {
		t.Fatalf("error %v was not classified as validation", err)
	}
	gotField, gotMessage, ok := harness.ValidationDetails(err)
	if !ok || gotField != field || gotMessage != message || err.Error() != message {
		t.Fatalf("validation = (%q, %q, %v), error %q; want (%q, %q, true)",
			gotField, gotMessage, ok, err, field, message)
	}
}

type ClientHarness struct {
	DefaultTimeout       time.Duration
	NormalizeBaseURL     func(string) (string, error)
	ValidateConfig       func(string, time.Duration) (string, time.Duration, error)
	IsInvalidConfig      func(error) bool
	ConfigDetails        func(error) (string, string, bool)
	NewAPIError          func() error
	IsAPI                func(error) bool
	NewResponseError     func(error) error
	IsUnexpectedResponse func(error) bool
	NewConnectionError   func(error) error
	IsConnection         func(error) bool
}

func AssertSolutionClient(t *testing.T, harness ClientHarness) {
	t.Helper()
	if harness.DefaultTimeout != 5*time.Second {
		t.Fatalf("DefaultTimeout = %v", harness.DefaultTimeout)
	}

	urlTests := []struct {
		input string
		want  string
	}{
		{input: " http://127.0.0.1:8000/ ", want: "http://127.0.0.1:8000"},
		{input: "https://example.test/api///", want: "https://example.test/api"},
	}
	for _, test := range urlTests {
		got, err := harness.NormalizeBaseURL(test.input)
		if err != nil || got != test.want {
			t.Fatalf("NormalizeBaseURL(%q) = %q, %v; want %q", test.input, got, err, test.want)
		}
	}
	for _, value := range []string{"", "example.test", "ftp://example.test", "http:///missing", "http://user@example.test", "http://example.test?q=1", "http://example.test/#fragment"} {
		_, err := harness.NormalizeBaseURL(value)
		assertConfigError(t, harness, err, "base-url", "base URL must be an absolute HTTP URL")
	}

	baseURL, timeout, err := harness.ValidateConfig("http://example.test/", time.Second)
	if err != nil || baseURL != "http://example.test" || timeout != time.Second {
		t.Fatalf("ValidateConfig() = %q, %v, %v", baseURL, timeout, err)
	}
	for _, timeout := range []time.Duration{0, -time.Second} {
		_, _, err := harness.ValidateConfig("http://example.test", timeout)
		assertConfigError(t, harness, err, "timeout", "timeout must be positive and finite")
	}

	cause := errors.New("cause")
	categories := []struct {
		name string
		err  error
		is   func(error) bool
	}{
		{name: "API", err: harness.NewAPIError(), is: harness.IsAPI},
		{name: "response", err: harness.NewResponseError(cause), is: harness.IsUnexpectedResponse},
		{name: "connection", err: harness.NewConnectionError(cause), is: harness.IsConnection},
	}
	for _, category := range categories {
		t.Run(category.name, func(t *testing.T) {
			if !category.is(fmt.Errorf("wrapped: %w", category.err)) {
				t.Fatalf("error %v was not classified", category.err)
			}
		})
	}
	if !errors.Is(harness.NewResponseError(cause), cause) {
		t.Fatal("response error did not preserve cause")
	}
	if !errors.Is(harness.NewConnectionError(cause), cause) {
		t.Fatal("connection error did not preserve cause")
	}
}

func assertConfigError(t *testing.T, harness ClientHarness, err error, field, message string) {
	t.Helper()
	if err == nil || !harness.IsInvalidConfig(fmt.Errorf("wrapped: %w", err)) {
		t.Fatalf("error = %v, want invalid configuration", err)
	}
	gotField, gotMessage, ok := harness.ConfigDetails(err)
	if !ok || gotField != field || gotMessage != message {
		t.Fatalf("configuration error = (%q, %q, %v)", gotField, gotMessage, ok)
	}
}

func AssertStarterExplicit(t *testing.T, notImplemented error, calls func() int, operations []func() error) {
	t.Helper()
	for index, operation := range operations {
		err := operation()
		if !errors.Is(err, notImplemented) {
			t.Fatalf("operation %d error = %v, want ErrNotImplemented", index, err)
		}
	}
	if got := calls(); got != 0 {
		t.Fatalf("dependency calls = %d, want 0", got)
	}
}

type contextKey struct{}
