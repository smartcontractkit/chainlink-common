package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/smartcontractkit/chainlink/core/logger"
)

// Required is an array of env vars that must be set for each config type
var Required = struct {
	Core    []string
	Webhook []string
}{
	Core:    []string{"DATABASE_URL", "KEYSTORE_PASSWORD"}, // default relay only requires WS URL
	Webhook: []string{"IC_ACCESSKEY", "IC_SECRET", "CI_ACCESSKEY", "CI_SECRET"},
}

// Config contains all of the general configs for the CL node and the specific webhook secrets
// CL core node configs will be used as much as possible to avoid specificying many different env parameters
type Config struct {
	CoreConfig
	WebhookConfig
}

// GetConfig fetches all the configs from the local .env file
func GetConfig() (Config, error) {
	log := logger.Default.Named("config")
	// overwrite existing env vars with parameters in env file
	if err := godotenv.Overload(".env"); err != nil {
		log.Info("Note: .env file not found using defaults and existing env vars")
	}

	core, err := NewCoreConfig()
	if err != nil {
		return Config{}, err
	}

	webhook, err := NewWebhookConfig()
	if err != nil {
		return Config{}, err
	}

	return Config{core, webhook}, nil
}

// ValidateRequired called in each subset of configs
func ValidateRequired(vars []string) error {
	// validation
	for _, key := range vars {
		if env := os.Getenv(key); env == "" {
			return fmt.Errorf("Required env var: %s not found", key)
		}
	}
	return nil
}
