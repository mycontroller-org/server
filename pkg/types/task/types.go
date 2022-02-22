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
	DampeningTypeNone        = "none"
	DampeningTypeConsecutive = "consecutive"
	DampeningTypeEvaluations = "evaluations"
	DampeningTypeActiveTime  = "active_time"
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
	Variables         map[string]string    `json:"variables"`
	Dampening         Dampening            `json:"dampening"`
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
