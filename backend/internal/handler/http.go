package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"group-platform/internal/service"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type Envelope struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Data      any    `json:"data"`
	RequestID string `json:"request_id"`
}

func New(s service.Service, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		b := make([]byte, 16)
		_, err := rand.Read(b)
		if err != nil {
			http.Error(w, "unavailable", 503)
			return
		}
		id := hex.EncodeToString(b)
		status := 200
		code := "OK"
		message := "服务正常"
		var data any
		route := "unregistered"
		switch {
		case strings.HasPrefix(r.URL.Path, "/api/"):
			status = 401
			code = "UNAUTHORIZED"
			message = "请先登录"
		case r.URL.Path != "/health" && r.URL.Path != "/live":
			status = 404
			code = "NOT_FOUND"
			message = "资源不存在"
		default:
			route = r.URL.Path
			if r.Method != http.MethodGet {
				status = 405
				code = "METHOD_NOT_ALLOWED"
				message = "请使用GET请求"
				w.Header().Set("Allow", "GET")
			} else if r.URL.RawQuery != "" {
				status = 400
				code = "INVALID_ARGUMENT"
				message = "该接口不接受查询参数"
			} else if route == "/live" {
				data = map[string]string{"status": "alive"}
			} else {
				h, ok := s.Check(r.Context())
				data = h
				if !ok {
					status = 503
					code = "DEPENDENCY_UNAVAILABLE"
					message = "依赖服务暂不可用"
				}
			}
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("X-Request-ID", id)
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(Envelope{code, message, data, id})
		logger.Info("request", "module", "http", "request_id", id, "method", r.Method, "route", route, "status", status, "duration_ms", time.Since(start).Milliseconds())
	})
}
