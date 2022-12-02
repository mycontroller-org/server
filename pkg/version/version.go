package version

import (
	"fmt"
	"runtime"

	"github.com/shirou/gopsutil/v3/host"
)

var (
	gitCommit string
	version   string
	buildDate string
)

// Version holds version data
type Version struct {
	Version   string `json:"version" yaml:"version"`
	GitCommit string `json:"gitCommit" yaml:"gitCommit"`
	BuildDate string `json:"buildDate" yaml:"buildDate"`
	GoLang    string `json:"goLang" yaml:"goLang"`
	Platform  string `json:"platform" yaml:"platform"`
	Arch      string `json:"arch" yaml:"arch"`
	HostID    string `json:"hostId" yaml:"hostId"`
}

// Get returns the Version object
func Get() Version {
	hostId, err := host.HostID()
	if err != nil {
		hostId = err.Error()
	}
	return Version{
		GitCommit: gitCommit,
		Version:   version,
		BuildDate: buildDate,
		GoLang:    runtime.Version(),
		Platform:  runtime.GOOS,
		Arch:      runtime.GOARCH,
		HostID:    hostId,
	}
}

// String returns the values as string
func (v Version) String() string {
	return fmt.Sprintf("{version:%s, buildDate:%s, gitCommit:%s, goLang:%s, platform:%s, arch:%s, hostId:%s}",
		v.Version, v.BuildDate, v.GitCommit, v.GoLang, v.Platform, v.Arch, v.HostID)
}
