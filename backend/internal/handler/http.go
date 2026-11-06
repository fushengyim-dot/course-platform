package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"group-platform/internal/service"
	"log/slog"
	"net/http"
	"strings"
	"net/url"
	"time"
)

type Envelope struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Data      any    `json:"data"`
	RequestID string `json:"request_id"`
}

func New(s service.Service, logger *slog.Logger) http.Handler {
	return NewWithPublicInformation(s, service.PublicInformation{}, logger)
}

func NewWithPublicInformation(s service.Service, information service.PublicInformation, logger *slog.Logger) http.Handler {
	return NewWithPublicData(s, information, service.PublicInstitutionDirectory{}, logger)
}

func NewWithPublicData(s service.Service, information service.PublicInformation, institutions service.PublicInstitutionDirectory, logger *slog.Logger) http.Handler {
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
		case r.URL.Path == "/api/v1/public-institutions" || strings.HasPrefix(r.URL.Path, "/api/v1/public-institutions/"):
			route = "public-institutions"
			if r.Method != http.MethodGet {
				status, code, message = 405, "METHOD_NOT_ALLOWED", "请使用GET请求"
				w.Header().Set("Allow", "GET")
			} else if r.URL.RawQuery != "" {
				status, code, message = 400, "INVALID_ARGUMENT", "该接口不接受查询参数"
			} else if r.URL.Path == "/api/v1/public-institutions" {
				if items, err := institutions.List(r.Context()); err != nil {
					status, code, message = 503, "UNAVAILABLE", "院校信息暂时不可用"
				} else {
					data = items
				}
			} else {
				id := strings.TrimPrefix(r.URL.Path, "/api/v1/public-institutions/")
				if strings.Contains(id, "/") {
					status, code, message = 404, "NOT_FOUND", "资源不存在"
				} else if item, err := institutions.Get(r.Context(), id); err != nil {
					if err == service.ErrPublicInstitutionNotFound {
						status, code, message = 404, "NOT_FOUND", "资源不存在"
					} else {
						status, code, message = 503, "UNAVAILABLE", "院校信息暂时不可用"
					}
				} else {
					data = item
				}
			}
		case r.URL.Path == "/api/v1/public-information" || strings.HasPrefix(r.URL.Path, "/api/v1/public-information/"):
			route = "public-information"
			if r.Method != http.MethodGet {
				status, code, message = 405, "METHOD_NOT_ALLOWED", "请使用GET请求"
				w.Header().Set("Allow", "GET")
			} else if r.URL.Path == "/api/v1/public-information/calendar" {
				values := r.URL.Query()
				if len(values) != 1 || len(values["month"]) != 1 {
					status, code, message = 400, "INVALID_ARGUMENT", "月份格式无效"
				} else if calendar, err := information.Calendar(r.Context(), values.Get("month")); err != nil {
					if err == service.ErrInvalidPublicQuery {
						status, code, message = 400, "INVALID_ARGUMENT", "月份格式无效"
					} else {
						status, code, message = 503, "DEPENDENCY_UNAVAILABLE", "信息暂时不可用"
					}
				} else {
					data = calendar
				}
			} else if r.URL.Path == "/api/v1/public-information" {
				query, ok := publicInformationQuery(r.URL.Query())
				if !ok { status, code, message = 400, "INVALID_ARGUMENT", "检索条件无效" } else if page, err := information.List(r.Context(), query); err != nil {
					if err == service.ErrInvalidPublicQuery { status, code, message = 400, "INVALID_ARGUMENT", "检索条件无效" } else { status, code, message = 503, "UNAVAILABLE", "信息暂时不可用" }
				} else { data = page }
			} else if r.URL.RawQuery != "" {
				status, code, message = 400, "INVALID_ARGUMENT", "详情接口不接受查询参数"
			} else {
				id := strings.TrimPrefix(r.URL.Path, "/api/v1/public-information/")
				if strings.Contains(id, "/") { status, code, message = 404, "NOT_FOUND", "资源不存在" } else if item, err := information.Get(r.Context(), id); err != nil {
					if err == service.ErrPublicInformationNotFound { status, code, message = 404, "NOT_FOUND", "资源不存在" } else { status, code, message = 503, "UNAVAILABLE", "信息暂时不可用" }
				} else { data = item }
			}
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

func publicInformationQuery(values url.Values) (service.PublicInformationQuery, bool) {
	allowed := map[string]bool{"institution":true,"college":true,"expected_publish_from":true,"expected_publish_to":true,"sort":true,"page_size":true,"cursor":true}
	for key, entries := range values { if !allowed[key] || len(entries) != 1 { return service.PublicInformationQuery{}, false } }
	for _, key := range []string{"page_size"} { if len(values[key]) != 0 && len(values[key][0]) > 3 { return service.PublicInformationQuery{}, false } }
	pageSize := 0
	if values.Get("page_size") != "" { for _, c := range values.Get("page_size") { if c < '0' || c > '9' { return service.PublicInformationQuery{}, false }; pageSize = pageSize*10 + int(c-'0') } }
	return service.PublicInformationQuery{Institution:values.Get("institution"),College:values.Get("college"),ExpectedPublishFrom:values.Get("expected_publish_from"),ExpectedPublishTo:values.Get("expected_publish_to"),Sort:values.Get("sort"),PageSize:pageSize,Cursor:values.Get("cursor")}, true
}
