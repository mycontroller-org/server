package metric

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
