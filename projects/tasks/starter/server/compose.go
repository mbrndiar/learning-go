package server

import (
	"context"
	"log/slog"

	"github.com/mbrndiar/learning-go/projects/tasks/starter/task"
)

// Run validates config, builds the selected HTTP server and storage backend,
// owns their cleanup, and serves until ctx is done. logger receives HTTP
// boundary diagnostics from the selected adapter.
func Run(ctx context.Context, config Config, logger *slog.Logger) error {
	return task.ErrNotImplemented
}
