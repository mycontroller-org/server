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
	ID                string               `json:"id" yaml:"id"`
	Description       string               `json:"description" yaml:"description"`
	Labels            cmap.CustomStringMap `json:"labels" yaml:"labels"`
	Enabled           bool                 `json:"enabled" yaml:"enabled"`
	IgnoreDuplicate   bool                 `json:"ignoreDuplicate" yaml:"ignoreDuplicate"`
	AutoDisable       bool                 `json:"autoDisable" yaml:"autoDisable"`
	ReEnable          bool                 `json:"reEnable" yaml:"reEnable"`
	ReEnableDuration  string               `json:"reEnableDuration" yaml:"reEnableDuration"`
	Variables         map[string]string    `json:"variables" yaml:"variables"`
	Dampening         DampeningConfig      `json:"dampening" yaml:"dampening"`
	TriggerOnEvent    bool                 `json:"triggerOnEvent" yaml:"triggerOnEvent"`
	EventFilter       EventFilter          `json:"eventFilter" yaml:"eventFilter"`
	ExecutionInterval string               `json:"executionInterval" yaml:"executionInterval"`
	EvaluationType    string               `json:"evaluationType" yaml:"evaluationType"`
	EvaluationConfig  EvaluationConfig     `json:"evaluationConfig" yaml:"evaluationConfig"`
	HandlerParameters map[string]string    `json:"handlerParameters" yaml:"handlerParameters"`
	Handlers          []string             `json:"handlers" yaml:"handlers"`
	ModifiedOn        time.Time            `json:"modifiedOn" yaml:"modifiedOn"`
	State             *State               `json:"state" yaml:"state"`
}

// EventFilter struct
type EventFilter struct {
	EventTypes  []string             `json:"eventTypes" yaml:"eventTypes"`
	EntityTypes []string             `json:"entityTypes" yaml:"entityTypes"`
	Filters     cmap.CustomStringMap `json:"filters" yaml:"filters"`
}

// EvaluationConfig struct
type EvaluationConfig struct {
	Rule       Rule        `json:"rule" yaml:"rule"`
	Javascript string      `json:"javascript" yaml:"javascript"`
	Webhook    WebhookData `json:"webhook" yaml:"webhook"`
}

// Rule struct
type Rule struct {
	MatchAll   bool         `json:"matchAll" yaml:"matchAll"`
	Conditions []Conditions `json:"conditions" yaml:"conditions"`
}

// WebhookData struct
type WebhookData struct {
	URL             string                 `json:"url" yaml:"url"`
	Method          string                 `json:"method" yaml:"method"`
	Insecure        bool                   `json:"insecure" yaml:"insecure"`
	Headers         map[string]string      `json:"headers" yaml:"headers"`
	QueryParameters map[string]interface{} `json:"queryParameters" yaml:"queryParameters"`
	IncludeConfig   bool                   `json:"includeConfig" yaml:"includeConfig"`
}

// Conditions struct
type Conditions struct {
	Variable string      `json:"variable" yaml:"variable"`
	Operator string      `json:"operator" yaml:"operator"`
	Value    interface{} `json:"value" yaml:"value"`
}

// DampeningConfig struct
type DampeningConfig struct {
	Type           string `json:"type" yaml:"type"`
	Occurrences    int64  `json:"occurrences" yaml:"occurrences"`
	Evaluation     int64  `json:"evaluation" yaml:"evaluation"`
	ActiveDuration string `json:"activeDuration" yaml:"activeDuration"`
}

// State struct
type State struct {
	LastEvaluation    time.Time        `json:"lastEvaluation" yaml:"lastEvaluation"`
	LastSuccess       time.Time        `json:"lastSuccess" yaml:"lastSuccess"`
	Message           string           `json:"message" yaml:"message"`
	LastDuration      string           `json:"lastDuration" yaml:"lastDuration"`
	LastStatus        bool             `json:"lastStatus" yaml:"lastStatus"`
	ExecutedCount     int64            `json:"executedCount" yaml:"executedCount"`
	ExecutionsHistory []ExecutionState `json:"executionsHistory" yaml:"executionsHistory"`
	ActiveSince       time.Time        `json:"activeSince" yaml:"activeSince"`
}

type ExecutionState struct {
	Triggered bool      `json:"triggered" yaml:"triggered"`
	Timestamp time.Time `json:"timestamp" yaml:"timestamp"`
}
