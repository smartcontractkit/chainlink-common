package config

import (
	"fmt"
	"testing"

	"github.com/smartcontractkit/chainlink-relay/core/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetConfig(t *testing.T) {
	// set env vars
	test.MockSetRequiredConfigs(t, Required.Core)
	test.MockSetRequiredConfigs(t, Required.Webhook)

	cfg, err := GetConfig()
	require.NoError(t, err)

	// test a few variables from each config
	assert.Equal(t, test.MockTestEnv, cfg.EthereumURL()) //required
	assert.Equal(t, false, cfg.Dev())
	assert.Equal(t, test.MockTestEnv, cfg.ICSecret()) // required
	assert.Equal(t, test.MockTestEnv, cfg.CIKey())    // required
}

func TestGetConfig_Fail_MissingEnvVar(t *testing.T) {
	_, err := GetConfig()
	assert.EqualError(t, err, fmt.Sprintf("Required env var: %s not found", Required.Core[0]))
}
