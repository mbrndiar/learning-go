// Package httptestlesson implements a tiny HTTP greeting API used to
// demonstrate net/http/httptest: testing handlers directly with
// httptest.NewRecorder, and testing full round trips with
// httptest.NewServer.
package httptestlesson

import (
	"encoding/json"
	"net/http"
)

// greetResponse is the JSON body returned by the greet handler.
type greetResponse struct {
	Message string `json:"message"`
}

// NewMux builds the application's routes. Returning an http.Handler (rather
// than starting a server) keeps the handler easy to test in isolation.
func NewMux() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /greet", handleGreet)
	return mux
}

func handleGreet(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "missing required query parameter: name", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(greetResponse{Message: "Hello, " + name + "!"})
}
