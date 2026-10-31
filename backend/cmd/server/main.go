package main

import (
	"context"
	"fmt"
	"group-platform/internal/adapter"
	"group-platform/internal/config"
	"group-platform/internal/handler"
	"group-platform/internal/service"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
func run() error {
	c, e := config.Load(os.Getenv)
	if e != nil {
		return e
	}
	if len(os.Args) > 1 && os.Args[1] == "probe" {
		client := http.Client{Timeout: 3 * time.Second}
		_, port, _ := net.SplitHostPort(c.HTTPAddr)
		r, e := client.Get("http://" + net.JoinHostPort("127.0.0.1", port) + "/health")
		if e != nil {
			return fmt.Errorf("not ready")
		}
		defer r.Body.Close()
		if r.StatusCode != 200 {
			return fmt.Errorf("not ready")
		}
		return nil
	}
	db, e := adapter.OpenMySQL(c)
	if e != nil {
		return e
	}
	defer db.Close()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	client := &http.Client{Timeout: 2 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	svc := service.Service{Version: c.Version, MySQL: adapter.SQLProbe{DB: db}, Qdrant: adapter.HTTPProbe{URL: c.QdrantURL, Client: client}, Now: time.Now}
	srv := &http.Server{Addr: c.HTTPAddr, Handler: handler.New(svc, logger), ReadHeaderTimeout: 3 * time.Second, ReadTimeout: 5 * time.Second, WriteTimeout: 5 * time.Second, IdleTimeout: 30 * time.Second, MaxHeaderBytes: 8192}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	listener, err := net.Listen("tcp", c.HTTPAddr)
	if err != nil {
		return fmt.Errorf("server startup failed")
	}
	return serve(ctx, srv, listener)
}

func serve(ctx context.Context, srv *http.Server, listener net.Listener) error {
	done := make(chan error, 1)
	go func() { done <- srv.Serve(listener) }()
	select {
	case e := <-done:
		if e != http.ErrServerClosed {
			return fmt.Errorf("server startup failed")
		}
		return nil
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutdown)
	}
}
