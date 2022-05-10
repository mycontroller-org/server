package storage

import "errors"

var (
	ErrNoDocuments = errors.New("no documents in result")
	ErrNilFilter   = errors.New("filter can not be nil")
	ErrNilData     = errors.New("data can not be nil")
)
