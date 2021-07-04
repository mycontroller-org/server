package storage

import (
	"github.com/mycontroller-org/server/v2/plugin/database/storage/memory"
)

func init() {
	Register(memory.PluginMemory, memory.NewClient)
}
