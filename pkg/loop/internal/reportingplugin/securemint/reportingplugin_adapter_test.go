package securemint

import (
	"context"
	"encoding/binary"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	sm "github.com/smartcontractkit/chainlink-common/pkg/types/core/securemint"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

func TestReportingPluginBytesToChainSelectorAdapter_Reports(t *testing.T) {
	tests := []struct {
		name          string
		mockReports   []ocr3types.ReportPlus[[]byte]
		mockError     error
		expectedError bool
		expectedCount int
	}{
		{
			name: "successful conversion with single report",
			mockReports: []ocr3types.ReportPlus[[]byte]{
				{
					ReportWithInfo: ocr3types.ReportWithInfo[[]byte]{
						Report: []byte("test report"),
						Info:   []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, // uint64(1) in little endian
					},
				},
			},
			mockError:     nil,
			expectedError: false,
			expectedCount: 1,
		},
		{
			name: "successful conversion with multiple reports",
			mockReports: []ocr3types.ReportPlus[[]byte]{
				{
					ReportWithInfo: ocr3types.ReportWithInfo[[]byte]{
						Report: []byte("test report 1"),
						Info:   []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, // uint64(1)
					},
				},
				{
					ReportWithInfo: ocr3types.ReportWithInfo[[]byte]{
						Report: []byte("test report 2"),
						Info:   []byte{0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, // uint64(2)
					},
				},
			},
			mockError:     nil,
			expectedError: false,
			expectedCount: 2,
		},
		{
			name:          "empty reports",
			mockReports:   []ocr3types.ReportPlus[[]byte]{},
			mockError:     nil,
			expectedError: false,
			expectedCount: 0,
		},
		{
			name: "short info bytes (less than 8 bytes)",
			mockReports: []ocr3types.ReportPlus[[]byte]{
				{
					ReportWithInfo: ocr3types.ReportWithInfo[[]byte]{
						Report: []byte("test report"),
						Info:   []byte{0x01, 0x02}, // Only 2 bytes
					},
				},
			},
			mockError:     errors.New("info is less than 8 bytes"),
			expectedError: true,
			expectedCount: 1,
		},
		{
			name:          "underlying plugin error",
			mockReports:   nil,
			mockError:     errors.New("plugin error"),
			expectedError: true,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPlugin := new(MockReportingPluginBytes)
			adapter := &reportingPluginBytesToChainSelectorAdapter{plugin: mockPlugin}

			mockPlugin.On("Reports", mock.Anything, mock.Anything, mock.Anything).Return(tt.mockReports, tt.mockError)

			reports, err := adapter.Reports(context.Background(), 1, ocr3types.Outcome([]byte("test outcome")))

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, reports)
			} else {
				assert.NoError(t, err)
				assert.Len(t, reports, tt.expectedCount)

				// Verify conversion for each report
				for i, report := range reports {
					assert.Equal(t, tt.mockReports[i].ReportWithInfo.Report, report.ReportWithInfo.Report)
					expectedChainSelector := sm.ChainSelector(binary.LittleEndian.Uint64(tt.mockReports[i].ReportWithInfo.Info[:8]))
					assert.Equal(t, expectedChainSelector, report.ReportWithInfo.Info)
				}
			}

			mockPlugin.AssertExpectations(t)
		})
	}
}

func TestReportingPluginBytesToChainSelectorAdapter_ShouldAcceptAttestedReport(t *testing.T) {
	mockPlugin := new(MockReportingPluginBytes)
	adapter := &reportingPluginBytesToChainSelectorAdapter{plugin: mockPlugin}

	chainSelector := sm.ChainSelector(123)
	report := ocr3types.ReportWithInfo[sm.ChainSelector]{
		Report: []byte("test report"),
		Info:   chainSelector,
	}

	// Convert chain selector to bytes for mock expectation
	expectedBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(expectedBytes, uint64(chainSelector))
	expectedReport := ocr3types.ReportWithInfo[[]byte]{
		Report: []byte("test report"),
		Info:   expectedBytes,
	}

	mockPlugin.On("ShouldAcceptAttestedReport", mock.Anything, uint64(1), expectedReport).Return(true, nil)

	accepted, err := adapter.ShouldAcceptAttestedReport(context.Background(), 1, report)

	assert.NoError(t, err)
	assert.True(t, accepted)
	mockPlugin.AssertExpectations(t)
}

func TestReportingPluginBytesToChainSelectorAdapter_ShouldTransmitAcceptedReport(t *testing.T) {
	mockPlugin := new(MockReportingPluginBytes)
	adapter := &reportingPluginBytesToChainSelectorAdapter{plugin: mockPlugin}

	chainSelector := sm.ChainSelector(456)
	report := ocr3types.ReportWithInfo[sm.ChainSelector]{
		Report: []byte("test report"),
		Info:   chainSelector,
	}

	// Convert chain selector to bytes for mock expectation
	expectedBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(expectedBytes, uint64(chainSelector))
	expectedReport := ocr3types.ReportWithInfo[[]byte]{
		Report: []byte("test report"),
		Info:   expectedBytes,
	}

	mockPlugin.On("ShouldTransmitAcceptedReport", mock.Anything, uint64(2), expectedReport).Return(false, nil)

	transmit, err := adapter.ShouldTransmitAcceptedReport(context.Background(), 2, report)

	assert.NoError(t, err)
	assert.False(t, transmit)
	mockPlugin.AssertExpectations(t)
}

func TestReportingPluginChainSelectorToBytesAdapter_Reports(t *testing.T) {
	tests := []struct {
		name          string
		mockReports   []ocr3types.ReportPlus[sm.ChainSelector]
		mockError     error
		expectedError bool
		expectedCount int
	}{
		{
			name: "successful conversion with single report",
			mockReports: []ocr3types.ReportPlus[sm.ChainSelector]{
				{
					ReportWithInfo: ocr3types.ReportWithInfo[sm.ChainSelector]{
						Report: []byte("test report"),
						Info:   sm.ChainSelector(1),
					},
				},
			},
			mockError:     nil,
			expectedError: false,
			expectedCount: 1,
		},
		{
			name: "successful conversion with multiple reports",
			mockReports: []ocr3types.ReportPlus[sm.ChainSelector]{
				{
					ReportWithInfo: ocr3types.ReportWithInfo[sm.ChainSelector]{
						Report: []byte("test report 1"),
						Info:   sm.ChainSelector(1),
					},
				},
				{
					ReportWithInfo: ocr3types.ReportWithInfo[sm.ChainSelector]{
						Report: []byte("test report 2"),
						Info:   sm.ChainSelector(2),
					},
				},
			},
			mockError:     nil,
			expectedError: false,
			expectedCount: 2,
		},
		{
			name:          "empty reports",
			mockReports:   []ocr3types.ReportPlus[sm.ChainSelector]{},
			mockError:     nil,
			expectedError: false,
			expectedCount: 0,
		},
		{
			name:          "underlying plugin error",
			mockReports:   nil,
			mockError:     errors.New("plugin error"),
			expectedError: true,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPlugin := new(MockReportingPluginChainSelector)
			adapter := &reportingPluginChainSelectorToBytesAdapter{plugin: mockPlugin}

			mockPlugin.On("Reports", mock.Anything, mock.Anything, mock.Anything).Return(tt.mockReports, tt.mockError)

			reports, err := adapter.Reports(context.Background(), 1, ocr3types.Outcome([]byte("test outcome")))

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, reports)
			} else {
				assert.NoError(t, err)
				assert.Len(t, reports, tt.expectedCount)

				// Verify conversion for each report
				for i, report := range reports {
					assert.Equal(t, tt.mockReports[i].ReportWithInfo.Report, report.ReportWithInfo.Report)

					// Check bytes conversion
					expectedBytes := make([]byte, 8)
					binary.LittleEndian.PutUint64(expectedBytes, uint64(tt.mockReports[i].ReportWithInfo.Info))
					assert.Equal(t, expectedBytes, report.ReportWithInfo.Info)
				}
			}

			mockPlugin.AssertExpectations(t)
		})
	}
}

func TestReportingPluginChainSelectorToBytesAdapter_ShouldAcceptAttestedReport(t *testing.T) {
	mockPlugin := new(MockReportingPluginChainSelector)
	adapter := &reportingPluginChainSelectorToBytesAdapter{plugin: mockPlugin}

	chainSelectorBytes := []byte{0x7B, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00} // uint64(123) in little endian
	report := ocr3types.ReportWithInfo[[]byte]{
		Report: []byte("test report"),
		Info:   chainSelectorBytes,
	}

	// Convert bytes to chain selector for mock expectation
	expectedChainSelector := sm.ChainSelector(binary.LittleEndian.Uint64(chainSelectorBytes))
	expectedReport := ocr3types.ReportWithInfo[sm.ChainSelector]{
		Report: []byte("test report"),
		Info:   expectedChainSelector,
	}

	mockPlugin.On("ShouldAcceptAttestedReport", mock.Anything, uint64(1), expectedReport).Return(true, nil)

	accepted, err := adapter.ShouldAcceptAttestedReport(context.Background(), 1, report)

	assert.NoError(t, err)
	assert.True(t, accepted)
	mockPlugin.AssertExpectations(t)
}

func TestReportingPluginChainSelectorToBytesAdapter_ShouldTransmitAcceptedReport(t *testing.T) {
	mockPlugin := new(MockReportingPluginChainSelector)
	adapter := &reportingPluginChainSelectorToBytesAdapter{plugin: mockPlugin}

	chainSelectorBytes := []byte{0xC8, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00} // uint64(456) in little endian
	report := ocr3types.ReportWithInfo[[]byte]{
		Report: []byte("test report"),
		Info:   chainSelectorBytes,
	}

	// Convert bytes to chain selector for mock expectation
	expectedChainSelector := sm.ChainSelector(binary.LittleEndian.Uint64(chainSelectorBytes))
	expectedReport := ocr3types.ReportWithInfo[sm.ChainSelector]{
		Report: []byte("test report"),
		Info:   expectedChainSelector,
	}

	mockPlugin.On("ShouldTransmitAcceptedReport", mock.Anything, uint64(2), expectedReport).Return(false, nil)

	transmit, err := adapter.ShouldTransmitAcceptedReport(context.Background(), 2, report)

	assert.NoError(t, err)
	assert.False(t, transmit)
	mockPlugin.AssertExpectations(t)
}

// MockReportingPluginBytes is a mock implementation of ocr3types.ReportingPlugin[[]byte]
type MockReportingPluginBytes struct {
	mock.Mock
}

func (m *MockReportingPluginBytes) Query(ctx context.Context, outctx ocr3types.OutcomeContext) (types.Query, error) {
	args := m.Called(ctx, outctx)
	return args.Get(0).(types.Query), args.Error(1)
}

func (m *MockReportingPluginBytes) Observation(ctx context.Context, outctx ocr3types.OutcomeContext, query types.Query) (types.Observation, error) {
	args := m.Called(ctx, outctx, query)
	return args.Get(0).(types.Observation), args.Error(1)
}

func (m *MockReportingPluginBytes) ValidateObservation(ctx context.Context, outctx ocr3types.OutcomeContext, query types.Query, ao types.AttributedObservation) error {
	args := m.Called(ctx, outctx, query, ao)
	return args.Error(0)
}

func (m *MockReportingPluginBytes) ObservationQuorum(ctx context.Context, outctx ocr3types.OutcomeContext, query types.Query, aos []types.AttributedObservation) (bool, error) {
	args := m.Called(ctx, outctx, query, aos)
	return args.Bool(0), args.Error(1)
}

func (m *MockReportingPluginBytes) Outcome(ctx context.Context, outctx ocr3types.OutcomeContext, query types.Query, aos []types.AttributedObservation) (ocr3types.Outcome, error) {
	args := m.Called(ctx, outctx, query, aos)
	return args.Get(0).(ocr3types.Outcome), args.Error(1)
}

func (m *MockReportingPluginBytes) Reports(ctx context.Context, seqNr uint64, outcome ocr3types.Outcome) ([]ocr3types.ReportPlus[[]byte], error) {
	args := m.Called(ctx, seqNr, outcome)
	return args.Get(0).([]ocr3types.ReportPlus[[]byte]), args.Error(1)
}

func (m *MockReportingPluginBytes) ShouldAcceptAttestedReport(ctx context.Context, seqNr uint64, report ocr3types.ReportWithInfo[[]byte]) (bool, error) {
	args := m.Called(ctx, seqNr, report)
	return args.Bool(0), args.Error(1)
}

func (m *MockReportingPluginBytes) ShouldTransmitAcceptedReport(ctx context.Context, seqNr uint64, report ocr3types.ReportWithInfo[[]byte]) (bool, error) {
	args := m.Called(ctx, seqNr, report)
	return args.Bool(0), args.Error(1)
}

func (m *MockReportingPluginBytes) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockReportingPluginChainSelector is a mock implementation of ocr3types.ReportingPlugin[securemint.ChainSelector]
type MockReportingPluginChainSelector struct {
	mock.Mock
}

func (m *MockReportingPluginChainSelector) Query(ctx context.Context, outctx ocr3types.OutcomeContext) (types.Query, error) {
	args := m.Called(ctx, outctx)
	return args.Get(0).(types.Query), args.Error(1)
}

func (m *MockReportingPluginChainSelector) Observation(ctx context.Context, outctx ocr3types.OutcomeContext, query types.Query) (types.Observation, error) {
	args := m.Called(ctx, outctx, query)
	return args.Get(0).(types.Observation), args.Error(1)
}

func (m *MockReportingPluginChainSelector) ValidateObservation(ctx context.Context, outctx ocr3types.OutcomeContext, query types.Query, ao types.AttributedObservation) error {
	args := m.Called(ctx, outctx, query, ao)
	return args.Error(0)
}

func (m *MockReportingPluginChainSelector) ObservationQuorum(ctx context.Context, outctx ocr3types.OutcomeContext, query types.Query, aos []types.AttributedObservation) (bool, error) {
	args := m.Called(ctx, outctx, query, aos)
	return args.Bool(0), args.Error(1)
}

func (m *MockReportingPluginChainSelector) Outcome(ctx context.Context, outctx ocr3types.OutcomeContext, query types.Query, aos []types.AttributedObservation) (ocr3types.Outcome, error) {
	args := m.Called(ctx, outctx, query, aos)
	return args.Get(0).(ocr3types.Outcome), args.Error(1)
}

func (m *MockReportingPluginChainSelector) Reports(ctx context.Context, seqNr uint64, outcome ocr3types.Outcome) ([]ocr3types.ReportPlus[sm.ChainSelector], error) {
	args := m.Called(ctx, seqNr, outcome)
	return args.Get(0).([]ocr3types.ReportPlus[sm.ChainSelector]), args.Error(1)
}

func (m *MockReportingPluginChainSelector) ShouldAcceptAttestedReport(ctx context.Context, seqNr uint64, report ocr3types.ReportWithInfo[sm.ChainSelector]) (bool, error) {
	args := m.Called(ctx, seqNr, report)
	return args.Bool(0), args.Error(1)
}

func (m *MockReportingPluginChainSelector) ShouldTransmitAcceptedReport(ctx context.Context, seqNr uint64, report ocr3types.ReportWithInfo[sm.ChainSelector]) (bool, error) {
	args := m.Called(ctx, seqNr, report)
	return args.Bool(0), args.Error(1)
}

func (m *MockReportingPluginChainSelector) Close() error {
	args := m.Called()
	return args.Error(0)
}
