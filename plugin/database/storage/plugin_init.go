package storage

import (
	"github.com/mycontroller-org/server/v2/plugin/database/storage/memory"
	mongo "github.com/mycontroller-org/server/v2/plugin/database/storage/mongodb"
)

func init() {
	Register(memory.PluginMemory, memory.NewClient)
	Register(mongo.PluginMongoDB, mongo.NewClient)
}
