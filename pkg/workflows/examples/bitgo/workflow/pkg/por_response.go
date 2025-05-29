package pkg

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

type RawReserveInfo struct {
	LastUpdated  time.Time       `json:"lastUpdated" consensus:"median"`
	TotalReserve decimal.Decimal `json:"totalReserve" consensus:"median"`
}

type ReserveInfo struct {
	LastUpdated  int64           `json:"lastUpdated" consensus:"median"`
	TotalReserve decimal.Decimal `json:"totalReserve" consensus:"median"`
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
