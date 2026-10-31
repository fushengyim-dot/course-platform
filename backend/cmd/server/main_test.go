package main

import (
	"context"
	"io"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestNormalShutdownDrainsInflight(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	entered := make(chan struct{})
	release := make(chan struct{})
	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(entered)
		<-release
		_, _ = w.Write([]byte("completed"))
	})}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stopped := make(chan error, 1)
	go func() { stopped <- serve(ctx, srv, listener) }()
	response := make(chan string, 1)
	go func() {
		c := http.Client{Timeout: 3 * time.Second}
		r, e := c.Get("http://" + listener.Addr().String())
		if e != nil {
			response <- "request failed"
			return
		}
		defer r.Body.Close()
		b, _ := io.ReadAll(r.Body)
		response <- string(b)
	}()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("request did not start")
	}
	cancel()
	select {
	case <-stopped:
		t.Fatal("server stopped before in-flight request finished")
	case <-time.After(30 * time.Millisecond):
	}
	close(release)
	if body := <-response; body != "completed" {
		t.Fatal(body)
	}
	select {
	case e := <-stopped:
		if e != nil {
			t.Fatal(e)
		}
	case <-time.After(time.Second):
		t.Fatal("shutdown did not finish")
	}
}
