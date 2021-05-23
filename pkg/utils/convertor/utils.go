package convertor

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

// ToString converts interface to string
func ToString(data interface{}) string {
	if data == nil {
		return ""
	}
	switch dataNew := data.(type) {
	case int8, int16, int32, int64, int,
		uint8, uint16, uint32, uint64, uint:
		return fmt.Sprintf("%d", dataNew)

	case float32, float64:
		return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%f", dataNew), "0"), ".")

	case bool:
		return fmt.Sprintf("%t", dataNew)

	case string:
		return fmt.Sprintf("%v", dataNew)

	case []byte:
		return string(dataNew)

	default:
		b, err := json.Marshal(data)
		if err != nil {
			zap.L().Error("error on converting to string", zap.Any("data", data), zap.Error(err))
			return fmt.Sprintf("%v", data)
		}
		return string(b)
	}
}

// ToBool converts interface to boolean
func ToBool(data interface{}) bool {
	value, ok := data.(bool)
	if ok {
		return value
	}
	switch ToString(data) {
	case "true", "1", "on", "enable":
		return true

	case "false", "0", "off", "disable":
		return false

	default:
		return false
	}
}

// ToFloat converts interface to float64
func ToFloat(data interface{}) float64 {
	if value, ok := data.(float64); ok {
		return value
	}
	strValue := ToString(data)
	parsedValue, err := strconv.ParseFloat(strValue, 64)
	if err != nil {
		return 0
	}
	return parsedValue
}

// ToInteger converts interface to int64
func ToInteger(data interface{}) int64 {
	if value, ok := data.(int64); ok {
		return value
	}

	floatValue := ToFloat(data)
	return int64(floatValue)
}
