package types

import (
	"context"
)

// Deprecated: use services.Service
//
//go:fix inline
type Service interface {
	Name() string
	Start(context.Context) error
	Close() error
	Ready() error
	HealthReport() map[string]error
}
