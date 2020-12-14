package json

import (
	"io"

	jsoniter "github.com/json-iterator/go"
)

var jsonIterator = jsoniter.ConfigCompatibleWithStandardLibrary

// Marshal adapter
func Marshal(data interface{}) ([]byte, error) {
	return jsonIterator.Marshal(data)
}

// MarshalIndent adapter
func MarshalIndent(data interface{}, prefix, indent string) ([]byte, error) {
	return jsonIterator.MarshalIndent(data, prefix, indent)
}

// MarshalToString adapter
func MarshalToString(data interface{}) (string, error) {
	return jsonIterator.MarshalToString(data)
}

// Unmarshal adapter
func Unmarshal(data []byte, v interface{}) error {
	return jsonIterator.Unmarshal(data, v)
}

// NewEncoder adapter
func NewEncoder(writer io.Writer) *jsoniter.Encoder {
	return jsonIterator.NewEncoder(writer)
}

// NewDecoder adapter
func NewDecoder(reader io.Reader) *jsoniter.Decoder {
	return jsonIterator.NewDecoder(reader)
}
