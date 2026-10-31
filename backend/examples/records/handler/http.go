package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"group-platform/examples/records/repository"
	"group-platform/examples/records/service"
	production "group-platform/internal/handler"
	"io"
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"
)

type Identity func(*http.Request) (string, bool)

var keyPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,128}$`)
var idPattern = regexp.MustCompile(`^[1-9][0-9]*$`)

func New(s service.Service, identity Identity) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := make([]byte, 16)
		if _, err := rand.Read(b); err != nil {
			http.Error(w, "unavailable", 503)
			return
		}
		id := hex.EncodeToString(b)
		reply := func(status int, code string, data any) {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Header().Set("Cache-Control", "no-store")
			w.Header().Set("X-Request-ID", id)
			w.WriteHeader(status)
			_ = json.NewEncoder(w).Encode(production.Envelope{Code: code, Message: code, Data: data, RequestID: id})
		}
		subject, ok := "", false
		if identity != nil {
			subject, ok = identity(r)
		}
		if !ok || strings.TrimSpace(subject) == "" {
			reply(401, "UNAUTHORIZED", nil)
			return
		}
		path := r.URL.Path
		collection := path == "/api/v1/records"
		recordID := strings.TrimPrefix(path, "/api/v1/records/")
		item := strings.HasPrefix(path, "/api/v1/records/") && idPattern.MatchString(recordID)
		if !collection && !item {
			reply(404, "NOT_FOUND", nil)
			return
		}
		method := "GET"
		if collection {
			method = "POST"
		}
		if r.Method != method {
			w.Header().Set("Allow", method)
			reply(405, "METHOD_NOT_ALLOWED", nil)
			return
		}
		if r.URL.RawQuery != "" || r.URL.ForceQuery {
			reply(400, "INVALID_ARGUMENT", nil)
			return
		}
		var result repository.Record
		var err error
		if collection {
			key := r.Header.Get("Idempotency-Key")
			if !keyPattern.MatchString(key) || len(r.Header.Values("Idempotency-Key")) != 1 {
				reply(400, "INVALID_ARGUMENT", nil)
				return
			}
			raw, e := io.ReadAll(http.MaxBytesReader(w, r.Body, 1024))
			if e != nil || !utf8.Valid(raw) {
				reply(400, "INVALID_ARGUMENT", nil)
				return
			}
			var body struct {
				Title string `json:"title"`
			}
			decoder := json.NewDecoder(strings.NewReader(string(raw)))
			decoder.DisallowUnknownFields()
			if decoder.Decode(&body) != nil {
				reply(400, "INVALID_ARGUMENT", nil)
				return
			}
			var trailing any
			if decoder.Decode(&trailing) != io.EOF {
				reply(400, "INVALID_ARGUMENT", nil)
				return
			}
			body.Title = strings.TrimSpace(body.Title)
			n := utf8.RuneCountInString(body.Title)
			if n < 1 || n > 80 {
				reply(400, "INVALID_ARGUMENT", nil)
				return
			}
			result, err = s.Create(subject, key, body.Title)
		} else {
			result, err = s.Get(subject, recordID)
		}
		switch {
		case errors.Is(err, service.ErrUnauthorized):
			reply(401, "UNAUTHORIZED", nil)
		case errors.Is(err, repository.ErrNotFound):
			reply(404, "NOT_FOUND", nil)
		case errors.Is(err, repository.ErrConflict):
			reply(409, "CONFLICT", nil)
		case err != nil:
			reply(503, "DEPENDENCY_UNAVAILABLE", nil)
		case collection:
			reply(201, "CREATED", result)
		default:
			reply(200, "OK", result)
		}
	})
}
