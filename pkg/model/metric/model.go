package metric

import "time"

// Metric types
const (
	MetricTypeNone       = "none"
	MetricTypeCounter    = "counter"
	MetricTypeGauge      = "gauge"
	MetricTypeGaugeFloat = "gauge_float"
	MetricTypeBinary     = "binary"
	MetricTypeGEO        = "geo" // Geo Coordinates or GPS
)

// MyController follows unit details from grafana, take unit details from here
// Source: https://github.com/grafana/grafana/blob/v6.7.1/packages/grafana-data/src/valueFormats/categories.ts#L23
const (
	UnitNone       = "none"
	UnitCelsius    = "celsius"
	UnitFahrenheit = "fahrenheit"
	UnitKelvin     = "kelvin"
	UnitHumidity   = "humidity"
	UnitPercent    = "percent"
	UnitVoltage    = "volt"
	UnitAmpere     = "amp"
)

// QueryConfig parameters
type QueryConfig struct {
	Global     Query   `json:"global"`
	Individual []Query `json:"individual"`
}

// Query paramaters
type Query struct {
	Name       string            `json:"name"`
	MetricType string            `json:"metricType"`
	Start      string            `json:"start"`
	Stop       string            `json:"stop"`
	Window     string            `json:"window"`
	Tags       map[string]string `json:"tags"`
	Functions  []string          `json:"functions"`
}

// Clone a query
func (q *Query) Clone() Query {
	tags := make(map[string]string)
	if q.Tags != nil {
		for k, v := range q.Tags {
			tags[k] = v
		}
	}
	functions := []string{}
	if q.Functions != nil {
		functions = q.Functions
	}
	return Query{
		Name:       q.Name,
		MetricType: q.MetricType,
		Start:      q.Start,
		Stop:       q.Stop,
		Window:     q.Window,
		Tags:       tags,
		Functions:  functions,
	}
}

// Merge data from another query
func (q *Query) Merge(new *Query) {
	if new != nil {
		// update default values
		if q.Tags == nil {
			q.Tags = make(map[string]string)
		}
		if q.Functions == nil {
			q.Functions = []string{}
		}
		// update vales
		if new.Name != "" {
			q.Name = new.Name
		}
		if new.MetricType != "" {
			q.MetricType = new.MetricType
		}
		if new.Start != "" {
			q.Start = new.Start
		}
		if new.Stop != "" {
			q.Stop = new.Stop
		}
		if new.Window != "" {
			q.Window = new.Window
		}
		if len(new.Tags) > 0 {
			for k, v := range new.Tags {
				q.Tags[k] = v
			}
		}
		if len(new.Functions) > 0 {
			for _, newFn := range new.Functions {
				found := false
				for _, orgFn := range q.Functions {
					if orgFn == newFn {
						found = true
						break
					}
				}
				if !found {
					q.Functions = append(q.Functions, newFn)
				}
			}
		}
	}
}

// Data struct
type Data struct {
	Time       time.Time              `json:"timestamp"`
	MetricType string                 `json:"metricType"`
	Metric     map[string]interface{} `json:"metric"`
}
