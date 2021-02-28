package utils

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fatih/structs"
	"github.com/mitchellh/mapstructure"
	json "github.com/mycontroller-org/backend/v2/pkg/json"
	"github.com/mycontroller-org/backend/v2/pkg/model/cmap"
	fieldML "github.com/mycontroller-org/backend/v2/pkg/model/field"
	gatewayML "github.com/mycontroller-org/backend/v2/pkg/model/gateway"
	nodeML "github.com/mycontroller-org/backend/v2/pkg/model/node"
	taskML "github.com/mycontroller-org/backend/v2/pkg/model/task"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"go.uber.org/zap"
)

// call registerTypes func only once
var registerTypesInitOnce sync.Once

func registerTypes() {
	gob.Register([]interface{}{})
	gob.Register(map[string]interface{}{})
	gob.Register(map[interface{}]interface{}{})
	gob.Register(cmap.CustomMap{})
	gob.Register(cmap.CustomStringMap{})
	gob.Register(fieldML.Field{})
	gob.Register(taskML.Config{})
	gob.Register(gatewayML.Config{})
	gob.Register(nodeML.Node{})
	gob.Register(primitive.A{})
}

// ToString converts interface to string
func ToString(data interface{}) string {
	if data == nil {
		return ""
	}
	switch data.(type) {
	case int, float64, string, bool:
		return fmt.Sprintf("%v", data)
	default:
		b, err := json.Marshal(data)
		if err != nil {
			zap.L().Error("Failed to convert to string", zap.Any("data", data), zap.Error(err))
			return fmt.Sprintf("%v", data)
		}
		return string(b)
	}
}

// ToBool converts interface to boolean
func ToBool(data interface{}) bool {
	value, ok := data.(bool)
	if !ok {
		switch strings.ToLower(fmt.Sprintf("%v", data)) {
		case "true", "1", "on", "enabled":
			return true
		default:
			return false
		}
	}
	return value
}

// ToFloat converts interface to float64
func ToFloat(data interface{}) float64 {
	value, ok := data.(float64)
	if !ok {
		strValue := fmt.Sprintf("%v", data)
		parsedValue, err := strconv.ParseFloat(strValue, 64)
		if err != nil {
			return 0
		}
		return parsedValue
	}
	return value
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

// MapToStruct converts string to struct
func MapToStruct(tagName string, in map[string]interface{}, out interface{}) error {
	if tagName == "" {
		return mapstructure.Decode(in, out)
	}
	cfg := &mapstructure.DecoderConfig{TagName: tagName, Result: out}
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
