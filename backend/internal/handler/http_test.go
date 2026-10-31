package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"group-platform/internal/service"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type probe func(context.Context) error

func (p probe) Check(c context.Context) error { return p(c) }
func setup(mysql, qdrant probe) (http.Handler, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	svc := service.Service{Version: "test-v1", MySQL: mysql, Qdrant: qdrant, Now: func() time.Time { return time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC) }}
	return New(svc, slog.New(slog.NewJSONHandler(buf, nil))), buf
}

var up = probe(func(context.Context) error { return nil })

func request(h http.Handler, method, path string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, nil)
	r.Header.Set("X-Request-ID", "untrusted-secret")
	r.Header.Set("Authorization", "Bearer secret-value")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}
func TestNormalHealthAndLog(t *testing.T) {
	h, b := setup(up, up)
	w := request(h, "GET", "/health")
	var body struct {
		Code string         `json:"code"`
		Data service.Health `json:"data"`
		ID   string         `json:"request_id"`
	}
	if e := json.Unmarshal(w.Body.Bytes(), &body); e != nil {
		t.Fatal(e)
	}
	if w.Code != 200 || body.Code != "OK" || body.Data.Version != "test-v1" || body.Data.Time != "2026-10-31T00:00:00Z" || body.Data.Dependencies["mysql"] != "up" || body.Data.Dependencies["qdrant"] != "up" || len(body.ID) != 32 {
		t.Fatalf("unexpected response %s", w.Body)
	}
	var log map[string]any
	if e := json.Unmarshal(b.Bytes(), &log); e != nil {
		t.Fatal(e)
	}
	if log["request_id"] != body.ID || log["module"] != "http" || log["status"] != float64(200) {
		t.Fatal("log correlation failed")
	}
	for _, k := range []string{"time", "level", "duration_ms"} {
		if _, ok := log[k]; !ok {
			t.Fatal("missing log field", k)
		}
	}
}
func TestFailureRoutes(t *testing.T) {
	h, b := setup(up, up)
	for _, c := range []struct {
		method, path, code string
		status             int
	}{{"GET", "/api/v1/resources", "UNAUTHORIZED", 401}, {"POST", "/api/v1/resources", "UNAUTHORIZED", 401}, {"GET", "/unknown", "NOT_FOUND", 404}, {"POST", "/health", "METHOD_NOT_ALLOWED", 405}, {"GET", "/health?secret=secret-value", "INVALID_ARGUMENT", 400}} {
		w := request(h, c.method, c.path)
		var env Envelope
		_ = json.Unmarshal(w.Body.Bytes(), &env)
		if w.Code != c.status || env.Code != c.code || env.Data != nil {
			t.Errorf("%s %s: %d %s", c.method, c.path, w.Code, w.Body)
		}
	}
	if strings.Contains(b.String(), "secret") {
		t.Fatal("request secret logged")
	}
}
func TestFailureDependencyAndLiveness(t *testing.T) {
	down := probe(func(context.Context) error { return errors.New("password=private-driver-error") })
	for _, p := range []struct{ a, b probe }{{down, up}, {up, down}, {down, down}} {
		h, buf := setup(p.a, p.b)
		w := request(h, "GET", "/health")
		if w.Code != 503 || !strings.Contains(w.Body.String(), "DEPENDENCY_UNAVAILABLE") || strings.Contains(w.Body.String(), "private") {
			t.Fatal("failure contract")
		}
		if request(h, "GET", "/live").Code != 200 {
			t.Fatal("dependency affected liveness")
		}
		if strings.Contains(buf.String(), "private") {
			t.Fatal("driver error leaked")
		}
	}
}
func TestBoundaryDeadline(t *testing.T) {
	slow := probe(func(c context.Context) error { <-c.Done(); return c.Err() })
	h, _ := setup(slow, slow)
	start := time.Now()
	w := request(h, "GET", "/health")
	if w.Code != 503 || time.Since(start) > 3*time.Second {
		t.Fatal("probe deadline not enforced")
	}
}
