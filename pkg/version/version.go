package version

import (
	"runtime"

	"github.com/mycontroller-org/backend/v2/pkg/utils"
)

var (
	gitCommit string
	version   string
	buildDate string
)

const (
	EnvironmentDocker     = "docker"
	EnvironmentKubernetes = "kubernetes"
	EnvironmentBareMetal  = "bare_metal"
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

// docker creates a .dockerenv file at the root of the directory tree inside the container.
// if this file exists then the viewer is running from inside a container so return true
// With the default configuration, Kubernetes will mount the serviceaccount secrets into pods.
func RunningIn() string {
	if utils.IsFileExists("/.dockerenv") {
		return EnvironmentDocker
	} else if utils.IsDirExists("/var/run/secrets/kubernetes.io") {
		return EnvironmentKubernetes
	}
	return EnvironmentBareMetal
}
