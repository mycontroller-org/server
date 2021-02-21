package scheduler

import (
	"fmt"
	"strings"
	"time"

	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
)

// cron types
const (
	TypeRepeat   = "repeat"
	TypeCron     = "cron"
	TypeSimple   = "simple"
	TypeSunrise  = "sunrise"
	TypeSunset   = "sunset"
	TypeMoonrise = "moonrise"
	TypeMoonset  = "moonset"
)

// frequency types
const (
	FrequencyDaily   = "daily"
	FrequencyWeekly  = "weekly"
	FrequencyMonthly = "monthly"
)

// Config for scheduler
type Config struct {
	ID          string               `json:"id"`
	Description string               `json:"description"`
	Enabled     bool                 `json:"enabled"`
	Labels      cmap.CustomStringMap `json:"labels"`
	Variables   map[string]string    `json:"variables"`
	Validity    Validity             `json:"validity"`
	Type        string               `json:"type"`
	Spec        cmap.CustomMap       `json:"spec"`
	Notify      []string             `json:"notify"`
	State       *State               `json:"state"`
}

// Validity of the scheduler
type Validity struct {
	Date                 DateRange `json:"date"`
	Time                 TimeRange `json:"time"`
	ValidateTimeEveryday bool      `json:"validateTimeEveryday"`
}

// DateRange struct
type DateRange struct {
	From CustomDate `json:"from"`
	To   CustomDate `json:"to"`
}

// TimeRange struct
type TimeRange struct {
	From CustomTime `json:"from"`
	To   CustomTime `json:"to"`
}

// State struct
type State struct {
	LastRun       time.Time `json:"lastRun"`
	LastStatus    bool      `json:"lastStatus"`
	Message       string    `json:"message"`
	ExecutedCount int64     `json:"executedCount"`
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

const customDateFormat = "2006-01-02"

// MarshalJSON custom implementation
func (cd CustomDate) MarshalJSON() ([]byte, error) {
	_time := time.Time(cd.Time)
	if _time.IsZero() {
		return []byte("\"\""), nil
	}
	stamp := fmt.Sprintf("\"%s\"", time.Time(cd.Time).Format(customDateFormat))
	return []byte(stamp), nil
}

// MarshalYAML implementation
func (cd CustomDate) MarshalYAML() (interface{}, error) {
	return cd.Time.Format(customDateFormat), nil
}

// UnmarshalJSON custom implementation
func (cd *CustomDate) UnmarshalJSON(data []byte) error {
	stringDate := strings.Trim(string(data), `"`)
	return cd.unmarshal(stringDate)
}

// UnmarshalYAML implementation
func (cd *CustomDate) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var stringDate string
	err := unmarshal(&stringDate)
	if err != nil {
		return nil
	}
	return cd.unmarshal(stringDate)
}

func (cd *CustomDate) unmarshal(stringDate string) error {
	var parsedDate time.Time

	if stringDate != "" {
		date, err := time.Parse(customDateFormat, stringDate)
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
	_time := time.Time(ct.Time)
	if _time.IsZero() {
		return []byte("\"\""), nil
	}
	stamp := fmt.Sprintf("\"%s\"", time.Time(ct.Time).Format(customTimeFormat))
	return []byte(stamp), nil
}

// MarshalYAML implementation
func (ct CustomTime) MarshalYAML() (interface{}, error) {
	return ct.Time.Format(customTimeFormat), nil
}

// UnmarshalJSON custom implementation
func (ct *CustomTime) UnmarshalJSON(data []byte) error {
	stringTime := strings.Trim(string(data), `"`)
	return ct.unmarshal(stringTime)
}

// UnmarshalYAML implementation
func (ct *CustomTime) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var stringTime string
	err := unmarshal(&stringTime)
	if err != nil {
		return nil
	}
	return ct.unmarshal(stringTime)
}

func (ct *CustomTime) unmarshal(stringTime string) error {
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
