package test

import (
	"os"
	"testing"
)

var (
	MockTestEnv = "http://testEnvVar.link"
)

func MockSetRequiredConfigs(t *testing.T, vars []string) {
	for _, key := range vars {
		// set env (clears after test is complete)
		t.Setenv(key, MockTestEnv) // this cannot be used if run in parallel (released in 1.17)
	}
}

func UnsetEIKeysSecrets() error {
	vars := []string{"IC_ACCESSKEY", "IC_SECRET", "CI_SECRET", "CI_ACCESSKEY"}
	for _, v := range vars {
		if err := os.Unsetenv(v); err != nil {
			return err
		}
	}
	return nil
}

type MockWebhookConfig struct {
	ICKeyStr    string
	ICSecretStr string
	CIKeyStr    string
	CISecretStr string
}

func (wc MockWebhookConfig) ICKey() string {
	return wc.ICKeyStr
}
func (wc MockWebhookConfig) ICSecret() string {
	return wc.ICSecretStr
}
func (wc MockWebhookConfig) CIKey() string {
	return wc.CIKeyStr
}
func (wc MockWebhookConfig) CISecret() string {
	return wc.CISecretStr
}
