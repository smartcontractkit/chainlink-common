package securemint

import (
	"context"
	"errors"
	"math/big"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	pb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/securemint"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockExternalAdapterClient is a mock implementation of pb.ExternalAdapterClient
type mockExternalAdapterClient struct {
	mock.Mock
}

func (m *mockExternalAdapterClient) GetPayload(ctx context.Context, in *pb.Blocks, opts ...grpc.CallOption) (*pb.ExternalAdapterPayload, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.ExternalAdapterPayload), args.Error(1)
}

// mockExternalAdapter is a mock implementation of core.ExternalAdapter
type mockExternalAdapter struct {
	mock.Mock
}

func (m *mockExternalAdapter) GetPayload(ctx context.Context, blocks core.Blocks) (core.ExternalAdapterPayload, error) {
	args := m.Called(ctx, blocks)
	return args.Get(0).(core.ExternalAdapterPayload), args.Error(1)
}

func TestExternalAdapterClient_GetPayload(t *testing.T) {
	tests := []struct {
		name           string
		inputBlocks    core.Blocks
		mockResponse   *pb.ExternalAdapterPayload
		mockError      error
		expectedResult core.ExternalAdapterPayload
		expectedError  bool
	}{
		{
			name: "successful request with single chain",
			inputBlocks: core.Blocks{
				core.ChainSelector(1): core.BlockNumber(100),
			},
			mockResponse: &pb.ExternalAdapterPayload{
				Mintables: map[uint64]*pb.BlockMintablePair{
					1: {
						BlockNumber: 100,
						Mintable:    "1000",
					},
				},
				ReserveInfo: &pb.ReserveInfo{
					ReserveAmount: "5000",
					Timestamp:     timestamppb.New(time.Unix(1640995200, 0).UTC()), // 2022-01-01 00:00:00 UTC
				},
				LatestBlocks: &pb.Blocks{
					Value: map[uint64]uint64{
						1: 100,
					},
				},
			},
			expectedResult: core.ExternalAdapterPayload{
				Mintables: map[core.ChainSelector]core.BlockMintablePair{
					core.ChainSelector(1): {
						Block:    core.BlockNumber(100),
						Mintable: big.NewInt(1000),
					},
				},
				ReserveInfo: core.ReserveInfo{
					ReserveAmount: big.NewInt(5000),
					Timestamp:     time.Unix(1640995200, 0).UTC(),
				},
				LatestBlocks: core.Blocks{
					core.ChainSelector(1): core.BlockNumber(100),
				},
			},
			expectedError: false,
		},
		{
			name: "successful request with multiple chains",
			inputBlocks: core.Blocks{
				core.ChainSelector(1): core.BlockNumber(100),
				core.ChainSelector(2): core.BlockNumber(200),
			},
			mockResponse: &pb.ExternalAdapterPayload{
				Mintables: map[uint64]*pb.BlockMintablePair{
					1: {
						BlockNumber: 100,
						Mintable:    "1000",
					},
					2: {
						BlockNumber: 200,
						Mintable:    "2000",
					},
				},
				ReserveInfo: &pb.ReserveInfo{
					ReserveAmount: "5000",
					Timestamp:     timestamppb.New(time.Unix(1640995200, 0).UTC()),
				},
				LatestBlocks: &pb.Blocks{
					Value: map[uint64]uint64{
						1: 100,
						2: 200,
					},
				},
			},
			expectedResult: core.ExternalAdapterPayload{
				Mintables: map[core.ChainSelector]core.BlockMintablePair{
					core.ChainSelector(1): {
						Block:    core.BlockNumber(100),
						Mintable: big.NewInt(1000),
					},
					core.ChainSelector(2): {
						Block:    core.BlockNumber(200),
						Mintable: big.NewInt(2000),
					},
				},
				ReserveInfo: core.ReserveInfo{
					ReserveAmount: big.NewInt(5000),
					Timestamp:     time.Unix(1640995200, 0).UTC(),
				},
				LatestBlocks: core.Blocks{
					core.ChainSelector(1): core.BlockNumber(100),
					core.ChainSelector(2): core.BlockNumber(200),
				},
			},
			expectedError: false,
		},
		{
			name:        "empty input blocks",
			inputBlocks: core.Blocks{},
			mockResponse: &pb.ExternalAdapterPayload{
				Mintables: map[uint64]*pb.BlockMintablePair{},
				ReserveInfo: &pb.ReserveInfo{
					ReserveAmount: "0",
					Timestamp:     timestamppb.New(time.Unix(1640995200, 0).UTC()),
				},
				LatestBlocks: &pb.Blocks{
					Value: map[uint64]uint64{},
				},
			},
			expectedResult: core.ExternalAdapterPayload{
				Mintables: map[core.ChainSelector]core.BlockMintablePair{},
				ReserveInfo: core.ReserveInfo{
					ReserveAmount: big.NewInt(0),
					Timestamp:     time.Unix(1640995200, 0).UTC(),
				},
				LatestBlocks: core.Blocks{},
			},
			expectedError: false,
		},
		{
			name: "grpc error",
			inputBlocks: core.Blocks{
				core.ChainSelector(1): core.BlockNumber(100),
			},
			mockResponse:  nil,
			mockError:     errors.New("grpc error"),
			expectedError: true,
		},
		{
			name: "invalid mintable string",
			inputBlocks: core.Blocks{
				core.ChainSelector(1): core.BlockNumber(100),
			},
			mockResponse: &pb.ExternalAdapterPayload{
				Mintables: map[uint64]*pb.BlockMintablePair{
					1: {
						BlockNumber: 100,
						Mintable:    "invalid",
					},
				},
				ReserveInfo: &pb.ReserveInfo{
					ReserveAmount: "5000",
					Timestamp:     timestamppb.New(time.Unix(1640995200, 0).UTC()),
				},
				LatestBlocks: &pb.Blocks{
					Value: map[uint64]uint64{
						1: 100,
					},
				},
			},
			expectedError: true,
		},
		{
			name: "invalid reserve amount string",
			inputBlocks: core.Blocks{
				core.ChainSelector(1): core.BlockNumber(100),
			},
			mockResponse: &pb.ExternalAdapterPayload{
				Mintables: map[uint64]*pb.BlockMintablePair{
					1: {
						BlockNumber: 100,
						Mintable:    "1000",
					},
				},
				ReserveInfo: &pb.ReserveInfo{
					ReserveAmount: "invalid",
					Timestamp:     timestamppb.New(time.Unix(1640995200, 0).UTC()),
				},
				LatestBlocks: &pb.Blocks{
					Value: map[uint64]uint64{
						1: 100,
					},
				},
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client
			mockClient := new(mockExternalAdapterClient)
			if tt.mockError != nil {
				mockClient.On("GetPayload", mock.Anything, mock.Anything, mock.Anything).Return((*pb.ExternalAdapterPayload)(nil), tt.mockError)
			} else {
				mockClient.On("GetPayload", mock.Anything, mock.Anything, mock.Anything).Return(tt.mockResponse, nil)
			}

			// Create client with mock
			client := &externalAdapterClient{
				lggr: logger.Test(t),
				grpc: mockClient,
			}

			// Call GetPayload
			result, err := client.GetPayload(context.Background(), tt.inputBlocks)

			// Assertions
			if tt.expectedError {
				assert.Error(t, err)
				assert.Equal(t, core.ExternalAdapterPayload{}, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			// Verify mock expectations
			mockClient.AssertExpectations(t)
		})
	}
}

func TestExternalAdapterServer_GetPayload(t *testing.T) {
	tests := []struct {
		name           string
		inputRequest   *pb.Blocks
		mockResponse   core.ExternalAdapterPayload
		mockError      error
		expectedResult *pb.ExternalAdapterPayload
		expectedError  bool
	}{
		{
			name: "successful request with single chain",
			inputRequest: &pb.Blocks{
				Value: map[uint64]uint64{
					1: 100,
				},
			},
			mockResponse: core.ExternalAdapterPayload{
				Mintables: map[core.ChainSelector]core.BlockMintablePair{
					core.ChainSelector(1): {
						Block:    core.BlockNumber(100),
						Mintable: big.NewInt(1000),
					},
				},
				ReserveInfo: core.ReserveInfo{
					ReserveAmount: big.NewInt(5000),
					Timestamp:     time.Unix(1640995200, 0).UTC(),
				},
				LatestBlocks: core.Blocks{
					core.ChainSelector(1): core.BlockNumber(100),
				},
			},
			expectedResult: &pb.ExternalAdapterPayload{
				Mintables: map[uint64]*pb.BlockMintablePair{
					1: {
						BlockNumber: 100,
						Mintable:    "1000",
					},
				},
				ReserveInfo: &pb.ReserveInfo{
					ReserveAmount: "5000",
					Timestamp:     timestamppb.New(time.Unix(1640995200, 0).UTC()),
				},
				LatestBlocks: &pb.Blocks{
					Value: map[uint64]uint64{
						1: 100,
					},
				},
			},
			expectedError: false,
		},
		{
			name: "successful request with multiple chains",
			inputRequest: &pb.Blocks{
				Value: map[uint64]uint64{
					1: 100,
					2: 200,
				},
			},
			mockResponse: core.ExternalAdapterPayload{
				Mintables: map[core.ChainSelector]core.BlockMintablePair{
					core.ChainSelector(1): {
						Block:    core.BlockNumber(100),
						Mintable: big.NewInt(1000),
					},
					core.ChainSelector(2): {
						Block:    core.BlockNumber(200),
						Mintable: big.NewInt(2000),
					},
				},
				ReserveInfo: core.ReserveInfo{
					ReserveAmount: big.NewInt(5000),
					Timestamp:     time.Unix(1640995200, 0).UTC(),
				},
				LatestBlocks: core.Blocks{
					core.ChainSelector(1): core.BlockNumber(100),
					core.ChainSelector(2): core.BlockNumber(200),
				},
			},
			expectedResult: &pb.ExternalAdapterPayload{
				Mintables: map[uint64]*pb.BlockMintablePair{
					1: {
						BlockNumber: 100,
						Mintable:    "1000",
					},
					2: {
						BlockNumber: 200,
						Mintable:    "2000",
					},
				},
				ReserveInfo: &pb.ReserveInfo{
					ReserveAmount: "5000",
					Timestamp:     timestamppb.New(time.Unix(1640995200, 0).UTC()),
				},
				LatestBlocks: &pb.Blocks{
					Value: map[uint64]uint64{
						1: 100,
						2: 200,
					},
				},
			},
			expectedError: false,
		},
		{
			name: "empty input request",
			inputRequest: &pb.Blocks{
				Value: map[uint64]uint64{},
			},
			mockResponse: core.ExternalAdapterPayload{
				Mintables: map[core.ChainSelector]core.BlockMintablePair{},
				ReserveInfo: core.ReserveInfo{
					ReserveAmount: big.NewInt(0),
					Timestamp:     time.Unix(1640995200, 0).UTC(),
				},
				LatestBlocks: core.Blocks{},
			},
			expectedResult: &pb.ExternalAdapterPayload{
				Mintables: map[uint64]*pb.BlockMintablePair{},
				ReserveInfo: &pb.ReserveInfo{
					ReserveAmount: "0",
					Timestamp:     timestamppb.New(time.Unix(1640995200, 0).UTC()),
				},
				LatestBlocks: &pb.Blocks{
					Value: map[uint64]uint64{},
				},
			},
			expectedError: false,
		},
		{
			name: "external adapter error",
			inputRequest: &pb.Blocks{
				Value: map[uint64]uint64{
					1: 100,
				},
			},
			mockResponse:  core.ExternalAdapterPayload{},
			mockError:     errors.New("external adapter error"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock external adapter
			mockAdapter := new(mockExternalAdapter)
			if tt.mockError != nil {
				mockAdapter.On("GetPayload", mock.Anything, mock.Anything).Return(core.ExternalAdapterPayload{}, tt.mockError)
			} else {
				mockAdapter.On("GetPayload", mock.Anything, mock.Anything).Return(tt.mockResponse, nil)
			}

			// Create server with mock
			server := &externalAdapterServer{
				lggr: logger.Test(t),
				impl: mockAdapter,
			}

			// Call GetPayload
			result, err := server.GetPayload(context.Background(), tt.inputRequest)

			// Assertions
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult.LatestBlocks.Value, result.LatestBlocks.Value)
				assert.Equal(t, tt.expectedResult.ReserveInfo.ReserveAmount, result.ReserveInfo.ReserveAmount)
				assert.Equal(t, tt.expectedResult.ReserveInfo.Timestamp.AsTime(), result.ReserveInfo.Timestamp.AsTime())

				// loop through the mintables and assert that the values are equal
				assert.Equal(t, len(tt.expectedResult.Mintables), len(result.Mintables))
				for chainSelector, blockMintablePair := range tt.expectedResult.Mintables {
					assert.Equal(t, blockMintablePair.BlockNumber, result.Mintables[chainSelector].BlockNumber)
					assert.Equal(t, blockMintablePair.Mintable, result.Mintables[chainSelector].Mintable)
				}
			}

			// Verify mock expectations
			mockAdapter.AssertExpectations(t)
		})
	}
}

func TestStringToBigInt(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    *big.Int
		expectError bool
	}{
		{
			name:        "valid positive integer",
			input:       "123456789",
			expected:    big.NewInt(123456789),
			expectError: false,
		},
		{
			name:        "valid zero",
			input:       "0",
			expected:    big.NewInt(0),
			expectError: false,
		},
		{
			name:        "valid large integer",
			input:       "999999999999999999999999999999",
			expected:    func() *big.Int { z, _ := new(big.Int).SetString("999999999999999999999999999999", 10); return z }(),
			expectError: false,
		},
		{
			name:        "invalid string",
			input:       "invalid",
			expected:    nil,
			expectError: true,
		},
		{
			name:        "empty string",
			input:       "",
			expected:    nil,
			expectError: true,
		},
		{
			name:        "non-numeric string",
			input:       "abc123",
			expected:    nil,
			expectError: true,
		},
		{
			name:        "string with spaces",
			input:       " 123 ",
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := stringToBigInt(tt.input)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
