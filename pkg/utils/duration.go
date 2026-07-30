package utils

import (
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
)

// Deprecated: use config.Duration
//
//go:fix inline
type Duration = config.Duration

// Deprecated: use config.NewDuration
//
//go:fix inline
func NewDuration(d time.Duration) (config.Duration, error) { return config.NewDuration(d) }

// Deprecated: use config.MustNewDuration
//
//go:fix inline
func MustNewDuration(d time.Duration) *config.Duration { return config.MustNewDuration(d) }
