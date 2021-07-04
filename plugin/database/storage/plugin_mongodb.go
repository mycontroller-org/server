package storage

import (
	mongo "github.com/mycontroller-org/server/v2/plugin/database/storage/mongodb"
)

func init() {
	Register(mongo.PluginMongoDB, mongo.NewClient)
}
