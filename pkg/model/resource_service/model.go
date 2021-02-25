package model

import (
	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
)

// Resource type details
const (
	TypeGateway                  = "gateway"
	TypeTask                     = "task"
	TypeNotifyHandler            = "notify_handler"
	TypeScheduler                = "scheduler"
	TypeResourceActionBySelector = "resource_action_by_selector"
)

// Command details
const (
	CommandUpdate      = "update"
	CommandUpdateState = "updateState"
	CommandGet         = "get"
	CommandSet         = "set"
	CommandAdd         = "add"
	CommandRemove      = "remove"
	CommandEnable      = "enable"
	CommandDisable     = "disable"
	CommandStart       = "start"
	CommandStop        = "stop"
	CommandReload      = "reload"
	CommandLoadAll     = "loadAll"
	CommandUnloadAll   = "unloadAll"
)

// Event details
type Event struct {
	Type         string
	Command      string
	ReplyCommand string
	ReplyTopic   string
	ID           string
	Labels       cmap.CustomStringMap
	Data         []byte `json:"-"` // ignore this field on logging
	Error        string
}

// SetData updates data in []byte format
func (e *Event) SetData(data interface{}) error {
	if data == nil {
		return nil
	}
	bytes, err := utils.StructToByte(data)
	if err != nil {
		return err
	}
	e.Data = bytes
	return nil
}

// ToStruct converts data to target interface
func (e *Event) ToStruct(out interface{}) error {
	return utils.ByteToStruct(e.Data, out)
}
