package beholder_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
)

func TestEmitOnlyAdapter_CloseIsNoOp(t *testing.T) {
	// Create a client with batch enabled, get the emitter, and verify
	// that closing the client (which closes the DualSourceEmitter which
	// closes the emitOnlyAdapter) does not close the batch emitter service.
	lggr := newTestLogger(t)

	client, err := beholder.NewClient(beholder.Config{
		OtelExporterGRPCEndpoint:       "localhost:4317",
		ChipIngressEmitterEnabled:      true,
		ChipIngressEmitterGRPCEndpoint: "localhost:9090",
		ChipIngressInsecureConnection:  true,
		ChipIngressBatchEmitterEnabled: true,
		ChipIngressLogger:              lggr,
		ChipIngressBufferSize:          10,
		ChipIngressMaxBatchSize:        5,
		ChipIngressSendInterval:        50 * time.Millisecond,
		ChipIngressSendTimeout:         1 * time.Second,
		ChipIngressDrainTimeout:        1 * time.Second,
	})
	require.NoError(t, err)

	managed := client.ManagedServices()
	require.Len(t, managed, 1)
	svc := managed[0]

	// Close the client (which closes the DualSourceEmitter → emitOnlyAdapter.Close → no-op).
	// OTel provider flush may fail (no real endpoint) — we only care that it doesn't
	// close the batch emitter service.
	_ = client.Close()

	// The batch emitter service should still be startable (not already closed)
	err = svc.Start(t.Context())
	require.NoError(t, err)

	// And closable
	err = svc.Close()
	require.NoError(t, err)

	// Second close should error (StopOnce)
	err = svc.Close()
	assert.Error(t, err, "second close should fail")
}
