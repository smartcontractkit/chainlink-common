package json

import (
	"bytes"
)

// UnmarshalJson unmarshals JSON data using our custom decoder that handles
// all numbers as strings to preserve numeric precision
func UnmarshalJson(data []byte, v any) error {
	decoder := newDecoder(bytes.NewReader(data))
	return decoder.Decode(v)
}

// MarshalJson marshals data to JSON using our custom encoder that converts
// all numbers to strings to preserve numeric precision
func MarshalJson(v any) ([]byte, error) {
	var buf bytes.Buffer
	encoder := newEncoder(&buf)
	if err := encoder.Encode(v); err != nil {
		return nil, err
	}
	// Remove trailing newline added by encoder
	data := buf.Bytes()
	if len(data) > 0 && data[len(data)-1] == '\n' {
		data = data[:len(data)-1]
	}
	return data, nil
}
