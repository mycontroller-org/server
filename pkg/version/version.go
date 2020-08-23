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
	GitCommit string `json:"gitCommit"`
	Version   string `json:"version"`
	BuildDate string `json:"buildDate"`
	GoLang    string `json:"goLang"`
}

// Get returns the Version object
func Get() Version {
	return Version{
		GitCommit: gitCommit,
		Version:   version,
		BuildDate: buildDate,
		GoLang:    runtime.Version(),
	}
}
