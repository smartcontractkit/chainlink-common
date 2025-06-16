package sdk

import (
	"encoding/json"
)

func ParseJSON[T any](bytes []byte) (*T, error) {
	var result T
	err := json.Unmarshal(bytes, &result)
	return &result, err
}
