package config

import (
	"os"
	"strings"

	"github.com/smartcontractkit/chainlink/core/store/config"
)

type CoreConfig interface {
	config.GeneralConfig
	KeystorePassword() string
	Mock() bool
}

// NewCoreConfig fetches configs defined by the chainlink core
//  - some additional parameters appended
//  - many parameters are repurposed/reused
// CHAINLINK_PORT: defines the lite client port
// DATABASE_URL: psql db URL
// CLIENT_NODE_URL: defines the actual chainlink node URL
// KEYSTORE_PASSWORD
// ETH_URL: websocket URL to respective blockchain
// ETH_HTTP_URL: http URL to respective blockchain
func NewCoreConfig() (CoreConfig, error) {
	cfg := config.NewGeneralConfig()
	if err := cfg.Validate(); err != nil {
		return coreConfig{}, err
	}

	return coreConfig{cfg}, ValidateRequired(Required.Core)
}

type coreConfig struct {
	config.GeneralConfig
}

func (cc coreConfig) KeystorePassword() string {
	return os.Getenv("KEYSTORE_PASSWORD")
}

// Not required -------------------
func (cc coreConfig) Mock() bool {
	return strings.ToLower(os.Getenv("MOCK_SERVICE")) == "true"
}
