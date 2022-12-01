package task

import (
	"time"

	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
)

// Evaluation types
const (
	EvaluationTypeRule       = "rule"
	EvaluationTypeJavascript = "javascript"
	EvaluationTypeWebhook    = "webhook"
)

// dampening types
const (
	DampeningTypeNone           = "none"
	DampeningTypeConsecutive    = "consecutive"
	DampeningTypeEvaluation     = "evaluation"
	DampeningTypeActiveDuration = "active_duration"
)

// keys used in script engine
const (
	KeyIsTriggered = "isTriggered" // expected value from script or from webhook to trigger
)

// Config struct
type Config struct {
	ID                string               `json:"id"`
	Description       string               `json:"description"`
	Labels            cmap.CustomStringMap `json:"labels"`
	Enabled           bool                 `json:"enabled"`
	IgnoreDuplicate   bool                 `json:"ignoreDuplicate"`
	AutoDisable       bool                 `json:"autoDisable"`
	ReEnable          bool                 `json:"reEnable"`
	ReEnableDuration  string               `json:"reEnableDuration"`
	Variables         map[string]string    `json:"variables"`
	Dampening         DampeningConfig      `json:"dampening"`
	TriggerOnEvent    bool                 `json:"triggerOnEvent"`
	EventFilter       EventFilter          `json:"eventFilter"`
	ExecutionInterval string               `json:"executionInterval"`
	EvaluationType    string               `json:"evaluationType"`
	EvaluationConfig  EvaluationConfig     `json:"evaluationConfig"`
	HandlerParameters map[string]string    `json:"handlerParameters"`
	Handlers          []string             `json:"handlers"`
	ModifiedOn        time.Time            `json:"modifiedOn"`
	State             *State               `json:"state"`
}

// EventFilter struct
type EventFilter struct {
	EventTypes  []string             `json:"eventTypes"`
	EntityTypes []string             `json:"entityTypes"`
	Filters     cmap.CustomStringMap `json:"filters"`
}

// EvaluationConfig struct
type EvaluationConfig struct {
	Rule       Rule        `json:"rule"`
	Javascript string      `json:"javascript"`
	Webhook    WebhookData `json:"webhook"`
}

// Rule struct
type Rule struct {
	MatchAll   bool         `json:"matchAll"`
	Conditions []Conditions `json:"conditions"`
}

// WebhookData struct
type WebhookData struct {
	URL             string                 `json:"url"`
	Method          string                 `json:"method"`
	Insecure        bool                   `json:"insecure"`
	Headers         map[string]string      `yaml:"headers"`
	QueryParameters map[string]interface{} `yaml:"queryParameters"`
	IncludeConfig   bool                   `json:"includeConfig"`
}

// Conditions struct
type Conditions struct {
	Variable string      `json:"variable"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

// DampeningConfig struct
type DampeningConfig struct {
	Type           string `json:"type"`
	Occurrences    int64  `json:"occurrences"`
	Evaluation     int64  `json:"evaluation"`
	ActiveDuration string `json:"activeDuration"`
}

// State struct
type State struct {
	LastEvaluation    time.Time        `json:"lastEvaluation"`
	LastSuccess       time.Time        `json:"lastSuccess"`
	Message           string           `json:"message"`
	LastDuration      string           `json:"lastDuration"`
	LastStatus        bool             `json:"lastStatus"`
	ExecutedCount     int64            `json:"executedCount"`
	ExecutionsHistory []ExecutionState `json:"executionsHistory"`
	ActiveSince       time.Time        `json:"activeSince"`
}

type ExecutionState struct {
	Triggered bool      `json:"triggered"`
	Timestamp time.Time `json:"timestamp"`
}
