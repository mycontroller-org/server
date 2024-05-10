package types

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
	Status  string    `json:"status" yaml:"status" structs:"status"`
	Message string    `json:"message" yaml:"message" structs:"message"`
	Since   time.Time `json:"since"  yaml:"since" structs:"since"`
}

// File struct
type File struct {
	Name         string    `json:"name" yaml:"name"`
	Size         int64     `json:"size" yaml:"size"`
	CreationTime time.Time `json:"creationTime" yaml:"creationTime"`
	ModifiedTime time.Time `json:"modifiedTime" yaml:"modifiedTime"`
	Data         string    `json:"data" yaml:"data"`
	IsDir        bool      `json:"isDir" yaml:"isDir"`
	FullPath     string    `json:"fullPath" yaml:"fullPath"`
}

// context key used on context
type ContextKey string
