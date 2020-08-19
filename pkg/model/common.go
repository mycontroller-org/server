package model

// config map keys
const (
	CFGUpdateName = "updateName"
)

// State
const (
	StateUp          = "up"
	StateDown        = "down"
	StateUnavailable = "unavailable"
)

// State data
type State struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Since   uint64 `json:"since"`
}
