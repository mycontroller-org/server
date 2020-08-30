package model

import (
	"time"
)

// State
const (
	StateUp          = "up"
	StateDown        = "down"
	StateUnavailable = "unavailable"
)

// State data
type State struct {
	Status  string    `json:"status"`
	Message string    `json:"message"`
	Since   time.Time `json:"since"`
}
