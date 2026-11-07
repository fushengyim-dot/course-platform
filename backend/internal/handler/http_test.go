package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"group-platform/internal/adapter"
	"group-platform/internal/repository"
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

func TestPublicInformationRoutesAndDefaultDeny(t *testing.T){
	fixture:=adapter.NewSyntheticPublicInformation()
	item:=fixture.Items[0]
	info:=service.PublicInformation{Repository:fixture}
	directory:=service.PublicInstitutionDirectory{Repository:fixture}
	h:=NewWithPublicData(service.Service{Version:"test-v1",MySQL:up,Qdrant:up,Now:func()time.Time{return time.Now()}},info,directory,slog.New(slog.NewJSONHandler(&bytes.Buffer{},nil)))
	for _,path:=range []string{"/api/v1/public-information","/api/v1/public-information?college=计智学院","/api/v1/public-information/"+item.ID}{w:=request(h,"GET",path);if w.Code!=200||!strings.Contains(w.Body.String(),"虚构演示来源"){t.Errorf("public read %s: %d %s",path,w.Code,w.Body.String())}}
	if w:=request(h,"GET","/api/v1/public-information?institution=示例大学&college=电工系");w.Code!=200||strings.Contains(w.Body.String(),item.Title){t.Errorf("institution/college association: %d %s",w.Code,w.Body.String())}
	if w:=request(h,"GET","/api/v1/public-information?expected_publish_from=2026-11-12&expected_publish_to=2026-11-12");w.Code!=200||!strings.Contains(w.Body.String(),item.Title){t.Errorf("expected publication date: %d %s",w.Code,w.Body.String())}
	calendar:=request(h,"GET","/api/v1/public-information/calendar?month=2026-11")
	if calendar.Code!=200||!strings.Contains(calendar.Body.String(),"2026-11-12")||!strings.Contains(calendar.Body.String(),"example.invalid")||!strings.Contains(calendar.Body.String(),"东岭示范大学"){t.Errorf("calendar: %d %s",calendar.Code,calendar.Body.String())}
	for _,path:=range []string{"/api/v1/public-information/calendar?month=2026-13","/api/v1/public-information/calendar?month=2026-11&extra=x","/api/v1/public-information/calendar?month=2026-11&month=2026-12"}{if w:=request(h,"GET",path);w.Code!=400{t.Errorf("invalid calendar request %s: %d",path,w.Code)}}
	if w:=request(h,"POST","/api/v1/public-information/calendar?month=2026-11");w.Code!=405{t.Errorf("calendar method: %d",w.Code)}
	if w:=request(h,"GET","/api/v1/public-information/10000000-0000-4000-8000-000000000099");w.Code!=404{t.Errorf("missing detail: %d",w.Code)}
	publicBody:=request(h,"GET","/api/v1/public-information").Body.String()
	for _,privateField:=range []string{"student_id","private_notes","crawl_log","snapshot_ref"}{if strings.Contains(publicBody,privateField){t.Errorf("public DTO includes internal field %q",privateField)}}
	for _,path:=range []string{"/api/v1/public-information?date_from=2026-10-01","/api/v1/public-information?program=artificial-intelligence","/api/v1/public-information?unknown=x","/api/v1/public-information?college=a&college=b"}{if w:=request(h,"GET",path);w.Code!=400{t.Errorf("bad query %s: %d",path,w.Code)}}
	if w:=request(h,"POST","/api/v1/public-information");w.Code!=405{t.Errorf("write method: %d",w.Code)}
	if w:=request(h,"GET","/api/v1/private-data");w.Code!=401{t.Errorf("default deny: %d",w.Code)}
}

func TestPublicInstitutionDirectoryRoutes(t *testing.T){
	fixture:=adapter.NewSyntheticPublicInformation()
	h:=NewWithPublicData(service.Service{Version:"test-v1",MySQL:up,Qdrant:up,Now:time.Now},service.PublicInformation{Repository:fixture},service.PublicInstitutionDirectory{Repository:fixture},slog.New(slog.NewJSONHandler(&bytes.Buffer{},nil)))
	list:=request(h,"GET","/api/v1/public-institutions")
	if list.Code!=200||!strings.Contains(list.Body.String(),"示例大学"){t.Fatalf("directory list: %d %s",list.Code,list.Body.String())}
	detail:=request(h,"GET","/api/v1/public-institutions/"+fixture.Institutions[0].ID)
	if detail.Code!=200||!strings.Contains(detail.Body.String(),"example.invalid")||!strings.Contains(detail.Body.String(),"计算与智能学院"){t.Fatalf("directory detail: %d %s",detail.Code,detail.Body.String())}
	if w:=request(h,"GET","/api/v1/public-institutions/20000000-0000-4000-8000-000000000099");w.Code!=404{t.Errorf("missing institution: %d",w.Code)}
	if w:=request(h,"POST","/api/v1/public-institutions");w.Code!=405{t.Errorf("directory write method: %d",w.Code)}
}

type failingPublicRepository struct{}

func (failingPublicRepository) ListPublished(context.Context) ([]repository.PublicInformation, error) {
	return nil, errors.New("private storage failure")
}
func (failingPublicRepository) GetPublished(context.Context, string) (repository.PublicInformation, bool, error) {
	return repository.PublicInformation{}, false, errors.New("private storage failure")
}
func (failingPublicRepository) ListPublicInstitutions(context.Context) ([]repository.PublicInstitution, error) {
	return nil, errors.New("private storage failure")
}
func (failingPublicRepository) GetPublicInstitution(context.Context, string) (repository.PublicInstitution, bool, error) {
	return repository.PublicInstitution{}, false, errors.New("private storage failure")
}

func TestFailurePublicReadRepositoryContract(t *testing.T) {
	repository := failingPublicRepository{}
	h := NewWithPublicData(
		service.Service{Version: "test-v1", MySQL: up, Qdrant: up, Now: time.Now},
		service.PublicInformation{Repository: repository},
		service.PublicInstitutionDirectory{Repository: repository},
		slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil)),
	)
	for _, path := range []string{
		"/api/v1/public-information",
		"/api/v1/public-information/10000000-0000-4000-8000-000000000001",
		"/api/v1/public-information/calendar?month=2026-11",
		"/api/v1/public-institutions",
		"/api/v1/public-institutions/20000000-0000-4000-8000-000000000001",
	} {
		w := request(h, "GET", path)
		var env Envelope
		if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
			t.Fatalf("decode %s: %v", path, err)
		}
		if w.Code != http.StatusServiceUnavailable || env.Code != "DEPENDENCY_UNAVAILABLE" || strings.Contains(w.Body.String(), "private storage failure") {
			t.Errorf("repository failure contract %s: %d %s", path, w.Code, w.Body.String())
		}
	}
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
