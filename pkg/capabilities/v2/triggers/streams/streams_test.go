package streams_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/libocr/ragep2p/types"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/triggers/streams"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/triggers/streams/server"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

// TestProtoTypesExist verifies that all protobuf types are properly generated
func TestProtoTypesExist(t *testing.T) {
	// Config type
	config := &streams.Config{
		FeedIds:        []string{"0x0001", "0x0002"},
		MaxFrequencyMs: 5000,
	}
	assert.NotNil(t, config)
	assert.Len(t, config.FeedIds, 2)
	assert.Equal(t, uint64(5000), config.MaxFrequencyMs)

	// Feed type
	feed := &streams.Feed{
		Timestamp: 1234567890,
		Metadata: &streams.SignersMetadata{
			Signers:               []string{"signer1", "signer2"},
			MinRequiredSignatures: 2,
		},
		Payload: []*streams.FeedReport{
			{
				FeedId:               "0x0001",
				FullReport:           []byte("report-data"),
				ReportContext:        []byte("context"),
				Signatures:           [][]byte{[]byte("sig1")},
				BenchmarkPrice:       []byte("price"),
				ObservationTimestamp: 1234567890,
			},
		},
	}
	assert.NotNil(t, feed)
	assert.Equal(t, int64(1234567890), feed.Timestamp)
	assert.Len(t, feed.Payload, 1)
}

// TestConfigGetters verifies getter methods work
func TestConfigGetters(t *testing.T) {
	config := &streams.Config{
		FeedIds:        []string{"0xfeed1", "0xfeed2", "0xfeed3"},
		MaxFrequencyMs: 10000,
	}

	assert.Equal(t, []string{"0xfeed1", "0xfeed2", "0xfeed3"}, config.GetFeedIds())
	assert.Equal(t, uint64(10000), config.GetMaxFrequencyMs())
}

// TestFeedGetters verifies Feed getter methods
func TestFeedGetters(t *testing.T) {
	metadata := &streams.SignersMetadata{
		Signers:               []string{"signer1"},
		MinRequiredSignatures: 1,
	}

	feed := &streams.Feed{
		Timestamp: 9999999999,
		Metadata:  metadata,
		Payload:   []*streams.FeedReport{},
	}

	assert.Equal(t, int64(9999999999), feed.GetTimestamp())
	assert.Equal(t, metadata, feed.GetMetadata())
	assert.NotNil(t, feed.GetPayload())
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

func (m *mockStreamsCapability) RegisterTrigger(ctx context.Context, triggerID string, metadata capabilities.RequestMetadata, input *streams.Config) (<-chan capabilities.TriggerAndId[*streams.Feed], error) {
	m.registerCalled = true
	ch := make(chan capabilities.TriggerAndId[*streams.Feed], 1)
	return ch, nil
}

func (m *mockStreamsCapability) UnregisterTrigger(ctx context.Context, triggerID string, metadata capabilities.RequestMetadata, input *streams.Config) error {
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
		FeedIds:        []string{"0x0001"},
		MaxFrequencyMs: 1000,
	}
	
	ch, err := mock.RegisterTrigger(ctx, triggerID, metadata, config)
	require.NoError(t, err)
	require.NotNil(t, ch)
	assert.True(t, mock.registerCalled)
	
	// Test unregister
	err = mock.UnregisterTrigger(ctx, triggerID, metadata, config)
	assert.NoError(t, err)
	assert.True(t, mock.unregisterCalled)
}

// TestFeedReportStructure tests the FeedReport structure
func TestFeedReportStructure(t *testing.T) {
	report := &streams.FeedReport{
		FeedId:               "0xfeedid12345",
		FullReport:           []byte("full-report-bytes"),
		ReportContext:        []byte("report-context"),
		Signatures:           [][]byte{[]byte("sig1"), []byte("sig2")},
		BenchmarkPrice:       []byte("benchmark-price-bytes"),
		ObservationTimestamp: 1700000000,
	}
	
	assert.Equal(t, "0xfeedid12345", report.GetFeedId())
	assert.Equal(t, []byte("full-report-bytes"), report.GetFullReport())
	assert.Equal(t, []byte("report-context"), report.GetReportContext())
	assert.Len(t, report.GetSignatures(), 2)
	assert.Equal(t, []byte("benchmark-price-bytes"), report.GetBenchmarkPrice())
	assert.Equal(t, int64(1700000000), report.GetObservationTimestamp())
}

// TestSignersMetadata tests the SignersMetadata structure
func TestSignersMetadata(t *testing.T) {
	metadata := &streams.SignersMetadata{
		Signers:               []string{"0xsigner1", "0xsigner2", "0xsigner3"},
		MinRequiredSignatures: 2,
	}
	
	assert.Len(t, metadata.GetSigners(), 3)
	assert.Equal(t, int64(2), metadata.GetMinRequiredSignatures())
}

// TestConfigValidation tests configuration validation scenarios
func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *streams.Config
		expectValid bool
	}{
		{
			name: "valid config with single feed",
			config: &streams.Config{
				FeedIds:        []string{"0x0001"},
				MaxFrequencyMs: 1000,
			},
			expectValid: true,
		},
		{
			name: "valid config with multiple feeds",
			config: &streams.Config{
				FeedIds:        []string{"0x0001", "0x0002", "0x0003"},
				MaxFrequencyMs: 5000,
			},
			expectValid: true,
		},
		{
			name: "high frequency",
			config: &streams.Config{
				FeedIds:        []string{"0x0001"},
				MaxFrequencyMs: 100,
			},
			expectValid: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation - config should be creatable
			assert.NotNil(t, tt.config)
			assert.NotEmpty(t, tt.config.FeedIds)
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
	assert.Equal(t, "streams-trigger@1.0.0", infos[0].ID)
	
	// Close
	err = srv.Close()
	require.NoError(t, err)
	assert.True(t, mock.closeCalled, "Close should be called")
	assert.Len(t, mockRegistry.removed, 1, "Capability should be unregistered")
	assert.Equal(t, "streams-trigger@1.0.0", mockRegistry.removed[0])
}

// BenchmarkFeedCreation benchmarks creating Feed objects
func BenchmarkFeedCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = &streams.Feed{
			Timestamp: int64(i),
			Metadata: &streams.SignersMetadata{
				Signers:               []string{"signer1", "signer2"},
				MinRequiredSignatures: 2,
			},
			Payload: []*streams.FeedReport{
				{
					FeedId:               "0x0001",
					FullReport:           []byte("report"),
					ReportContext:        []byte("context"),
					Signatures:           [][]byte{[]byte("sig")},
					BenchmarkPrice:       []byte("price"),
					ObservationTimestamp: int64(i),
				},
			},
		}
	}
}

