package schedule

import (
	"time"

	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	dateTimeTY "github.com/mycontroller-org/server/v2/pkg/types/cusom_datetime"
)

// cron types
const (
	TypeRepeat  = "repeat"
	TypeCron    = "cron"
	TypeSimple  = "simple"
	TypeSunrise = "sunrise"
	TypeSunset  = "sunset"
)

// frequency types
const (
	FrequencyDaily   = "daily"
	FrequencyWeekly  = "weekly"
	FrequencyMonthly = "monthly"
	FrequencyOnDate  = "on_date"
)

// Custom variable loader types
const (
	CustomVariableTypeNone       = "none"
	CustomVariableTypeJavascript = "javascript"
	CustomVariableTypeWebhook    = "webhook"
)

// Config for scheduler
type Config struct {
	ID                   string                 `json:"id" yaml:"id"`
	Description          string                 `json:"description" yaml:"description"`
	Enabled              bool                   `json:"enabled" yaml:"enabled"`
	Labels               cmap.CustomStringMap   `json:"labels" yaml:"labels"`
	Validity             Validity               `json:"validity" yaml:"validity"`
	Type                 string                 `json:"type" yaml:"type"`
	Spec                 cmap.CustomMap         `json:"spec" yaml:"spec"`
	Variables            map[string]interface{} `json:"variables" yaml:"variables"`
	CustomVariableType   string                 `json:"customVariableType" yaml:"customVariableType"`
	CustomVariableConfig CustomVariableConfig   `json:"customVariableConfig" yaml:"customVariableConfig"`
	HandlerParameters    map[string]interface{} `json:"handlerParameters" yaml:"handlerParameters"`
	Handlers             []string               `json:"handlers" yaml:"handlers"`
	ModifiedOn           time.Time              `json:"modifiedOn" yaml:"modifiedOn"`
	State                *State                 `json:"state" yaml:"state"`
}

// CustomVariableConfig struct
type CustomVariableConfig struct {
	Javascript string      `json:"javascript" yaml:"javascript"`
	Webhook    WebhookData `json:"webhook" yaml:"webhook"`
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

// Validity of the scheduler
type Validity struct {
	Enabled              bool      `json:"enabled" yaml:"enabled"`
	Date                 DateRange `json:"date" yaml:"date"`
	Time                 TimeRange `json:"time" yaml:"time"`
	ValidateTimeEveryday bool      `json:"validateTimeEveryday" yaml:"validateTimeEveryday"`
}

// DateRange struct
type DateRange struct {
	From dateTimeTY.CustomDate `json:"from" yaml:"from"`
	To   dateTimeTY.CustomDate `json:"to" yaml:"to"`
}

// TimeRange struct
type TimeRange struct {
	From dateTimeTY.CustomTime `json:"from" yaml:"from"`
	To   dateTimeTY.CustomTime `json:"to" yaml:"to"`
}

// State struct
type State struct {
	LastRun       time.Time `json:"lastRun" yaml:"lastRun"`
	LastStatus    bool      `json:"lastStatus" yaml:"lastStatus"`
	Message       string    `json:"message" yaml:"message"`
	ExecutedCount int64     `json:"executedCount" yaml:"executedCount"`
}

// spec for each type

// SpecRepeat struct
type SpecRepeat struct {
	Interval    string
	RepeatCount int64
}

// SpecCron struct
type SpecCron struct {
	CronExpression string
}

// SpecSimple struct
type SpecSimple struct {
	Frequency   string
	DayOfWeek   string
	DateOfMonth int
	Date        string
	Time        string
	Offset      string
}
