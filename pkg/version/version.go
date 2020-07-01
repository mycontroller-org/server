package version

import (
	"fmt"
	"runtime"
)

var (
	version   string
	buildDate string
)

// Version holds version data
type Version struct {
	MyController string `json:"myController"`
	BuildDate    string `json:"buildDate"`
	GoLang       string `json:"goLang"`
}

// Get returns the Version object
func Get() Version {
	return Version{
		MyController: version,
		BuildDate:    buildDate,
		GoLang:       runtime.Version(),
	}
}

func (v Version) String() string {
	return fmt.Sprintf(
		"Version(MyController='%v', BuildDate='%v', GoLang='%v')",
		v.MyController, v.BuildDate, v.GoLang,
	)
}
