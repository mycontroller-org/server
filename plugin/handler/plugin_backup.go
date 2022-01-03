//go:build server && !standalone
// +build server,!standalone

package handler

import (
	backup "github.com/mycontroller-org/server/v2/plugin/handler/backup"
)

func init() {
	Register(backup.PluginBackup, backup.NewBackupPlugin)
}
