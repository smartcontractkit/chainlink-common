package loop_test

import (
	"testing"

	"github.com/hashicorp/go-plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop"
	sctest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/capability/standard/test"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

func TestPluginStandardCapabilities(t *testing.T) {
	t.Parallel()

	log := logger.Test(t)

	stopCh := newStopCh(t)
	test.PluginTest(t, loop.PluginStandardCapabilitiesName,
		&loop.StandardCapabilitiesLoop{
			Logger:       log,
			PluginServer: sctest.StandardCapabilitiesService{},
			BrokerConfig: loop.BrokerConfig{
				Logger: logger.Test(t),
				StopCh: stopCh}},
		func(t *testing.T, s loop.StandardCapabilities) {
			ctx := tests.Context(t)
			infos, err := s.Infos(ctx)
			assert.NoError(t, err)
			assert.Equal(t, 2, len(infos))
			assert.Equal(t, capabilities.CapabilityTypeAction, infos[0].CapabilityType)
			assert.Equal(t, capabilities.CapabilityTypeTarget, infos[1].CapabilityType)

			err = s.Initialise(ctx, "", nil, nil, nil, nil, nil, nil, nil)
			assert.NoError(t, err)
		})
}

func TestRunningStandardCapabilitiesPluginOutOfProcess(t *testing.T) {
	t.Parallel()
	ctx := tests.Context(t)
	stopCh := newStopCh(t)

	scs := newOutOfProcessStandardCapabilitiesService(t, true, stopCh)

	infos, err := scs.Infos(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(infos))
	assert.Equal(t, capabilities.CapabilityTypeAction, infos[0].CapabilityType)
	assert.Equal(t, capabilities.CapabilityTypeTarget, infos[1].CapabilityType)

	err = scs.Initialise(ctx, "", nil, nil, nil, nil, nil, nil, nil)
	assert.NoError(t, err)
}

func newOutOfProcessStandardCapabilitiesService(t *testing.T, staticChecks bool, stopCh <-chan struct{}) loop.StandardCapabilities {
	scl := loop.StandardCapabilitiesLoop{Logger: logger.Test(t), BrokerConfig: loop.BrokerConfig{Logger: logger.Test(t), StopCh: stopCh}}
	cc := scl.ClientConfig()
	cc.SkipHostEnv = true
	cc.Cmd = NewHelperProcessCommand(loop.PluginStandardCapabilitiesName, staticChecks, 0)
	c := plugin.NewClient(cc)
	t.Cleanup(c.Kill)
	client, err := c.Client()
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close() })
	require.NoError(t, client.Ping())
	i, err := client.Dispense(loop.PluginStandardCapabilitiesName)
	require.NoError(t, err)
	return i.(loop.StandardCapabilities)
}
