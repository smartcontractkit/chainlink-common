package loop_test

import (
	"os/exec"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop"
	keystoretest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/keystore/test"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/goplugin"
	ccipocr3client "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/ext/ccipocr3"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test"
	"github.com/smartcontractkit/chainlink-common/pkg/services/servicetest"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/ccipocr3"
)

// TestCCIPSyncPersistence tests the persistence of sync requests across relayer restarts. This test is testing
// logic from chainlink-common/pkg/loop/internal/relayer/pluginprovider/ext/ccipocr3/chainaccessor.go but we need
// the full CCIPProvider and Relayer to properly test the persistence across restarts.
func TestCCIPChainAccessorSyncPersistence(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	// Observed logger for confirming PIDs
	lggr, logs := logger.TestObserved(t, zapcore.DebugLevel)

	// Relayer service (client side)
	relayerService := loop.NewRelayerService(
		lggr,
		loop.GRPCOpts{},
		func() *exec.Cmd {
			return NewHelperProcessCommand(loop.PluginRelayerName, false, 0)
		},
		test.ConfigTOML,
		keystoretest.Keystore,
		keystoretest.Keystore,
		nil,
	)

	// Kill hook is defined on the relayer client (service) because the client spawns the server child process
	hook := relayerService.XXXTestHook()
	servicetest.Run(t, relayerService)

	// Create CCIPProvider client and issue first Sync() call. This client should persist and reattach
	// to the new server after the kill hook is run.
	ccipProvider, err := relayerService.NewCCIPProvider(ctx, types.CCIPProviderArgs{
		ExternalJobID:        uuid.New(),
		ContractReaderConfig: []byte("asdf"),
		ChainWriterConfig:    []byte("asdf"),
		OffRampAddress:       []byte("0x1234123412341234123412341234123412341234"),
		PluginType:           0,
		TransmitterAddress:   "0x4321432143214321432143214321432143214321",
	})
	require.NoError(t, err)
	require.NotNil(t, ccipProvider)

	firstContractNameToSync := "OnRamp"
	firstContractAddressToSync := ccipocr3.UnknownAddress("0x123412341234")

	// Perform first Sync() call
	err = ccipProvider.ChainAccessor().Sync(ctx, firstContractNameToSync, firstContractAddressToSync)
	require.NoError(t, err)

	// Confirm first sync call was stored in the c.syncs map
	ccipProviderClient, ok := ccipProvider.(*ccipocr3client.CCIPProviderClient)
	require.True(t, ok)
	firstSyncs := ccipProviderClient.GetSyncRequests()
	require.Len(t, firstSyncs, 1, "Should have one sync request in ChainAccessorClient c.syncs")

	// Capture initial server side process ID before kill
	initialPID := extractLatestPluginPID(logs)
	require.NotZero(t, initialPID)

	// Kill the server process (RelayerService should auto-restart it)
	hook.Kill()

	// Give some time for the keep alive to kick in
	time.Sleep(2 * goplugin.KeepAliveTickDuration)

	// Capture process ID again after restart and verify it's different
	restartedPID := extractLatestPluginPID(logs)
	require.NotZero(t, restartedPID)
	assert.NotEqual(t, initialPID, restartedPID, "Server should have restarted with different process ID")

	// Verify new Sync() call still works and now the client map should have two
	secondContractNameToSync := "OffRamp"
	newContractAddress := ccipocr3.UnknownAddress("0x567856785678")
	err = ccipProvider.ChainAccessor().Sync(ctx, secondContractNameToSync, newContractAddress)
	require.NoError(t, err)
	finalSyncs := ccipProviderClient.GetSyncRequests()
	require.Len(t, finalSyncs, 2, "Should have both first and second sync requests in client memory")

	// Verify first sync entry persisted through restart
	assert.Contains(t, finalSyncs, firstContractNameToSync)
	assert.Equal(t, []byte(firstContractAddressToSync), finalSyncs[firstContractNameToSync])

	// Verify second sync entry was added
	assert.Contains(t, finalSyncs, secondContractNameToSync)
	assert.Equal(t, []byte(newContractAddress), finalSyncs[secondContractNameToSync])
}

// extractLatestPluginPID extracts the most recent plugin process ID from the logs using the `plugin started` log
func extractLatestPluginPID(logs *observer.ObservedLogs) int {
	var latestPID int
	for _, entry := range logs.All() {
		if entry.Message == "plugin started" {
			for _, field := range entry.Context {
				if field.Key == "pid" {
					latestPID = int(field.Integer)
				}
			}
		}
	}

	return latestPID
}
