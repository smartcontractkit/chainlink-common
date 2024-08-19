package ccip

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
	ccippb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/ccip"

	"github.com/smartcontractkit/chainlink-common/pkg/types/ccip"
	"github.com/smartcontractkit/chainlink-common/pkg/types/ccip/mocks"
)

func Test_OnchainConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		onchainConfig ccip.ExecOnchainConfig
		want          ccippb.ExecOnchainConfig
	}{
		{
			name:          "empty",
			onchainConfig: ccip.ExecOnchainConfig{},
			want: ccippb.ExecOnchainConfig{
				PermissionlessExecThresholdSeconds: &durationpb.Duration{Seconds: 0},
			},
		},
		{
			name: "normal",
			onchainConfig: ccip.ExecOnchainConfig{
				PermissionLessExecutionThresholdSeconds: 34434 * time.Second,
				Router:                                  "0x123",
				MaxDataBytes:                            4,
				MaxNumberOfTokensPerMsg:                 5,
				PriceRegistry:                           "0x165623",
				MaxPoolReleaseOrMintGas:                 7,
				MaxTokenTransferGas:                     8,
			},
			want: ccippb.ExecOnchainConfig{
				PermissionlessExecThresholdSeconds: &durationpb.Duration{Seconds: 34434},
				Router:                             "0x123",
				MaxDataBytes:                       4,
				MaxNumberOfTokensPerMsg:            5,
				PriceRegistry:                      "0x165623",
				MaxPoolReleaseOrMintGas:            7,
				MaxTokenTransferGas:                8,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			offRampReaderMock := mocks.NewOffRampReader(t)
			offRampReaderMock.On("OnchainConfig", mock.Anything).Return(tt.onchainConfig, nil)
			ctx := context.Background()
			server := OffRampReaderGRPCServer{
				impl: offRampReaderMock,
			}

			got, err := server.OnchainConfig(ctx, nil)
			require.NoError(t, err)
			assert.Equal(t, tt.want.PermissionlessExecThresholdSeconds.Seconds, got.Config.PermissionlessExecThresholdSeconds.Seconds)
			assert.Equal(t, tt.want.Router, got.Config.Router)
			assert.Equal(t, tt.want.MaxDataBytes, got.Config.MaxDataBytes)
			assert.Equal(t, tt.want.MaxNumberOfTokensPerMsg, got.Config.MaxNumberOfTokensPerMsg)
			assert.Equal(t, tt.want.PriceRegistry, got.Config.PriceRegistry)
			assert.Equal(t, tt.want.MaxPoolReleaseOrMintGas, got.Config.MaxPoolReleaseOrMintGas)
			assert.Equal(t, tt.want.MaxTokenTransferGas, got.Config.MaxTokenTransferGas)
		})
	}
}

func Test_OffchainConfig(t *testing.T) {
	t.Parallel()

	duration1, _ := config.NewDuration(111 * time.Second)
	duration2, _ := config.NewDuration(222 * time.Second)
	duration3, _ := config.NewDuration(333 * time.Second)

	tests := []struct {
		name           string
		offchainConfig ccip.ExecOffchainConfig
		want           ccippb.ExecOffchainConfig
	}{
		{
			name:           "empty",
			offchainConfig: ccip.ExecOffchainConfig{},
			want: ccippb.ExecOffchainConfig{
				InflightCacheExpiry:       &durationpb.Duration{Seconds: 0},
				RootSnoozeTime:            &durationpb.Duration{Seconds: 0},
				MessageVisibilityInterval: &durationpb.Duration{Seconds: 0},
			},
		},
		{
			name: "normal",
			offchainConfig: ccip.ExecOffchainConfig{
				DestOptimisticConfirmations: 10,
				BatchGasLimit:               10000,
				RelativeBoostPerWaitHour:    5,
				InflightCacheExpiry:         duration1,
				RootSnoozeTime:              duration2,
				MessageVisibilityInterval:   duration3,
				BatchingStrategyID:          1,
			},
			want: ccippb.ExecOffchainConfig{
				DestOptimisticConfirmations: 10,
				BatchGasLimit:               10000,
				RelativeBoostPerWaitHour:    5,
				InflightCacheExpiry:         &durationpb.Duration{Seconds: 111},
				RootSnoozeTime:              &durationpb.Duration{Seconds: 222},
				MessageVisibilityInterval:   &durationpb.Duration{Seconds: 333},
				BatchingStrategyId:          1,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			offRampReaderMock := mocks.NewOffRampReader(t)
			offRampReaderMock.On("OffchainConfig", mock.Anything).Return(tt.offchainConfig, nil)
			ctx := context.Background()
			server := OffRampReaderGRPCServer{
				impl: offRampReaderMock,
			}

			got, err := server.OffchainConfig(ctx, nil)
			require.NoError(t, err)

			assert.Equal(t, tt.want.DestOptimisticConfirmations, got.Config.DestOptimisticConfirmations)
			assert.Equal(t, tt.want.BatchGasLimit, got.Config.BatchGasLimit)
			assert.Equal(t, tt.want.RelativeBoostPerWaitHour, got.Config.RelativeBoostPerWaitHour)
			assert.Equal(t, tt.want.BatchingStrategyId, got.Config.BatchingStrategyId)

			assert.Equal(t, tt.want.InflightCacheExpiry.Seconds, got.Config.InflightCacheExpiry.Seconds)
			assert.Equal(t, tt.want.RootSnoozeTime.Seconds, got.Config.RootSnoozeTime.Seconds)
			assert.Equal(t, tt.want.MessageVisibilityInterval.Seconds, got.Config.MessageVisibilityInterval.Seconds)

			// Decode the offchain config back to original struct
			decodedConfig, err := offChainConfig(got.Config)
			require.NoError(t, err)

			// Ensure the decoded config matches the original input config
			assert.Equal(t, tt.offchainConfig, decodedConfig)
		})
	}
}

func Test_byte32Slice(t *testing.T) {
	tooLong := make([]byte, 33)
	tooLong[32] = 32
	tooShort := make([]byte, 31)
	tooShort[30] = 30
	type args struct {
		pbVal [][]byte
	}
	tests := []struct {
		name     string
		args     args
		ifaceVal [][32]byte
		wantErr  bool
	}{
		{name: "empty", args: args{pbVal: [][]byte{}}, ifaceVal: [][32]byte{}, wantErr: false},
		{name: "non-empty",
			args: args{
				pbVal: [][]byte{
					{0: 1, 31: 2},
					{0: 3, 31: 4},
				},
			},
			ifaceVal: [][32]byte{
				{0: 1, 31: 2},
				{0: 3, 31: 4},
			},
			wantErr: false},
		{name: "too long", args: args{pbVal: [][]byte{tooLong}}, ifaceVal: nil, wantErr: true},
		{name: "too short", args: args{pbVal: [][]byte{tooShort}}, ifaceVal: nil, wantErr: true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("pb-to-iface %s", tt.name), func(t *testing.T) {
			t.Parallel()

			got, err := byte32Slice(tt.args.pbVal)
			if (err != nil) != tt.wantErr {
				t.Errorf("byte32Slice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.ifaceVal) {
				t.Errorf("byte32Slice() = %v, want %v", got, tt.ifaceVal)
			}
		})

		t.Run(fmt.Sprintf("iface-to-pb %s", tt.name), func(t *testing.T) {
			t.Parallel()

			// there are no errors in this direction so skip tests that expect errors
			if tt.wantErr {
				return
			}
			got := byte32SliceToPB(tt.ifaceVal)

			if !reflect.DeepEqual(got, tt.args.pbVal) {
				t.Errorf("byte32SlicePB() = %v, want %v", got, tt.args.pbVal)
			}
		})
	}

	// special case for nil
	t.Run("nil pb-to-iface", func(t *testing.T) {
		t.Parallel()
		got, err := byte32Slice(nil)
		if err != nil {
			t.Errorf("byte32Slice() error = %v, wantErr %v", err, false)
			return
		}
		expected := [][32]byte{}
		if !reflect.DeepEqual(got, expected) {
			t.Errorf("byte32Slice() = %v, want %v", got, expected)
		}
	})

	t.Run("nil iface-to-pb", func(t *testing.T) {
		t.Parallel()
		// there are no errors in this direction so skip tests that expect errors
		got := byte32SliceToPB(nil)
		expected := [][]byte{}
		if !reflect.DeepEqual(got, expected) {
			t.Errorf("byte32SlicePB() = %v, want %v", got, expected)
		}
	})
}
