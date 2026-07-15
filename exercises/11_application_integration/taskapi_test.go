package taskapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestTaskValidate(t *testing.T) {
	now := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	past := now.Add(-time.Hour)
	future := now.Add(time.Hour)

	tests := []struct {
		name    string
		task    Task
		wantErr bool
	}{
		{"valid without due date", Task{Title: "Write report"}, false},
		{"valid with future due date", Task{Title: "Write report", DueDate: &future}, false},
		{"empty title", Task{Title: ""}, true},
		{"whitespace-only title", Task{Title: "   "}, true},
		{"title too long", Task{Title: strings.Repeat("a", maxTitleLen+1)}, true},
		{"title at max length", Task{Title: strings.Repeat("a", maxTitleLen)}, false},
		{"multibyte title counts code points", Task{Title: strings.Repeat("é", maxTitleLen)}, false},
		{"due date in the past", Task{Title: "Write report", DueDate: &past}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.Validate(now)
			if tt.wantErr && err == nil {
				t.Fatalf("Validate() = nil, want an error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("Validate() unexpected error: %v", err)
			}
		})
	}
}

func newTestStore(t *testing.T) *SQLTaskStore {
	t.Helper()
	db, err := OpenFakeDB(t.Name())
	if err != nil {
		t.Fatalf("OpenFakeDB: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return NewSQLTaskStore(db)
}

func TestSQLTaskStoreCreateAndGet(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()
	due := time.Date(2030, time.June, 1, 12, 0, 0, 0, time.UTC)

	created, err := store.Create(ctx, Task{Title: "Ship it", DueDate: &due})
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}
	if created.ID == 0 {
		t.Fatalf("Create() did not assign an ID: %+v", created)
	}

	got, err := store.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("Get(%d) error: %v", created.ID, err)
	}
	if got.Title != "Ship it" || got.Done {
		t.Errorf("Get(%d) = %+v, want Title=%q Done=false", created.ID, got, "Ship it")
	}
	if got.DueDate == nil || !got.DueDate.Equal(due) {
		t.Errorf("Get(%d).DueDate = %v, want %v", created.ID, got.DueDate, due)
	}
}

func TestSQLTaskStoreCreateWithoutDueDate(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	created, err := store.Create(ctx, Task{Title: "No due date"})
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}
	got, err := store.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("Get(%d) error: %v", created.ID, err)
	}
	if got.DueDate != nil {
		t.Errorf("Get(%d).DueDate = %v, want nil", created.ID, got.DueDate)
	}
}

func TestSQLTaskStoreGetNotFound(t *testing.T) {
	store := newTestStore(t)
	_, err := store.Get(context.Background(), 12345)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get(12345) error = %v, want ErrNotFound", err)
	}
}

func TestSQLTaskStoreList(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	for _, title := range []string{"first", "second", "third"} {
		if _, err := store.Create(ctx, Task{Title: title}); err != nil {
			t.Fatalf("Create(%q) error: %v", title, err)
		}
	}

	tasks, err := store.List(ctx)
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(tasks) != 3 {
		t.Fatalf("List() returned %d tasks, want 3", len(tasks))
	}
	for i := 1; i < len(tasks); i++ {
		if tasks[i].ID <= tasks[i-1].ID {
			t.Errorf("List() not ordered by ID ascending: %+v", tasks)
		}
	}
}

func TestCreateTaskHandlerValidation(t *testing.T) {
	store := newTestStore(t)
	server := httptest.NewServer(NewServer(store))
	defer server.Close()

	resp, err := http.Post(server.URL+"/tasks", "application/json", strings.NewReader(`{"title":""}`))
	if err != nil {
		t.Fatalf("POST /tasks: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decoding error body: %v", err)
	}
	if body["error"] == "" {
		t.Errorf("error body = %v, want a non-empty \"error\" field", body)
	}
}

func TestCreateAndGetTaskHandlers(t *testing.T) {
	store := newTestStore(t)
	server := httptest.NewServer(NewServer(store))
	defer server.Close()

	reqBody := `{"title":"Write tests"}`
	resp, err := http.Post(server.URL+"/tasks", "application/json", strings.NewReader(reqBody))
	if err != nil {
		t.Fatalf("POST /tasks: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusCreated)
	}
	var created Task
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatalf("decoding created task: %v", err)
	}
	if created.ID == 0 || created.Title != "Write tests" {
		t.Fatalf("created task = %+v, want a non-zero ID and Title=%q", created, "Write tests")
	}

	getResp, err := http.Get(server.URL + "/tasks/" + strconv.FormatInt(created.ID, 10))
	if err != nil {
		t.Fatalf("GET /tasks/%d: %v", created.ID, err)
	}
	defer getResp.Body.Close()
	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", getResp.StatusCode, http.StatusOK)
	}
}

func TestGetTaskHandlerNotFound(t *testing.T) {
	store := newTestStore(t)
	server := httptest.NewServer(NewServer(store))
	defer server.Close()

	resp, err := http.Get(server.URL + "/tasks/999999")
	if err != nil {
		t.Fatalf("GET /tasks/999999: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestWithTimeoutShortensContextDeadline(t *testing.T) {
	var sawErr error
	slow := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(200 * time.Millisecond):
		case <-r.Context().Done():
			sawErr = r.Context().Err()
		}
		w.WriteHeader(http.StatusOK)
	})

	wrapped := WithTimeout(slow, 20*time.Millisecond)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if sawErr == nil {
		t.Fatal("handler's context was not canceled before the slow work finished; WithTimeout did not shorten the deadline")
	}
}

func TestWithLoggingRecordsStatus(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	wrapped := WithLogging(handler, logger)

	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	out := buf.String()
	if !strings.Contains(out, "GET") || !strings.Contains(out, "/tasks") {
		t.Errorf("log output = %q, want it to mention method GET and path /tasks", out)
	}
	if !strings.Contains(out, "418") {
		t.Errorf("log output = %q, want it to mention status 418", out)
	}
}

func TestClientCreateAndGetTask(t *testing.T) {
	store := newTestStore(t)
	server := httptest.NewServer(NewServer(store))
	defer server.Close()

	client := &Client{BaseURL: server.URL}
	created, err := client.CreateTask(context.Background(), Task{Title: "Client task"})
	if err != nil {
		t.Fatalf("CreateTask() error: %v", err)
	}
	if created.ID == 0 || created.Title != "Client task" {
		t.Fatalf("CreateTask() = %+v, want a non-zero ID and Title=%q", created, "Client task")
	}

	got, err := client.GetTask(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("GetTask(%d) error: %v", created.ID, err)
	}
	if got.Title != "Client task" {
		t.Errorf("GetTask(%d) = %+v, want Title=%q", created.ID, got, "Client task")
	}
}

func TestClientGetTaskNotFound(t *testing.T) {
	store := newTestStore(t)
	server := httptest.NewServer(NewServer(store))
	defer server.Close()

	client := &Client{BaseURL: server.URL}
	_, err := client.GetTask(context.Background(), 999999)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetTask(999999) error = %v, want ErrNotFound", err)
	}
}

func TestClientCreateTaskContextTimeout(t *testing.T) {
	slow := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(Task{ID: 1, Title: "irrelevant"})
	})
	server := httptest.NewServer(slow)
	defer server.Close()

	client := &Client{BaseURL: server.URL}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	_, err := client.CreateTask(ctx, Task{Title: "too slow"})
	if err == nil {
		t.Fatal("CreateTask() = nil error, want a context deadline error")
	}
}
