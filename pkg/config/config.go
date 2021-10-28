package config

import (
	"fmt"
	"os"
)

// called in each subset of configs
func ValidateRequired(vars []string) error {
	// validation
	for _, key := range vars {
		if env := os.Getenv(key); env == "" {
			return fmt.Errorf("Required env var: %s not found", key)
		}
	}
	return nil
}
