package mercury

import (
	"encoding/json"
)

type OffchainConfig struct {
	ExpirationWindow uint32 // Integer number of seconds
	BaseUSDFeeCents  uint32
}

func DecodeOffchainConfig(b []byte) (o OffchainConfig, err error) {
	// TODO: consider protobuf for better efficiency
	err = json.Unmarshal(b, &o)
	return
}

func (c OffchainConfig) Encode() ([]byte, error) {
	return json.Marshal(c)
}
