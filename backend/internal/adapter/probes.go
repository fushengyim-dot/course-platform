package adapter

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"group-platform/internal/config"
	"net"
	"net/http"
	"strings"
	"time"
)

type SQLProbe struct{ DB *sql.DB }

func (p SQLProbe) Check(ctx context.Context) error { return p.DB.PingContext(ctx) }
func OpenMySQL(c config.Config) (*sql.DB, error) {
	cfg := mysql.NewConfig()
	cfg.User = c.MySQLUser
	cfg.Passwd = c.MySQLPassword
	cfg.Net = "tcp"
	cfg.Addr = net.JoinHostPort(c.MySQLHost, c.MySQLPort)
	cfg.DBName = c.MySQLDatabase
	cfg.Timeout = 2 * time.Second
	cfg.ReadTimeout = 2 * time.Second
	cfg.WriteTimeout = 2 * time.Second
	db, e := sql.Open("mysql", cfg.FormatDSN())
	if e != nil {
		return nil, fmt.Errorf("mysql initialization failed")
	}
	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(time.Minute)
	return db, nil
}

type HTTPProbe struct {
	URL    string
	Client *http.Client
}

func (p HTTPProbe) Check(ctx context.Context) error {
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(p.URL, "/")+"/readyz", nil)
	if e != nil {
		return fmt.Errorf("dependency request failed")
	}
	r, e := p.Client.Do(req)
	if e != nil {
		return fmt.Errorf("dependency unavailable")
	}
	defer r.Body.Close()
	if r.StatusCode != 200 {
		return fmt.Errorf("dependency not ready")
	}
	return nil
}
