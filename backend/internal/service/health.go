package service

import (
	"context"
	"group-platform/internal/repository"
	"time"
)

type Health struct {
	Version      string            `json:"version"`
	Time         string            `json:"time"`
	Dependencies map[string]string `json:"dependencies"`
}
type Service struct {
	Version       string
	MySQL, Qdrant repository.Probe
	Now           func() time.Time
}

func (s Service) Check(ctx context.Context) (Health, bool) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	h := Health{Version: s.Version, Time: s.Now().UTC().Format(time.RFC3339Nano), Dependencies: map[string]string{}}
	ok := true
	for _, p := range []struct {
		name  string
		probe repository.Probe
	}{{"mysql", s.MySQL}, {"qdrant", s.Qdrant}} {
		state := "up"
		if p.probe.Check(ctx) != nil {
			state = "down"
			ok = false
		}
		h.Dependencies[p.name] = state
	}
	return h, ok
}
