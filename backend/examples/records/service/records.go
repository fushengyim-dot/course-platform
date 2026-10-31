package service

import (
	"errors"
	"group-platform/examples/records/repository"
)

var ErrUnauthorized = errors.New("unauthorized")

type Service struct{ Repository repository.Records }

func (s Service) Create(subject, key, title string) (repository.Record, error) {
	if subject == "" {
		return repository.Record{}, ErrUnauthorized
	}
	return s.Repository.Create(subject, key, title)
}
func (s Service) Get(subject, id string) (repository.Record, error) {
	if subject == "" {
		return repository.Record{}, ErrUnauthorized
	}
	r, err := s.Repository.Get(id)
	if err != nil {
		return repository.Record{}, err
	}
	if r.Subject != subject {
		return repository.Record{}, repository.ErrNotFound
	}
	return r, nil
}
