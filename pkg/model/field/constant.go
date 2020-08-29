package field

// Field types
const (
	MetricTypeNone       = "none"
	MetricTypeCounter    = "counter"
	MetricTypeGauge      = "gauge"
	MetricTypeGaugeFloat = "gauge_float"
	MetricTypeBinary     = "binary"
	MetricTypeGEO        = "geo" // Geo Coordinates or GPS
)

// Unit types
const (
	UnitNone       = ""
	UnitCentigrade = "°C"
	UnitFahrenheit = "°F"
	UnitVolt       = "V"
	UnitMillivolt  = "mV"
	UnitMicrovolt  = "µV"
	UnitAmp        = "A"
	UnitMilliamp   = "mA"
	UnitMicroamp   = "µA"
	UnitHertz      = "Hz"
	UnitPercentage = "%"
	UnitOhm        = "Ω"
)

// Field names
const (
	FieldName           = "name"
	FieldVersion        = "version"
	FieldLibraryVersion = "library_version"
	FieldParentID       = "parent_id"
	FieldSignalStrength = "signal_strength"
	FieldBatteryLevel   = "battery_level"
	FieldType           = "type"
	FieldUnit           = "unit"
	FieldLocked         = "locked"
)
