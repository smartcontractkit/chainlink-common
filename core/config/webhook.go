package config

import "os"

// WebhookConfig is the interface for retreiving webhook related configs
type WebhookConfig interface {
	ICKey() string
	ICSecret() string
	CIKey() string
	CISecret() string
}

// NewWebhookConfig provides the retrieved configuration
func NewWebhookConfig() (WebhookConfig, error) {
	return webhookConfig{}, ValidateRequired(Required.Webhook)
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
