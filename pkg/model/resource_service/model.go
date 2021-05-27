package model

import (
	"errors"
	"reflect"

	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
	"github.com/mycontroller-org/backend/v2/pkg/utils"
	"github.com/mycontroller-org/backend/v2/pkg/utils/convertor"
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
	Type         string               `json:"type"`
	Command      string               `json:"command"`
	ReplyCommand string               `json:"replyCommand"`
	ReplyTopic   string               `json:"replyTopic"`
	ID           string               `json:"id"`
	Labels       cmap.CustomStringMap `json:"lables"`
	Data         interface{}          `json:"data"`
	Error        string               `json:"error"`
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

// LoadData loads the data to given interface
func (e *ServiceEvent) LoadData(out interface{}) error {

	switch out.(type) {
	case string:
		out = convertor.ToString(e.Data)
		return nil

	case []string:
		if stringSlice, ok := e.Data.([]string); ok {
			out = stringSlice
			return nil
		}
	}

	mapData, ok := e.Data.(map[string]interface{})
	if !ok {
		return errors.New("data is not in map[string]interface{} type")
	}
	return utils.MapToStruct(utils.TagNameJSON, mapData, out)
}
