package yamlutils

import (
	"encoding/base64"

	"gopkg.in/yaml.v3"
)

// UnmarshalBase64Yaml converts base64 data into given interface
func UnmarshalBase64Yaml(base64String string, out interface{}) error {
	if yamlBytes, err := base64.StdEncoding.DecodeString(base64String); err == nil {
		return yaml.Unmarshal(yamlBytes, out)
	}
	return yaml.Unmarshal([]byte(base64String), out)
}

// MarshalBase64Yaml converts interface to base64
func MarshalBase64Yaml(in interface{}) (string, error) {
	yamlBytes, err := yaml.Marshal(in)
	if err != nil {
		return "", err
	}
	// base64string := base64.StdEncoding.EncodeToString(yamlBytes)
	// return base64string, nil
	return string(yamlBytes), nil
}
