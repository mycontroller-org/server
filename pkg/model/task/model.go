package task

import (
	"time"

	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
)

// dampening types
const (
	dampeningTypeNone        = "none"
	dampeningTypeConsecutive = "consecutive"
	dampeningTypeEvaluations = "evaluations"
	dampeningTypeActiveTime  = "active_time"
)

// Config struct
type Config struct {
	ID                string               `json:"id"`
	Description       string               `json:"description"`
	Labels            cmap.CustomStringMap `json:"labels"`
	Enabled           bool                 `json:"enabled"`
	IgnoreDuplicate   bool                 `json:"ignoreDuplicate"`
	AutoDisable       bool                 `json:"autoDisable"`
	Variables         map[string]string    `json:"variables"`
	Dampening         Dampening            `json:"dampening"`
	TriggerOnEvent    bool                 `json:"triggerOnEvent"`
	EventFilter       EventFilter          `json:"eventFilter"`
	ExecutionInterval string               `json:"executionInterval"`
	RemoteCall        bool                 `json:"remoteCall"`
	RemoteCallConfig  interface{}          `json:"remoteCallConfig"`
	Rule              Rule                 `json:"rule"`
	Notify            []string             `json:"notify"`
	State             *State               `json:"state"`
}

// EventFilter struct
type EventFilter struct {
	Selectors     cmap.CustomStringMap `json:"selectors"`
	ResourceTypes []string             `json:"resourceTypes"`
}

// Resource struct
type Resource struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	SubType string `json:"subType"`
}

// Rule struct
type Rule struct {
	MatchAll   bool         `json:"matchAll"`
	Conditions []Conditions `json:"conditions"`
}

// Conditions struct
type Conditions struct {
	Variable string      `json:"variable"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

// Dampening struct
type Dampening struct {
	Type        string `json:"type"`
	Occurrences int64  `json:"occurrences"`
	Evaluations int64  `json:"evaluations"`
	ActiveTime  string `json:"activeTime"`
}

// State struct
type State struct {
	LastEvaluation time.Time `json:"lastEvaluation"`
	LastSuccess    time.Time `json:"lastSuccess"`
	Message        string    `json:"message"`
	LastStatus     bool      `json:"lastStatus"`
	ExecutedCount  int64     `json:"executedCount"`
	Executions     []bool    `json:"executions"`
	ActiveFrom     string    `json:"activeFrom"`
}
