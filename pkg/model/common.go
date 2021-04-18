package model

import (
	"time"
)

// State
const (
	StatusOk          = "ok"
	StatusUp          = "up"
	StatusDown        = "down"
	StatusError       = "error"
	StatusUnavailable = "unavailable"
)

// State data
type State struct {
	Status  string    `json:"status"`
	Message string    `json:"message"`
	Since   time.Time `json:"since"`
}

// File struct
type File struct {
	Name         string    `json:"name"`
	Size         int64     `json:"size"`
	CreationTime time.Time `json:"creationTime"`
	ModifiedTime time.Time `json:"modifiedTime"`
	Data         string    `json:"data"`
	IsDir        bool      `json:"isDir"`
	FullPath     string    `json:"fullPath"`
}
