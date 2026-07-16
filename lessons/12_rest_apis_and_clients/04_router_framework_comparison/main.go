// Command 04_router_framework_comparison implements the same route with
// net/http, Chi, and Gin so their APIs can be compared directly.
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/go-chi/chi/v5"
)

type RouterComparison struct {
	Name             string
	RouteExample     string
	PathParameterAPI string
	MiddlewareStyle  string
}

func comparisons() []RouterComparison {
	return []RouterComparison{
		{"net/http", `mux.HandleFunc("GET /items/{id}", handler)`, `r.PathValue("id")`, `func(http.Handler) http.Handler`},
		{"Chi", `r.Get("/items/{id}", handler)`, `chi.URLParam(r, "id")`, `func(http.Handler) http.Handler`},
		{"Gin", `r.GET("/items/:id", handler)`, `c.Param("id")`, `gin.HandlerFunc`},
	}
}

func standardLibraryRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /items/{id}", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, r.PathValue("id"))
	})
	return mux
}

func chiRouter() http.Handler {
	router := chi.NewRouter()
	router.Get("/items/{id}", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, chi.URLParam(r, "id"))
	})
	return router
}

func ginRouter() http.Handler {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.GET("/items/:id", func(c *gin.Context) {
		c.String(http.StatusOK, c.Param("id"))
	})
	return router
}

func main() {
	routers := []struct {
		name    string
		handler http.Handler
	}{
		{"net/http", standardLibraryRouter()},
		{"Chi", chiRouter()},
		{"Gin", ginRouter()},
	}
	for _, example := range routers {
		recorder := httptest.NewRecorder()
		example.handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/items/12", nil))
		fmt.Printf("%s path parameter: %s\n", example.name, recorder.Body.String())
	}
}
