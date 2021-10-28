package webhook

import (
	"fmt"
	"testing"

	"github.com/smartcontractkit/chainlink-relay/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWebhookConfig_Success(t *testing.T) {
	test.MockSetRequiredConfigs(t, RequiredConfigs)

	// create core configs
	cfg, err := NewWebhookConfig()
	require.NoError(t, err)

	// test a few variables
	assert.Equal(t, test.MockTestEnv, cfg.ICKey())    // required param
	assert.Equal(t, test.MockTestEnv, cfg.CISecret()) // required param
}

func TestNewWebhookConfig_MissingRequired(t *testing.T) {
	// create core configs
	_, err := NewWebhookConfig()
	require.EqualError(t, err, fmt.Sprintf("Required env var: %s not found", RequiredConfigs[0]))
}
