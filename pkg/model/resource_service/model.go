package model

import (
	"reflect"

	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
)

// Resource type details
const (
	TypeGateway                  = "gateway"
	TypeNode                     = "node"
	TypeTask                     = "task"
	TypeHandler                  = "handler"
	TypeScheduler                = "scheduler"
	TypeResourceActionBySelector = "resource_action_by_selector"
	TypeFirmware                 = "firmware"
)

// Command details
const (
	CommandUpdate        = "update"
	CommandUpdateState   = "updateState"
	CommandGet           = "get"
	CommandList          = "list"
	CommandGetIds        = "getIds"
	CommandSet           = "set"
	CommandAdd           = "add"
	CommandRemove        = "remove"
	CommandEnable        = "enable"
	CommandDisable       = "disable"
	CommandStart         = "start"
	CommandStop          = "stop"
	CommandReload        = "reload"
	CommandLoadAll       = "loadAll"
	CommandUnloadAll     = "unloadAll"
	CommandBlocks        = "blocks"
	CommandFirmwareState = "firmwareState"
	CommandSetLabel      = "setLabel"
)

// ServiceEvent details
type ServiceEvent struct {
	Type         string
	Command      string
	ReplyCommand string
	ReplyTopic   string
	ID           string
	Labels       cmap.CustomStringMap
	// 	Data         []byte `json:"-"` // ignore this field on logging
	Data  interface{} `json:"-"` // ignore this field on logging
	Error string
}

func (e *ServiceEvent) SetData(data interface{}) {
	if reflect.ValueOf(data).Kind() == reflect.Ptr {
		e.Data = reflect.ValueOf(data).Elem().Interface()
		return
	}
	e.Data = data
}

func (e *ServiceEvent) GetData() interface{} {
	if e.Data == nil {
		return nil
	}
	if reflect.ValueOf(e.Data).Kind() == reflect.Ptr {
		return reflect.ValueOf(e.Data).Elem().Interface()
	}
	return e.Data
}
