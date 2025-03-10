package version

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/shirou/gopsutil/v4/host"
)

var (
	gitCommit string

	// if we use go build, go run, server failed to start. needs version details
	// adding a static version here, this will be replaced on actual build
	// TODO: create a make file to address this issue on development time
	version string = "2.1.1-devel"

	buildDate string

	runOnce  sync.Once
	_version Version
)

// Version holds version data
type Version struct {
	Version   string `json:"version" yaml:"version"`
	GitCommit string `json:"gitCommit" yaml:"gitCommit"`
	BuildDate string `json:"buildDate" yaml:"buildDate"`
	GoVersion string `json:"goVersion" yaml:"goVersion"`
	Compiler  string `json:"compiler" yaml:"compiler"`
	Platform  string `json:"platform" yaml:"platform"`
	Arch      string `json:"arch" yaml:"arch"`
	HostID    string `json:"hostId" yaml:"hostId"`
}

// Get returns the Version object
func Get() Version {
	runOnce.Do(func() {
		hostId, err := host.HostID()
		if err != nil {
			hostId = err.Error()
		}
		_version = Version{
			GitCommit: gitCommit,
			Version:   version,
			BuildDate: buildDate,
			GoVersion: runtime.Version(),
			Compiler:  runtime.Compiler,
			Platform:  runtime.GOOS,
			Arch:      runtime.GOARCH,
			HostID:    hostId,
		}
	})
	return _version
}

// String returns the values as string
func (v Version) String() string {
	return fmt.Sprintf("{version:%s, buildDate:%s, gitCommit:%s, goVersion:%s, compiler:%s, platform:%s, arch:%s, hostId:%s}",
		v.Version, v.BuildDate, v.GitCommit, v.GoVersion, v.Compiler, v.Platform, v.Arch, v.HostID)
}
