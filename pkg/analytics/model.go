package analytics

// Payload of the analytics
type Payload struct {
	APIVersion  string      `json:"apiVersion"`
	AnalyticsID string      `json:"analyticsId"`
	UserID      string      `json:"userId"`
	Application Application `json:"application"`
	Location    Location    `json:"location"`
}

// Application details
type Application struct {
	Version   string   `json:"version"`
	GitCommit string   `json:"gitCommit"`
	Platform  string   `json:"platform"`
	Arch      string   `json:"arch"`
	GoLang    string   `json:"goLang"`
	RunningIn string   `json:"runningIn"`
	Gateways  []string `json:"gateways"`
	Handlers  []string `json:"handlers"`
}

// Location details
type Location struct {
	City    string `json:"city"`
	Region  string `json:"region"`
	Country string `json:"country"`
}
