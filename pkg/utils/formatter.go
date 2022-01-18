package utils

import (
	"bytes"
	"encoding/gob"
	"reflect"
	"sync"
	"time"

	"github.com/fatih/structs"
	"github.com/mitchellh/mapstructure"
	json "github.com/mycontroller-org/server/v2/pkg/json"
	"github.com/mycontroller-org/server/v2/pkg/types"
	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
	dataRepositoryTY "github.com/mycontroller-org/server/v2/pkg/types/data_repository"
	fieldTY "github.com/mycontroller-org/server/v2/pkg/types/field"
	firmwareTY "github.com/mycontroller-org/server/v2/pkg/types/firmware"
	nodeTY "github.com/mycontroller-org/server/v2/pkg/types/node"
	"github.com/mycontroller-org/server/v2/pkg/types/schedule"
	scheduleTY "github.com/mycontroller-org/server/v2/pkg/types/schedule"
	sourceTY "github.com/mycontroller-org/server/v2/pkg/types/source"
	taskTY "github.com/mycontroller-org/server/v2/pkg/types/task"
	"github.com/mycontroller-org/server/v2/pkg/utils/convertor"
	gatewayTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
	handlerTY "github.com/mycontroller-org/server/v2/plugin/handler/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// call registerTypes func only once
var registerTypesInitOnce sync.Once

func registerTypes() {
	gob.Register([]interface{}{})
	gob.Register(map[string]interface{}{})
	gob.Register(map[interface{}]interface{}{})
	gob.Register(cmap.CustomMap{})
	gob.Register(cmap.CustomStringMap{})
	gob.Register(fieldTY.Field{})
	gob.Register(fieldTY.Payload{})
	gob.Register(fieldTY.PayloadFormatter{})
	gob.Register(taskTY.Config{})
	gob.Register(taskTY.Dampening{})
	gob.Register(taskTY.EventFilter{})
	gob.Register(taskTY.EvaluationConfig{})
	gob.Register(taskTY.Rule{})
	gob.Register(taskTY.Conditions{})
	gob.Register(taskTY.State{})
	gob.Register(gatewayTY.Config{})
	gob.Register(nodeTY.Node{})
	gob.Register(sourceTY.Source{})
	gob.Register(primitive.A{})
	gob.Register(handlerTY.Config{})
	gob.Register(handlerTY.ResourceData{})
	gob.Register(scheduleTY.Config{})
	gob.Register(scheduleTY.Validity{})
	gob.Register(scheduleTY.DateRange{})
	gob.Register(scheduleTY.TimeRange{})
	gob.Register(scheduleTY.CustomVariableConfig{})
	gob.Register(scheduleTY.State{})
	gob.Register(dataRepositoryTY.Config{})
	gob.Register(firmwareTY.Firmware{})
	gob.Register(firmwareTY.FileConfig{})
	gob.Register(firmwareTY.FirmwareBlock{})
	gob.Register(types.State{})
	gob.Register(time.Time{})
}

// ToStruct converts bytes to target struct
func ToStruct(data []byte, out interface{}) error {
	return json.Unmarshal(data, out)
}

// StructToByte converts interface to []byte
func StructToByte(data interface{}) ([]byte, error) {
	registerTypesInitOnce.Do(registerTypes)
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ByteToStruct converts []byte to interface
func ByteToStruct(data []byte, out interface{}) error {
	registerTypesInitOnce.Do(registerTypes)
	var buf bytes.Buffer
	_, err := buf.Write(data)
	if err != nil {
		return err
	}
	dec := gob.NewDecoder(&buf)
	return dec.Decode(out)
}

// ByteToMap converts []byte map[string]interface{}
func ByteToMap(data []byte) (map[string]interface{}, error) {
	out := make(map[string]interface{})
	return out, ByteToStruct(data, out)
}

// mapToStructDecodeHookFunc will be called on MapToStruct
func mapToStructDecodeHookFunc(fromType reflect.Type, toType reflect.Type, value interface{}) (interface{}, error) {
	// zap.L().Info("mapToStructDecodeHookFunc decoder", zap.String("fromType", fromType.String()), zap.String("toType", toType.String()), zap.Any("value", value))
	switch toType {
	case reflect.TypeOf(time.Time{}):
		value, err := time.Parse(time.RFC3339Nano, value.(string))
		if err != nil {
			return nil, err
		}
		return value, nil

	case reflect.TypeOf(schedule.CustomDate{}):
		cd := schedule.CustomDate{}
		err := cd.Unmarshal(convertor.ToString(value))
		if err != nil {
			return nil, err
		}
		return &cd, nil

	case reflect.TypeOf(schedule.CustomTime{}):
		ct := schedule.CustomTime{}
		err := ct.Unmarshal(convertor.ToString(value))
		if err != nil {
			return nil, err
		}
		return &ct, nil

	}
	return value, nil
}

// MapToStruct converts string to struct
func MapToStruct(tagName string, in map[string]interface{}, out interface{}) error {
	if tagName == "" {
		return mapstructure.Decode(in, out)
	}
	cfg := &mapstructure.DecoderConfig{
		TagName:          tagName,
		WeaklyTypedInput: true,
		DecodeHook:       mapToStructDecodeHookFunc,
		Result:           out,
	}
	decoder, err := mapstructure.NewDecoder(cfg)
	if err != nil {
		return err
	}
	return decoder.Decode(in)
}

// StructToMap converts struct to a map
func StructToMap(data interface{}) map[string]interface{} {
	return structs.Map(data)
}

// ToDuration converts the string duration to time.Duration, if failed returns the default
func ToDuration(duration string, defaultDuration time.Duration) time.Duration {
	parsedDuration, err := time.ParseDuration(duration)
	if err != nil {
		return defaultDuration
	}
	return parsedDuration
}

// ValidDuration verifies the duration, if it hits error returns the default duration
func ValidDuration(duration, defaultDuration string) string {
	if duration == "" {
		return defaultDuration
	}
	_, err := time.ParseDuration(duration)
	if err == nil {
		return duration
	}
	return defaultDuration
}
