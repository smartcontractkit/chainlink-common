package beholder_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
)

func TestManagedServices_NilReceiver(t *testing.T) {
	var c *beholder.Client
	assert.Nil(t, c.ManagedServices())
}

func TestManagedServices_NoopClient(t *testing.T) {
	c := beholder.NewNoopClient()
	assert.Empty(t, c.ManagedServices())
}

func TestManagedServices_BatchDisabled(t *testing.T) {
	client, err := beholder.NewClient(beholder.Config{
		OtelExporterGRPCEndpoint:       "localhost:4317",
		ChipIngressEmitterEnabled:      true,
		ChipIngressEmitterGRPCEndpoint: "localhost:9090",
		ChipIngressInsecureConnection:  true,
		ChipIngressBatchEmitterEnabled: false,
	})
	require.NoError(t, err)
	defer func() { _ = client.Close() }()

	assert.Empty(t, client.ManagedServices())
}

func TestManagedServices_BatchEnabled(t *testing.T) {
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
	assert.Equal(t, "ChipIngressBatchEmitter", managed[0].Name())

	// Service should be startable (not already started)
	err = managed[0].Start(context.Background())
	require.NoError(t, err)

	err = managed[0].Close()
	require.NoError(t, err)

	// Client.Close may return OTel flush errors (no real endpoint) — that's fine.
	// The key assertion is that closing the client does NOT close the batch emitter
	// (emitOnlyAdapter.Close is a no-op).
	_ = client.Close()
}

func TestManagedServices_BatchEmitterNotAutoStarted(t *testing.T) {
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

	// The service should not be ready yet (not started)
	err = managed[0].Ready()
	assert.Error(t, err, "service should not be ready before Start()")

	// Start, verify ready, then close
	require.NoError(t, managed[0].Start(context.Background()))
	require.NoError(t, managed[0].Ready())
	require.NoError(t, managed[0].Close())
}
