package telemetry

// telemetry payload
type Payload struct {
	APIVersion  string      `json:"apiVersion"`
	TelemetryID string      `json:"telemetryId"`
	UserID      string      `json:"userId"`
	Application Application `json:"application"`
	Location    Location    `json:"location"`
}

// Application details
type Application struct {
	Version   string   `json:"version"`
	GitCommit string   `json:"gitCommit"`
	BuildDate string   `json:"buildDate"`
	Platform  string   `json:"platform"`
	Arch      string   `json:"arch"`
	GoLang    string   `json:"goLang"`
	RunningIn string   `json:"runningIn"`
	Uptime    uint64   `json:"uptime"`
	Gateways  []string `json:"gateways"`
	Handlers  []string `json:"handlers"`
}

// Location details
type Location struct {
	City    string `json:"city"`
	Region  string `json:"region"`
	Country string `json:"country"`
}
