package beholder

import (
	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
)

// NewChipIngressHeaderProvider creates a new chipingress.HeaderProvider
// using the same auth logic as NewGRPCClient.
// Returns nil provider (no error) if no auth is configured.
func NewChipIngressHeaderProvider(cfg Config) (chipingress.HeaderProvider, error) {
	auth, _, err := newRotatingAuthFromConfig(cfg)
	if err != nil {
		return nil, err
	}
	if auth != nil {
		return auth, nil
	}

	if len(cfg.AuthHeaders) > 0 {
		return NewStaticAuth(cfg.AuthHeaders, !cfg.InsecureConnection), nil
	}

	return nil, nil
}
