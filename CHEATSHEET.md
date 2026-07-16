# 🐹 Go Cheat Sheet

## Values and declarations

```go
var count int
name := "Ada"
const maxRetries = 3
```

Zero values: numbers `0`, booleans `false`, strings `""`, and pointers,
slices, maps, channels, functions, and interfaces `nil`.

## Control flow

```go
if value, err := parse(input); err != nil {
    return err
} else {
    fmt.Println(value)
}

for index, value := range values {
    fmt.Println(index, value)
}

switch status {
case "ready":
    fmt.Println("go")
default:
    fmt.Println("wait")
}
```

## Functions, pointers, and errors

```go
func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}

func increment(value *int) {
    (*value)++
}
```

Wrap errors with context:

```go
return fmt.Errorf("load config: %w", err)
```

Use `errors.Is` and `errors.As` instead of comparing error strings.

## Slices and maps

```go
numbers := []int{1, 2, 3}
numbers = append(numbers, 4)
copyOfNumbers := slices.Clone(numbers)

scores := map[string]int{"Ada": 10}
score, found := scores["Grace"]
delete(scores, "Ada")
```

## Structs, methods, and interfaces

```go
type Task struct {
    ID    int    `json:"id"`
    Title string `json:"title"`
    Done  bool   `json:"done"`
}

func (task *Task) Complete() {
    task.Done = true
}

type Storage interface {
    Add(context.Context, string) (Task, error)
}
```

Interfaces are satisfied implicitly. Keep them small and define them where they
are consumed.

## Generics

```go
func Contains[T comparable](values []T, target T) bool {
    for _, value := range values {
        if value == target {
            return true
        }
    }
    return false
}
```

## Testing

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name string
        a, b int
        want int
    }{
        {"positive", 2, 3, 5},
        {"zero", 0, 0, 0},
    }

    for _, test := range tests {
        t.Run(test.name, func(t *testing.T) {
            if got := Add(test.a, test.b); got != test.want {
                t.Fatalf("Add() = %d, want %d", got, test.want)
            }
        })
    }
}
```

## Concurrency

```go
results := make(chan int)
go func() {
    defer close(results)
    results <- work()
}()

select {
case result := <-results:
    fmt.Println(result)
case <-ctx.Done():
    return ctx.Err()
}
```

Only the sender closes a channel. Every goroutine should have a clear owner and
termination path.

## SQL and SQLite

```go
row := db.QueryRowContext(ctx,
    "SELECT id, title FROM tasks WHERE id = ?", id)
if err := row.Scan(&task.ID, &task.Title); err != nil {
    return fmt.Errorf("find task: %w", err)
}
```

Pass values as query arguments, close multi-row results, check `rows.Err()`,
and `defer tx.Rollback()` immediately after `BeginTx`. SQLite `PRAGMA` settings
and `:memory:` connection behavior are database-specific.

## HTTP and JSON

```go
mux := http.NewServeMux()
mux.HandleFunc("GET /tasks/{id}", handler)

encoder := json.NewEncoder(response)
decoder := json.NewDecoder(request.Body)
```

Always use finite client/server timeouts and validate decoded data.

## Essential commands

```bash
go run ./path/to/package
go test ./path/to/package
go test -race ./path/to/package
go test -coverprofile=coverage.out ./path/to/package
go tool cover -func=coverage.out
go vet ./...
gofmt -w .
go mod tidy
go doc package.Symbol
```

At this repository root, `go test ./...` intentionally fails until every
exercise starter is implemented. CI instead compiles starters without running
their behavior tests and tests the completed course surfaces; see
[`docs/QUALITY.md`](docs/QUALITY.md).

For the required Tasks project, use its explicit starter/solution split:

```bash
go test -timeout=2m -count=1 ./projects/tasks/starter/...
go test -timeout=3m -count=1 \
  ./projects/tasks/solution/... \
  ./projects/tasks/tests/... \
  ./projects/tasks
```

## Further reading

- <https://go.dev/doc/>
- <https://go.dev/tour/>
- <https://go.dev/doc/effective_go>
- <https://go.dev/wiki/CodeReviewComments>
