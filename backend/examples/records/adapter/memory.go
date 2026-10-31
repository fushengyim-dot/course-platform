package adapter

import (
	"group-platform/examples/records/repository"
	"strconv"
	"sync"
)

type Memory struct {
	mu      sync.Mutex
	records map[string]repository.Record
	keys    map[[2]string]string
}

func New() *Memory {
	return &Memory{records: map[string]repository.Record{}, keys: map[[2]string]string{}}
}
func (m *Memory) Create(subject, key, title string) (repository.Record, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	k := [2]string{subject, key}
	if id, ok := m.keys[k]; ok {
		r := m.records[id]
		if r.Title != title {
			return repository.Record{}, repository.ErrConflict
		}
		return r, nil
	}
	r := repository.Record{ID: strconv.Itoa(len(m.records) + 1), Title: title, Subject: subject}
	m.records[r.ID] = r
	m.keys[k] = r.ID
	return r, nil
}
func (m *Memory) Get(id string) (repository.Record, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.records[id]
	if !ok {
		return r, repository.ErrNotFound
	}
	return r, nil
}
func (m *Memory) Count() int { m.mu.Lock(); defer m.mu.Unlock(); return len(m.records) }
