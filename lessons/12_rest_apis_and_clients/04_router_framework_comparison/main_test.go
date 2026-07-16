package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEachRouterReadsTheSamePathParameter(t *testing.T) {
	routers := map[string]http.Handler{
		"net/http": standardLibraryRouter(),
		"Chi":      chiRouter(),
		"Gin":      ginRouter(),
	}
	for name, handler := range routers {
		t.Run(name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/items/12", nil))
			if recorder.Code != http.StatusOK || recorder.Body.String() != "12" {
				t.Fatalf("response = %d %q", recorder.Code, recorder.Body.String())
			}
		})
	}

	if rows := comparisons(); len(rows) != 3 {
		t.Fatalf("comparisons = %+v", rows)
	}
}
