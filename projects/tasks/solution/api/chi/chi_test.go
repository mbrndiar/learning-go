package chi_test

import (
	"log/slog"
	"net/http"
	"testing"

	"github.com/mbrndiar/learning-go/projects/tasks/solution/api"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/api/chi"
	"github.com/mbrndiar/learning-go/projects/tasks/tests/m5"
)

func TestMilestone5HTTPContract(t *testing.T) {
	m5.AssertServerContract(t, func(service api.Service, logger *slog.Logger) http.Handler {
		return chi.New(service, logger)
	})
}
