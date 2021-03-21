package version

import (
	"runtime"
)

var (
	gitCommit string
	version   string
	buildDate string
)

// Version holds version data
type Version struct {
	Version   string `json:"version"`
	GitCommit string `json:"gitCommit"`
	BuildDate string `json:"buildDate"`
	GoLang    string `json:"goLang"`
	Platform  string `json:"platform"`
	Arch      string `json:"arch"`
}

// Get returns the Version object
func Get() Version {
	return Version{
		GitCommit: gitCommit,
		Version:   version,
		BuildDate: buildDate,
		GoLang:    runtime.Version(),
		Platform:  runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
}
