package model

import (
	"time"
)

// State
const (
	StateUp          = "up"
	StateDown        = "down"
	StateError       = "error"
	StateUnavailable = "unavailable"
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
	ModifiedTime time.Time `json:"modifiedTime"`
	Data         string    `json:"data"`
}
