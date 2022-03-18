package customtypes

import (
	"strings"

	"github.com/mycontroller-org/server/v2/pkg/utils/convertor"
)

// StringData keeps the value as string
type StringData string

// converts to string
func (sd StringData) String() string {
	return string(sd)
}

// UnmarshalJSON custom implementation
func (sd *StringData) UnmarshalJSON(data []byte) error {
	return sd.unmarshal(string(data))
}

// UnmarshalYAML custom implementation
func (sd *StringData) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var stringTime string
	err := unmarshal(&stringTime)
	if err != nil {
		return nil
	}
	return sd.unmarshal(stringTime)
}

func (sd *StringData) unmarshal(rawString string) error {
	stringData := strings.Trim(rawString, `"`)
	value := convertor.ToString(stringData)
	*sd = StringData(value)
	return nil
}
