package printversion

import (
	"fmt"
	"os"
	"strings"

	"github.com/mycontroller-org/server/v2/pkg/json"
	"github.com/mycontroller-org/server/v2/pkg/version"
)

// PrintVersion prints the version and exits
func PrintVersion() {
	if len(os.Args) > 1 && strings.ToLower(os.Args[1]) == "version" {
		ver := version.Get()
		verBytes, err := json.Marshal(ver)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(string(verBytes))
		os.Exit(0)
	}
}
