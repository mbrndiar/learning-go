package kvstore_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/mbrndiar/learning-go/capstones/comparative/starter/kvstore/app"
	"github.com/mbrndiar/learning-go/capstones/comparative/starter/kvstore/domain"
	"github.com/mbrndiar/learning-go/capstones/comparative/starter/kvstore/storage"
	"github.com/mbrndiar/learning-go/capstones/comparative/tests/contract"
)

func TestHarness(t *testing.T) {
	opener := storage.NewSQLiteOpener()
	contract.RunHarness(t, contract.Harness{
		Name:        "starter",
		Implemented: domain.Implemented,
		ParseValue: func() error {
			_, err := domain.ParseValue(json.RawMessage(`null`))
			return err
		},
		OpenStore: func(ctx context.Context) error {
			_, err := opener.Open(ctx, "unused.db")
			return err
		},
		IsIncomplete: func(err error) bool {
			return errors.Is(err, domain.ErrNotImplemented)
		},
		RunCLI: func(ctx context.Context, stdout, stderr *bytes.Buffer) int {
			return app.Run(ctx, []string{"--db", "unused.db", "list"}, stdout, stderr)
		},
		Placeholder:   `{"ok":false,"error":{"category":"not_implemented","details":{}}}` + "\n",
		PlaceholderRC: 1,
	})
}
