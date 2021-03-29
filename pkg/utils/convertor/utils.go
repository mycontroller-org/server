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
		switch strings.ToLower(strings.TrimSpace(fmt.Sprintf("%v", data))) {
		case "true", "1", "on", "enable":
			return true

		case "false", "0", "off", "disable":
			return false

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
		strValue := strings.TrimSpace(fmt.Sprintf("%v", data))
		parsedValue, err := strconv.ParseFloat(strValue, 64)
		if err != nil {
			return 0
		}
		return parsedValue
	}
	return value
}

// ToInteger converts interface to int64
func ToInteger(data interface{}) int64 {
	value, ok := data.(int64)
	if !ok {
		strValue := strings.TrimSpace(fmt.Sprintf("%v", data))
		parsedValue, err := strconv.ParseInt(strValue, 10, 64)
		if err != nil {
			return 0
		}
		return parsedValue
	}
	return value
}
