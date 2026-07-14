// Command 01_http_routing_pathvalue shows net/http's method-aware routing
// patterns ("GET /items/{id}") and how to read path segments with
// (*http.Request).PathValue, without any router dependency beyond the
// standard library.
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
)

// item is the resource this lesson's routes expose. JSON encoding details
// (validation, unknown fields, error envelopes) are covered in the next
// lesson; here the focus is purely on routing and PathValue.
type item struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// store is a tiny in-memory collection guarded by a Mutex, since multiple
// requests can arrive concurrently on separate goroutines (net/http runs
// each request's handler in its own goroutine).
type store struct {
	mu     sync.Mutex
	nextID int
	items  map[int]item
}

func newStore() *store {
	return &store{nextID: 1, items: make(map[int]item)}
}

func (s *store) add(name string) item {
	s.mu.Lock()
	defer s.mu.Unlock()

	it := item{ID: s.nextID, Name: name}
	s.items[it.ID] = it
	s.nextID++
	return it
}

func (s *store) get(id int) (item, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	it, ok := s.items[id]
	return it, ok
}

func (s *store) update(id int, name string) (item, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	it, ok := s.items[id]
	if !ok {
		return item{}, false
	}
	it.Name = name
	s.items[id] = it
	return it, true
}

func (s *store) delete(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.items[id]; !ok {
		return false
	}
	delete(s.items, id)
	return true
}

func (s *store) list() []item {
	s.mu.Lock()
	defer s.mu.Unlock()

	all := make([]item, 0, len(s.items))
	for _, it := range s.items {
		all = append(all, it)
	}
	return all
}

// pathID extracts and parses the {id} path segment. r.PathValue reads the
// segment net/http's router already matched against the route pattern, so
// no manual string splitting of r.URL.Path is needed.
func pathID(r *http.Request) (int, error) {
	return strconv.Atoi(r.PathValue("id"))
}

// newMux wires up method-aware routes. The "METHOD /path" pattern syntax
// (added to net/http in Go 1.22) means a GET and a POST to the same path
// can be routed to two different handlers without any manual method
// switching inside a single handler.
func newMux(s *store) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /items", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(s.list())
	})

	mux.HandleFunc("POST /items", func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		created := s.add(body.Name)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(created)
	})

	mux.HandleFunc("GET /items/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := pathID(r)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		it, ok := s.get(id)
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(it)
	})

	mux.HandleFunc("PUT /items/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := pathID(r)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		var body struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		it, ok := s.update(id, body.Name)
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(it)
	})

	mux.HandleFunc("DELETE /items/{id}", func(w http.ResponseWriter, r *http.Request) {
		id, err := pathID(r)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		if !s.delete(id) {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})

	return mux
}

func main() {
	mux := newMux(newStore())

	// httptest.NewServer starts a real local server on an ephemeral port,
	// which keeps this demo self-contained: no fixed port to collide with,
	// and it is always closed before main returns.
	server := httptest.NewServer(mux)
	defer server.Close()

	resp, err := http.Post(server.URL+"/items", "application/json", strings.NewReader(`{"name":"first"}`))
	if err != nil {
		panic(err)
	}
	fmt.Println("POST /items ->", resp.StatusCode)
	resp.Body.Close()

	resp, err = http.Get(server.URL + "/items/1")
	if err != nil {
		panic(err)
	}
	fmt.Println("GET /items/1 ->", resp.StatusCode)
	resp.Body.Close()

	resp, err = http.Get(server.URL + "/items/99")
	if err != nil {
		panic(err)
	}
	fmt.Println("GET /items/99 ->", resp.StatusCode)
	resp.Body.Close()
}
