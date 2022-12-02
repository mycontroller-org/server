package schedule

import (
	"fmt"
	"strings"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
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
	ID                   string               `json:"id" yaml:"id"`
	Description          string               `json:"description" yaml:"description"`
	Enabled              bool                 `json:"enabled" yaml:"enabled"`
	Labels               cmap.CustomStringMap `json:"labels" yaml:"labels"`
	Validity             Validity             `json:"validity" yaml:"validity"`
	Type                 string               `json:"type" yaml:"type"`
	Spec                 cmap.CustomMap       `json:"spec" yaml:"spec"`
	Variables            map[string]string    `json:"variables" yaml:"variables"`
	CustomVariableType   string               `json:"customVariableType" yaml:"customVariableType"`
	CustomVariableConfig CustomVariableConfig `json:"customVariableConfig" yaml:"customVariableConfig"`
	HandlerParameters    map[string]string    `json:"handlerParameters" yaml:"handlerParameters"`
	Handlers             []string             `json:"handlers" yaml:"handlers"`
	ModifiedOn           time.Time            `json:"modifiedOn" yaml:"modifiedOn"`
	State                *State               `json:"state" yaml:"state"`
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
	From CustomDate `json:"from" yaml:"from"`
	To   CustomDate `json:"to" yaml:"to"`
}

// TimeRange struct
type TimeRange struct {
	From CustomTime `json:"from" yaml:"from"`
	To   CustomTime `json:"to" yaml:"to"`
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

// CustomDate used in validity
// Note: If we use CustomDate as time.Time alias. facing issue with gob encoding
// Issue: gob: type scheduler.CustomDate has no exported fields
// was used as "type CustomDate time.Time"
type CustomDate struct {
	time.Time
}

const CustomDateFormat = "2006-01-02"

// MarshalJSON custom implementation
func (cd CustomDate) MarshalJSON() ([]byte, error) {
	if cd.Time.IsZero() {
		return []byte("\"\""), nil
	}
	stamp := fmt.Sprintf("\"%s\"", cd.Time.Format(CustomDateFormat))
	return []byte(stamp), nil
}

// MarshalYAML implementation
func (cd CustomDate) MarshalYAML() (interface{}, error) {
	if cd.Time.IsZero() {
		return "", nil
	}
	return cd.Time.Format(CustomDateFormat), nil
}

// UnmarshalJSON custom implementation
func (cd *CustomDate) UnmarshalJSON(data []byte) error {
	stringDate := strings.Trim(string(data), `"`)
	return cd.Unmarshal(stringDate)
}

// UnmarshalYAML implementation
func (cd *CustomDate) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var stringDate string
	err := unmarshal(&stringDate)
	if err != nil {
		return nil
	}
	return cd.Unmarshal(stringDate)
}

func (cd *CustomDate) Unmarshal(stringDate string) error {
	var parsedDate time.Time

	if stringDate != "" {
		date, err := time.Parse(CustomDateFormat, stringDate)
		if err != nil {
			return err
		}
		parsedDate = date
	}
	*cd = CustomDate{Time: parsedDate}
	return nil
}

// CustomTime used in validity
type CustomTime struct {
	time.Time
}

const customTimeFormat = "15:04:05"

// MarshalJSON custom implementation
func (ct CustomTime) MarshalJSON() ([]byte, error) {
	if ct.Time.IsZero() {
		return []byte("\"\""), nil
	}
	stamp := fmt.Sprintf("\"%s\"", ct.Time.Format(customTimeFormat))
	return []byte(stamp), nil
}

// MarshalYAML implementation
func (ct CustomTime) MarshalYAML() (interface{}, error) {
	if ct.Time.IsZero() {
		return "", nil
	}
	return ct.Time.Format(customTimeFormat), nil
}

// UnmarshalJSON custom implementation
func (ct *CustomTime) UnmarshalJSON(data []byte) error {
	stringTime := strings.Trim(string(data), `"`)
	return ct.Unmarshal(stringTime)
}

// UnmarshalYAML implementation
func (ct *CustomTime) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var stringTime string
	err := unmarshal(&stringTime)
	if err != nil {
		return nil
	}
	return ct.Unmarshal(stringTime)
}

func (ct *CustomTime) Unmarshal(stringTime string) error {
	var parsedTime time.Time

	if stringTime != "" {
		if strings.Count(stringTime, ":") == 1 {
			stringTime = stringTime + ":00"
		}
		_time, err := time.Parse(customTimeFormat, stringTime)
		if err != nil {
			return err
		}
		parsedTime = _time
	}

	*ct = CustomTime{Time: parsedTime}
	return nil
}
