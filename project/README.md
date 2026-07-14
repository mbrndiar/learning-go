# 🔗 Connected Task Projects

Three packages form one progressive capstone:

- [`taskapi/`](taskapi/) owns remote task data in SQLite and exposes JSON over
  HTTP.
- [`taskclient/`](taskclient/) provides a reusable typed client and standalone
  CLI.
- [`taskmanager/`](taskmanager/) owns the domain model and selects local JSON or
  REST storage.

```text
Task Manager CLI -> Manager -> Storage
                             |-> FileStorage -> tasks.json
                             `-> RESTStorage -> Client -> API -> SQLiteStore
```

The projects demonstrate structs, interfaces, dependency injection, explicit
errors, atomic file writes, JSON, HTTP, contexts, SQLite, testing, and
concurrency without hiding the fundamentals behind frameworks.

## 🚀 Try the capstone

```bash
# Local JSON backend
go run ./project/taskmanager/cmd/task-manager add "Local task"

# Keep this API running in one terminal
go run ./project/taskapi/cmd/task-api

# Then use the HTTP client from another terminal
go run ./project/taskclient/cmd/task-client add "Remote task"
```

Each project directory contains its own architecture, usage, testing, and
extension documentation.
