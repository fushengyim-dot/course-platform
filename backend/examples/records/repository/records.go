package repository

import "errors"

type Record struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Subject string `json:"-"`
}

var ErrNotFound = errors.New("not found")
var ErrConflict = errors.New("conflict")

type Records interface {
	Create(subject, key, title string) (Record, error)
	Get(id string) (Record, error)
}
