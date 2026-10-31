package handler_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"group-platform/examples/records/adapter"
	example "group-platform/examples/records/handler"
	"group-platform/examples/records/repository"
	"group-platform/examples/records/service"
	production "group-platform/internal/handler"
	health "group-platform/internal/service"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

func fixture() (http.Handler, *adapter.Memory) {
	m := adapter.New()
	return example.New(service.Service{Repository: m}, func(r *http.Request) (string, bool) { s := r.Header.Get("Test-Subject"); return s, s != "" }), m
}
func call(h http.Handler, method, path, subject, key, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	if subject != "" {
		r.Header.Set("Test-Subject", subject)
	}
	if key != "" {
		r.Header.Set("Idempotency-Key", key)
	}
	r.Header.Set("X-Request-ID", "untrusted")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}
func expect(t *testing.T, w *httptest.ResponseRecorder, status int, code string) repository.Record {
	t.Helper()
	var env struct {
		Code, Message string
		Data          repository.Record
		ID            string `json:"request_id"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatal(err)
	}
	if w.Code != status || env.Code != code || env.Message == "" || len(env.ID) != 32 || env.ID != w.Header().Get("X-Request-ID") {
		t.Fatalf("want %d %s got %d %s", status, code, w.Code, w.Body)
	}
	if status >= 400 {
		var e map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &e)
		if e["data"] != nil {
			t.Fatal("error contains data")
		}
	}
	return env.Data
}
func TestNormalCreateRead(t *testing.T) {
	h, m := fixture()
	r := expect(t, call(h, "POST", "/api/v1/records", "alice", "k", `{"title":"  hello  "}`), 201, "CREATED")
	if r.Title != "hello" || r.ID == "" || r.Subject != "" {
		t.Fatal(r)
	}
	got := expect(t, call(h, "GET", "/api/v1/records/"+r.ID, "alice", "", ""), 200, "OK")
	if got != r || m.Count() != 1 {
		t.Fatal("read mismatch")
	}
}
func TestFailureOwnership(t *testing.T) {
	h, _ := fixture()
	r := expect(t, call(h, "POST", "/api/v1/records", "alice", "k", `{"title":"secret title"}`), 201, "CREATED")
	expect(t, call(h, "GET", "/api/v1/records/"+r.ID, "bob", "", ""), 404, "NOT_FOUND")
	expect(t, call(h, "GET", "/api/v1/records/999", "alice", "", ""), 404, "NOT_FOUND")
}

type unavailable struct{ calls atomic.Int32 }

func (s *unavailable) Create(a, b, c string) (repository.Record, error) {
	s.calls.Add(1)
	return repository.Record{}, errors.New("private-storage-password")
}
func (s *unavailable) Get(id string) (repository.Record, error) {
	s.calls.Add(1)
	return repository.Record{}, errors.New("private-storage-password")
}
func TestFailureUnauthorized(t *testing.T) {
	store := &unavailable{}
	for _, identity := range []example.Identity{nil, func(*http.Request) (string, bool) { return "", false }, func(*http.Request) (string, bool) { return " ", true }, func(*http.Request) (string, bool) { return "alice", false }} {
		h := example.New(service.Service{Repository: store}, identity)
		for _, method := range []string{"GET", "POST"} {
			expect(t, call(h, method, "/api/v1/records/1", "", "k", `{"title":"x"}`), 401, "UNAUTHORIZED")
		}
	}
	if store.calls.Load() != 0 {
		t.Fatal("unauthorized accessed storage")
	}
}
func TestFailureStorage(t *testing.T) {
	store := &unavailable{}
	h := example.New(service.Service{Repository: store}, func(*http.Request) (string, bool) { return "alice", true })
	for _, c := range []struct{ method, path string }{{"POST", "/api/v1/records"}, {"GET", "/api/v1/records/1"}} {
		w := call(h, c.method, c.path, "alice", "k", `{"title":"x"}`)
		expect(t, w, 503, "DEPENDENCY_UNAVAILABLE")
		if strings.Contains(w.Body.String(), "private") {
			t.Fatal("storage details leaked")
		}
	}
	if store.calls.Load() != 2 {
		t.Fatal("expected repository calls")
	}
}
func TestBoundaryInvalidInput(t *testing.T) {
	h, m := fixture()
	bodies := []string{``, `null`, `{`, `{"title":" "}`, `{"title":42}`, `{"title":"x","extra":1}`, `{"title":"x"}{}`, `{"title":"x"} false`, fmt.Sprintf(`{"title":%q}`, strings.Repeat("界", 81)), `{"title":"x"}` + strings.Repeat(" ", 1024)}
	for _, body := range bodies {
		expect(t, call(h, "POST", "/api/v1/records", "alice", "k", body), 400, "INVALID_ARGUMENT")
	}
	for _, key := range []string{"", "bad key", strings.Repeat("x", 129), "界"} {
		expect(t, call(h, "POST", "/api/v1/records", "alice", key, `{"title":"x"}`), 400, "INVALID_ARGUMENT")
	}
	if m.Count() != 0 {
		t.Fatal("invalid input wrote storage")
	}
}
func TestBoundaryValidInput(t *testing.T) {
	h, m := fixture()
	for i, title := range []string{"x", strings.Repeat("界", 80), strings.Repeat("😀", 80)} {
		body, _ := json.Marshal(map[string]string{"title": title})
		expect(t, call(h, "POST", "/api/v1/records", "alice", fmt.Sprint(i), string(body)), 201, "CREATED")
	}
	body := `{"title":"x"}`
	body += strings.Repeat(" ", 1024-len(body))
	expect(t, call(h, "POST", "/api/v1/records", "alice", strings.Repeat("k", 128), body), 201, "CREATED")
	if m.Count() != 4 {
		t.Fatal("valid boundaries rejected")
	}
}
func TestBoundaryRoutes(t *testing.T) {
	store := &unavailable{}
	h := example.New(service.Service{Repository: store}, func(*http.Request) (string, bool) { return "alice", true })
	for _, c := range []struct {
		method, path, code string
		status             int
	}{{"GET", "/api/v1/records", "METHOD_NOT_ALLOWED", 405}, {"DELETE", "/api/v1/records/1", "METHOD_NOT_ALLOWED", 405}, {"POST", "/api/v1/records?x=y", "INVALID_ARGUMENT", 400}, {"GET", "/api/v1/records/1?", "INVALID_ARGUMENT", 400}, {"GET", "/unknown", "NOT_FOUND", 404}, {"GET", "/api/v1/records/1/extra", "NOT_FOUND", 404}} {
		expect(t, call(h, c.method, c.path, "alice", "k", `{"title":"x"}`), c.status, c.code)
	}
	if store.calls.Load() != 0 {
		t.Fatal("invalid route accessed storage")
	}
}
func TestNormalConcurrentIdempotency(t *testing.T) {
	h, m := fixture()
	const n = 32
	var wg sync.WaitGroup
	results := make(chan *httptest.ResponseRecorder, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- call(h, "POST", "/api/v1/records", "alice", "same", `{"title":"title"}`)
		}()
	}
	wg.Wait()
	close(results)
	id := ""
	for w := range results {
		r := expect(t, w, 201, "CREATED")
		if id != "" && r.ID != id {
			t.Fatal("duplicate record")
		}
		id = r.ID
	}
	if m.Count() != 1 {
		t.Fatal("duplicate storage")
	}
	expect(t, call(h, "POST", "/api/v1/records", "alice", "same", `{"title":"other"}`), 409, "CONFLICT")
	r := expect(t, call(h, "POST", "/api/v1/records", "bob", "same", `{"title":"other"}`), 201, "CREATED")
	if r.ID == id || m.Count() != 2 {
		t.Fatal("subject key isolation failed")
	}
}
func TestFailureProductionIsolation(t *testing.T) {
	h := production.New(health.Service{}, slog.New(slog.NewJSONHandler(io.Discard, nil)))
	for _, method := range []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"} {
		expect(t, call(h, method, "/api/v1/records", "alice", "k", `{"title":"x"}`), 401, "UNAUTHORIZED")
	}
}
