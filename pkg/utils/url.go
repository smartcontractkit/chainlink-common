package utils

import "github.com/smartcontractkit/chainlink-common/pkg/config"

// Deprecated: use config.URL
//
//go:fix inline
type URL = config.URL

// Deprecated: use config.ParseURL
//
//go:fix inline
func ParseURL(s string) (*config.URL, error) { return config.ParseURL(s) }

// Deprecated: use config.MustParseURL
//
//go:fix inline
func MustParseURL(s string) *config.URL { return config.MustParseURL(s) }
