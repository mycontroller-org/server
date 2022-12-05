package cusomdatetime

import (
	"fmt"
	"strings"
	"time"
)

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
