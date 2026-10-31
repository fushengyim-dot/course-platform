package adapter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNormalHTTPProbe(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/readyz" {
			t.Error("wrong readiness path")
		}
		w.WriteHeader(200)
	}))
	defer server.Close()
	p := HTTPProbe{URL: server.URL, Client: server.Client()}
	if e := p.Check(context.Background()); e != nil {
		t.Fatal(e)
	}
}
func TestFailureHTTPStatus(t *testing.T) {
	for _, status := range []int{401, 503, 302} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(status) }))
		p := HTTPProbe{URL: server.URL, Client: server.Client()}
		if p.Check(context.Background()) == nil {
			t.Error("accepted non-ready status", status)
		}
		server.Close()
	}
}
func TestBoundaryHTTPDeadline(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { <-r.Context().Done() }))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	p := HTTPProbe{URL: server.URL, Client: server.Client()}
	if p.Check(ctx) == nil {
		t.Fatal("timeout ignored")
	}
}
