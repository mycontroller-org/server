package commands

import (
	"os"
	"strings"
)

const (
	CommandVersion        = "version"
	CommandGenerateConfig = "config"

	CallerServer  = "server"
	CallerGateway = "gateway"
	CallerHandler = "handler"
)

// PrintVersion prints the version and exits
func ExecuteCommand(caller string) {
	if len(os.Args) > 1 {
		switch strings.ToLower(os.Args[1]) {
		case CommandVersion:
			PrintVersion(caller)

		case CommandGenerateConfig:
			PrintDefaultConfig(caller)

		}
		os.Exit(0)
	}
}
