package commands

import (
	"fmt"

	"github.com/mycontroller-org/server/v2/pkg/json"
	"github.com/mycontroller-org/server/v2/pkg/version"
)

// PrintVersion prints the version and exits
func PrintVersion(caller string) {
	ver := version.Get()
	verBytes, err := json.Marshal(ver)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(verBytes))
}
