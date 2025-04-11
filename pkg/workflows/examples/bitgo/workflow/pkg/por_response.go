package pkg

import (
	"encoding/json"
	"fmt"
	"time"
)

type ReserveInfo struct {
	LastUpdated  time.Time `json:"lastUpdated"`
	TotalReserve string    `json:"totalReserve"`
}

type PorResponse struct {
	DataSignature string `json:"dataSignature"`
	Ripcord       bool   `json:"ripcord"`
	Data          string `json:"data"`
}

func UnescapeJSONString(input string) (string, error) {
	var unescaped string
	err := json.Unmarshal([]byte(`"`+input+`"`), &unescaped)
	if err != nil {
		return "", fmt.Errorf("failed to unescape string: %w", err)
	}
	return unescaped, nil
}
