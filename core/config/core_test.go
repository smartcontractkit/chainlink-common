package config

import (
	"fmt"
	"testing"

	"github.com/smartcontractkit/chainlink-relay/core/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCoreConfig_Success(t *testing.T) {
	test.MockSetRequiredConfigs(t, Required.Core)

	// create core configs
	cfg, err := NewCoreConfig()
	require.NoError(t, err)

	// test a few variables
	assert.Equal(t, test.MockTestEnv, cfg.KeystorePassword()) // required param
	assert.Equal(t, false, cfg.Mock())                        // optional parameter with default
}

func TestNewCoreConfig_MissingRequired(t *testing.T) {
	// create core configs
	_, err := NewCoreConfig()
	require.EqualError(t, err, fmt.Sprintf("Required env var: %s not found", Required.Core[0]))
}

func TestNewCoreConfig_Fail_GeneralConfig(t *testing.T) {
	// https://github.com/smartcontractkit/chainlink/blob/ea7dab1aae73bd129e9e506dcd5498ecc9fc3dba/core/store/config/general_config.go#L298
	t.Setenv("P2P_ANNOUNCE_PORT", "1000")
	test.MockSetRequiredConfigs(t, Required.Core)

	// briefly validate that core config validate functions as expected
	_, err := GetConfig()
	require.EqualError(t, err, "P2P_ANNOUNCE_PORT was given as 1000 but P2P_ANNOUNCE_IP was unset. You must also set P2P_ANNOUNCE_IP if P2P_ANNOUNCE_PORT is set")
}
