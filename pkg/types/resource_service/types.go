package types

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	"github.com/mycontroller-org/server/v2/pkg/utils/convertor"
)

// Resource type details
const (
	TypeGateway        = "gateway"
	TypeNode           = "node"
	TypeTask           = "task"
	TypeHandler        = "handler"
	TypeScheduler      = "scheduler"
	TypeFirmware       = "firmware"
	TypeResourceAction = "resource_action"
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
	outVal := reflect.ValueOf(out)
	if outVal.Kind() != reflect.Ptr {
		return errors.New("out argument must be a pointer")
	}

	sliceVal := outVal.Elem()

	switch sliceVal.Type().Kind() {
	case reflect.Struct, reflect.Map:
		mapData, ok := e.Data.(map[string]interface{})
		if !ok {
			return fmt.Errorf("data is not in map[string]interface{} type, received:%T, out:%T", e.Data, out)
		}
		return utils.MapToStruct(utils.TagNameJSON, mapData, out)

	case reflect.Slice:
		elementType := sliceVal.Type().Elem()
		switch out.(type) {
		case *[]string:
			if sliceData, ok := e.Data.([]interface{}); ok {
				for index, data := range sliceData {
					if sliceVal.Len() == index { // slice is full
						newElem := reflect.New(elementType)
						sliceVal = reflect.Append(sliceVal, newElem.Elem())
						sliceVal = sliceVal.Slice(0, sliceVal.Cap())
					}
					sliceVal.Index(index).Set(reflect.ValueOf(convertor.ToString(data)))
				}
				outVal.Elem().Set(sliceVal.Slice(0, len(sliceData)))
				return nil
			}
			return fmt.Errorf("data is not in []interface{} type, received:%T, out:%T", e.Data, out)

		}

	case reflect.String, reflect.Interface:
		switch out.(type) {
		case *string:
			if outVal.Elem().CanSet() {
				outVal.Elem().Set(reflect.ValueOf(convertor.ToString(e.Data)))
				return nil
			}
			return errors.New("out field cannot be set")

		}
	}
	return fmt.Errorf("unknown type received, received:%T, out:%T", e.Data, out)
}
