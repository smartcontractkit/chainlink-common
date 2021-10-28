package webhook

import (
	"os"

	"github.com/smartcontractkit/chainlink-relay/pkg/config"
)

var RequiredConfigs []string = []string{"IC_ACCESSKEY", "IC_SECRET", "CI_ACCESSKEY", "CI_SECRET"}

// WebhookConfig is the interface for retreiving webhook related configs
type WebhookConfig interface {
	ICKey() string
	ICSecret() string
	CIKey() string
	CISecret() string
}

// NewWebhookConfig provides the retrieved configuration
func NewWebhookConfig() (WebhookConfig, error) {
	return webhookConfig{}, config.ValidateRequired(RequiredConfigs)
}

type webhookConfig struct{}

func (wc webhookConfig) ICKey() string {
	return os.Getenv("IC_ACCESSKEY")
}
func (wc webhookConfig) ICSecret() string {
	return os.Getenv("IC_SECRET")
}
func (wc webhookConfig) CIKey() string {
	return os.Getenv("CI_ACCESSKEY")
}
func (wc webhookConfig) CISecret() string {
	return os.Getenv("CI_SECRET")
}
