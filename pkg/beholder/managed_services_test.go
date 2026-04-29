package beholder_test

import (
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
	assert.Nil(t, c.ManagedServices())
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

	assert.Nil(t, client.ManagedServices())
}

func TestManagedServices_BatchEnabledReturnsClientService(t *testing.T) {
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
	defer func() { _ = client.Close() }()
	managedServices := client.ManagedServices()
	require.Len(t, managedServices, 1)
	assert.Same(t, client, managedServices[0])
}

func TestClientLifecycle_BatchEmitterStartsWithClient(t *testing.T) {
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

	// Before Start, client and emitter should report unready.
	assert.Error(t, client.Ready())
	err = client.Emitter.Emit(t.Context(), []byte("body"),
		beholder.AttrKeyDomain, "platform",
		beholder.AttrKeyEntity, "TestEvent",
	)
	assert.Error(t, err)

	require.NoError(t, client.Start(t.Context()))
	require.NoError(t, client.Ready())
	_ = client.Close()
}
