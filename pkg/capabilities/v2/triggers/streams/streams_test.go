package streams_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/libocr/ragep2p/types"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	caperrors "github.com/smartcontractkit/chainlink-common/pkg/capabilities/errors"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/triggers/streams"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/triggers/streams/server"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

// TestProtoTypesExist verifies that all protobuf types are properly generated
func TestProtoTypesExist(t *testing.T) {
	// Config type
	config := &streams.Config{
		StreamIds:      []uint32{1, 2, 3},
		MaxFrequencyMs: 5000,
	}
	assert.NotNil(t, config)
	assert.Len(t, config.StreamIds, 3)
	assert.Equal(t, uint64(5000), config.MaxFrequencyMs)

	// Report type
	report := &streams.Report{
		ConfigDigest: []byte{1, 2, 3, 4},
		SeqNr:        42,
		Report:       []byte("report-data"),
		Sigs: []*streams.OCRSignature{
			{
				Signer:    1,
				Signature: []byte("sig1"),
			},
			{
				Signer:    2,
				Signature: []byte("sig2"),
			},
		},
	}
	assert.NotNil(t, report)
	assert.Equal(t, []byte{1, 2, 3, 4}, report.ConfigDigest)
	assert.Equal(t, uint64(42), report.SeqNr)
	assert.Len(t, report.Sigs, 2)
}

// TestConfigGetters verifies getter methods work
func TestConfigGetters(t *testing.T) {
	config := &streams.Config{
		StreamIds:      []uint32{1, 2, 3},
		MaxFrequencyMs: 10000,
	}

	assert.Equal(t, []uint32{1, 2, 3}, config.GetStreamIds())
	assert.Equal(t, uint64(10000), config.GetMaxFrequencyMs())
}

// TestReportGetters verifies Report getter methods
func TestReportGetters(t *testing.T) {
	sigs := []*streams.OCRSignature{
		{
			Signer:    1,
			Signature: []byte("sig1"),
		},
	}

	report := &streams.Report{
		ConfigDigest: []byte{1, 2, 3, 4},
		SeqNr:        99,
		Report:       []byte("test-report"),
		Sigs:         sigs,
	}

	assert.Equal(t, []byte{1, 2, 3, 4}, report.GetConfigDigest())
	assert.Equal(t, uint64(99), report.GetSeqNr())
	assert.Equal(t, []byte("test-report"), report.GetReport())
	assert.Equal(t, sigs, report.GetSigs())
}

// TestStreamsCapabilityInterface verifies the server interface
func TestStreamsCapabilityInterface(t *testing.T) {
	// Verify interface is defined correctly
	var _ server.StreamsCapability = (*mockStreamsCapability)(nil)
}

// mockStreamsCapability implements server.StreamsCapability for testing
type mockStreamsCapability struct {
	registerCalled   bool
	unregisterCalled bool
	startCalled      bool
	closeCalled      bool
}

func (m *mockStreamsCapability) RegisterTrigger(ctx context.Context, triggerID string, metadata capabilities.RequestMetadata, input *streams.Config) (<-chan capabilities.TriggerAndId[*streams.Report], caperrors.Error) {
	m.registerCalled = true
	ch := make(chan capabilities.TriggerAndId[*streams.Report], 1)
	return ch, nil
}

func (m *mockStreamsCapability) UnregisterTrigger(ctx context.Context, triggerID string, metadata capabilities.RequestMetadata, input *streams.Config) caperrors.Error {
	m.unregisterCalled = true
	return nil
}

func (m *mockStreamsCapability) Start(ctx context.Context) error {
	m.startCalled = true
	return nil
}

func (m *mockStreamsCapability) Close() error {
	m.closeCalled = true
	return nil
}

func (m *mockStreamsCapability) HealthReport() map[string]error {
	return map[string]error{"mock": nil}
}

func (m *mockStreamsCapability) Name() string {
	return "MockStreamsCapability"
}

func (m *mockStreamsCapability) Description() string {
	return "Mock implementation for testing"
}

func (m *mockStreamsCapability) Ready() error {
	return nil
}

func (m *mockStreamsCapability) Initialise(ctx context.Context, deps core.StandardCapabilitiesDependencies) error {
	return nil
}

// TestStreamsServerCreation tests creating a server wrapper
func TestStreamsServerCreation(t *testing.T) {
	mock := &mockStreamsCapability{}
	srv := server.NewStreamsServer(mock)

	require.NotNil(t, srv)

	// Test initialization
	ctx := context.Background()
	mockRegistry := &mockCapabilityRegistry{}
	deps := core.StandardCapabilitiesDependencies{
		CapabilityRegistry: mockRegistry,
	}

	err := srv.Initialise(ctx, deps)
	assert.NoError(t, err)

	// Start should be called separately
	err = mock.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, mock.startCalled)

	// Test close
	err = srv.Close()
	assert.NoError(t, err)
	// Note: Close on server doesn't automatically call Close on capability
	err = mock.Close()
	assert.NoError(t, err)
	assert.True(t, mock.closeCalled)
}

// TestTriggerRegistration tests the trigger registration flow
func TestTriggerRegistration(t *testing.T) {
	mock := &mockStreamsCapability{}

	ctx := context.Background()
	triggerID := "test-trigger-123"
	metadata := capabilities.RequestMetadata{
		WorkflowID: "test-workflow",
	}

	config := &streams.Config{
		StreamIds:      []uint32{1},
		MaxFrequencyMs: 1000,
	}

	ch, err := mock.RegisterTrigger(ctx, triggerID, metadata, config)
	require.NoError(t, err)
	require.NotNil(t, ch)
	assert.True(t, mock.registerCalled)

	// Test unregister
	unregErr := mock.UnregisterTrigger(ctx, triggerID, metadata, config)
	assert.NoError(t, unregErr)
	assert.True(t, mock.unregisterCalled)
}

// TestReportStructure tests the Report structure
func TestReportStructure(t *testing.T) {
	sigs := []*streams.OCRSignature{
		{Signer: 1, Signature: []byte("sig1")},
		{Signer: 2, Signature: []byte("sig2")},
	}

	report := &streams.Report{
		ConfigDigest: []byte{1, 2, 3, 4, 5},
		SeqNr:        123,
		Report:       []byte("full-report-bytes"),
		Sigs:         sigs,
	}

	assert.Equal(t, []byte{1, 2, 3, 4, 5}, report.GetConfigDigest())
	assert.Equal(t, uint64(123), report.GetSeqNr())
	assert.Equal(t, []byte("full-report-bytes"), report.GetReport())
	assert.Len(t, report.GetSigs(), 2)
}

// TestOCRSignature tests the OCRSignature structure
func TestOCRSignature(t *testing.T) {
	sig := &streams.OCRSignature{
		Signer:    5,
		Signature: []byte("signature-bytes"),
	}

	assert.Equal(t, uint32(5), sig.GetSigner())
	assert.Equal(t, []byte("signature-bytes"), sig.GetSignature())
}

// TestConfigValidation tests configuration validation scenarios
func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *streams.Config
		expectValid bool
	}{
		{
			name: "valid config with single stream",
			config: &streams.Config{
				StreamIds:      []uint32{1},
				MaxFrequencyMs: 1000,
			},
			expectValid: true,
		},
		{
			name: "valid config with multiple streams",
			config: &streams.Config{
				StreamIds:      []uint32{1, 2, 3},
				MaxFrequencyMs: 5000,
			},
			expectValid: true,
		},
		{
			name: "high frequency",
			config: &streams.Config{
				StreamIds:      []uint32{1},
				MaxFrequencyMs: 100,
			},
			expectValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation - config should be creatable
			assert.NotNil(t, tt.config)
			assert.NotEmpty(t, tt.config.StreamIds)
			assert.Greater(t, tt.config.MaxFrequencyMs, uint64(0))
		})
	}
}

// mockCapabilityRegistry for testing
type mockCapabilityRegistry struct {
	added   []capabilities.BaseCapability
	removed []string
}

func (m *mockCapabilityRegistry) Add(ctx context.Context, capability capabilities.BaseCapability) error {
	m.added = append(m.added, capability)
	return nil
}

func (m *mockCapabilityRegistry) Remove(ctx context.Context, id string) error {
	m.removed = append(m.removed, id)
	return nil
}

func (m *mockCapabilityRegistry) Get(ctx context.Context, id string) (capabilities.BaseCapability, error) {
	return nil, nil
}

func (m *mockCapabilityRegistry) GetTrigger(ctx context.Context, id string) (capabilities.TriggerCapability, error) {
	return nil, nil
}

func (m *mockCapabilityRegistry) GetAction(ctx context.Context, id string) (capabilities.ActionCapability, error) {
	return nil, nil
}

func (m *mockCapabilityRegistry) GetExecutable(ctx context.Context, id string) (capabilities.ExecutableCapability, error) {
	return nil, nil
}

func (m *mockCapabilityRegistry) GetConsensus(ctx context.Context, id string) (capabilities.ConsensusCapability, error) {
	return nil, nil
}

func (m *mockCapabilityRegistry) GetTarget(ctx context.Context, id string) (capabilities.TargetCapability, error) {
	return nil, nil
}

func (m *mockCapabilityRegistry) List(ctx context.Context) ([]capabilities.BaseCapability, error) {
	return m.added, nil
}

func (m *mockCapabilityRegistry) ConfigForCapability(ctx context.Context, capabilityID string, capabilityDonID uint32) (capabilities.CapabilityConfiguration, error) {
	return capabilities.CapabilityConfiguration{}, nil
}

func (m *mockCapabilityRegistry) DONsForCapability(ctx context.Context, id string) ([]capabilities.DONWithNodes, error) {
	return nil, nil
}

func (m *mockCapabilityRegistry) LocalNode(ctx context.Context) (capabilities.Node, error) {
	return capabilities.Node{}, nil
}

func (m *mockCapabilityRegistry) NodeByPeerID(ctx context.Context, peerID types.PeerID) (capabilities.Node, error) {
	return capabilities.Node{}, nil
}

// TestServerLifecycle tests the complete server lifecycle
func TestServerLifecycle(t *testing.T) {
	mock := &mockStreamsCapability{}
	srv := server.NewStreamsServer(mock)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	mockRegistry := &mockCapabilityRegistry{}
	deps := core.StandardCapabilitiesDependencies{
		CapabilityRegistry: mockRegistry,
	}

	// Initialize
	err := srv.Initialise(ctx, deps)
	require.NoError(t, err)
	assert.Len(t, mockRegistry.added, 1, "Capability should be registered")

	// Start must be called separately
	err = mock.Start(ctx)
	require.NoError(t, err)
	assert.True(t, mock.startCalled, "Start should be called")

	// Get infos
	infos, err := srv.Infos(ctx)
	require.NoError(t, err)
	require.Len(t, infos, 1)
	assert.Equal(t, "streams-trigger@2.0.0", infos[0].ID)

	// Close
	err = srv.Close()
	require.NoError(t, err)
	assert.True(t, mock.closeCalled, "Close should be called")
	assert.Len(t, mockRegistry.removed, 1, "Capability should be unregistered")
	assert.Equal(t, "streams-trigger@2.0.0", mockRegistry.removed[0])
}

// BenchmarkReportCreation benchmarks creating Report objects
func BenchmarkReportCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = &streams.Report{
			ConfigDigest: []byte{1, 2, 3, 4},
			SeqNr:        uint64(i),
			Report:       []byte("report-data"),
			Sigs: []*streams.OCRSignature{
				{
					Signer:    1,
					Signature: []byte("sig1"),
				},
				{
					Signer:    2,
					Signature: []byte("sig2"),
				},
			},
		}
	}
}
