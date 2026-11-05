package batch

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	cepb "github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/mocks"
)

func bufferedBytesEvent(id string, payload int) *chipingress.CloudEventPb {
	return &chipingress.CloudEventPb{
		Id:     id,
		Source: "platform",
		Type:   "BufferedBytes",
		Data:   &cepb.CloudEvent_BinaryData{BinaryData: []byte(strings.Repeat("x", payload))},
	}
}

func TestClient_MaxBufferedBytes(t *testing.T) {
	t.Run("default is unlimited", func(t *testing.T) {
		client, err := NewBatchClient(mocks.NewClient(t), WithMessageBuffer(10))
		require.NoError(t, err)
		assert.Zero(t, client.maxBufferedBytes)

		for i := range 10 {
			require.NoError(t, client.QueueMessage(bufferedBytesEvent(strings.Repeat("a", i+1), 1<<20), nil))
		}
		assert.Zero(t, client.BufferedBytes(), "no accounting without the option")
		assert.ErrorIs(t, client.QueueMessage(bufferedBytesEvent("full", 1), nil), ErrMessageBufferFull)
	})

	t.Run("negative is unlimited", func(t *testing.T) {
		client, err := NewBatchClient(mocks.NewClient(t), WithMaxBufferedBytes(-1))
		require.NoError(t, err)
		assert.Zero(t, client.maxBufferedBytes)
	})

	t.Run("rejects over budget and releases on send", func(t *testing.T) {
		release := make(chan struct{})
		mockClient := mocks.NewClient(t)
		mockClient.EXPECT().PublishBatch(mock.Anything, mock.Anything).
			RunAndReturn(func(_ context.Context, in *chipingress.CloudEventBatch, _ ...grpc.CallOption) (*chipingress.PublishResponse, error) {
				<-release
				return &chipingress.PublishResponse{}, nil
			}).Maybe()
		mockClient.EXPECT().Close().Return(nil).Maybe()

		client, err := NewBatchClient(mockClient,
			WithBatchSize(2),
			WithBatchInterval(10*time.Millisecond),
			WithMessageBuffer(100),
			WithTransactionEnabled(true),
			WithMaxBufferedBytes(1<<30),
		)
		require.NoError(t, err)

		done := make(chan error, 3)
		cb := func(err error) { done <- err }
		require.NoError(t, client.QueueMessage(bufferedBytesEvent("e1", 4096), cb))
		size := client.BufferedBytes()
		require.Positive(t, size)
		client.maxBufferedBytes = 2*size + size/2 // room for two events

		client.Start(t.Context())
		defer client.Stop()

		require.NoError(t, client.QueueMessage(bufferedBytesEvent("e2", 4096), cb))
		err = client.QueueMessage(bufferedBytesEvent("e3", 4096), cb)
		require.ErrorIs(t, err, ErrBufferedBytesExceeded)
		require.ErrorIs(t, err, ErrMessageBufferFull)
		assert.Equal(t, ErrMessageBufferFull.Error(), ErrorCodeFor(err))
		assert.InDelta(t, 2*size, client.BufferedBytes(), 4, "accepted bytes stay reserved while the send is in flight")

		close(release)
		for range 2 {
			require.NoError(t, <-done)
		}
		require.Eventually(t, func() bool { return client.BufferedBytes() == 0 }, time.Second, 5*time.Millisecond)
		require.NoError(t, client.QueueMessage(bufferedBytesEvent("e4", 4096), cb), "budget is reusable after release")
		require.NoError(t, <-done)
	})

	t.Run("releases on send failure", func(t *testing.T) {
		mockClient := mocks.NewClient(t)
		mockClient.EXPECT().PublishBatch(mock.Anything, mock.Anything).Return(nil, assert.AnError)
		mockClient.EXPECT().Close().Return(nil).Maybe()
		client, err := NewBatchClient(mockClient, WithBatchSize(1), WithMaxBufferedBytes(1<<20))
		require.NoError(t, err)
		client.Start(t.Context())
		defer client.Stop()

		done := make(chan error, 1)
		require.NoError(t, client.QueueMessage(bufferedBytesEvent("f", 128), func(err error) { done <- err }))
		require.Error(t, <-done)
		require.Eventually(t, func() bool { return client.BufferedBytes() == 0 }, time.Second, 5*time.Millisecond)
	})

	t.Run("rejects a message larger than the whole budget", func(t *testing.T) {
		client, err := NewBatchClient(mocks.NewClient(t), WithMaxBufferedBytes(1024))
		require.NoError(t, err)
		assert.ErrorIs(t, client.QueueMessage(bufferedBytesEvent("big", 4096), nil), ErrBufferedBytesExceeded)
		assert.Zero(t, client.BufferedBytes())
		require.NoError(t, client.QueueMessage(bufferedBytesEvent("small", 1), nil))
	})

	t.Run("concurrent reserve and release stay within budget and return to zero", func(t *testing.T) {
		const budget = 1000
		client, err := NewBatchClient(mocks.NewClient(t), WithMaxBufferedBytes(budget))
		require.NoError(t, err)
		var wg sync.WaitGroup
		for range 64 {
			wg.Go(func() {
				for range 1000 {
					if client.reserveBytes(100) {
						assert.LessOrEqual(t, client.BufferedBytes(), int64(budget))
						client.releaseBytes(100)
					}
				}
			})
		}
		wg.Wait()
		assert.Zero(t, client.BufferedBytes())
	})

	t.Run("releases when the channel is full", func(t *testing.T) {
		client, err := NewBatchClient(mocks.NewClient(t), WithMessageBuffer(1), WithMaxBufferedBytes(1<<20))
		require.NoError(t, err)
		require.NoError(t, client.QueueMessage(bufferedBytesEvent("a", 64), nil))
		held := client.BufferedBytes()
		require.ErrorIs(t, client.QueueMessage(bufferedBytesEvent("b", 64), nil), ErrMessageBufferFull)
		assert.Equal(t, held, client.BufferedBytes())
	})

	t.Run("releases on shutdown", func(t *testing.T) {
		mockClient := mocks.NewClient(t)
		mockClient.EXPECT().Close().Return(nil)
		client, err := NewBatchClient(mockClient, WithMaxBufferedBytes(1<<20))
		require.NoError(t, err)
		client.afterFastPath = client.Stop
		require.ErrorIs(t, client.QueueMessage(bufferedBytesEvent("a", 64), nil), ErrClientShutdown)
		assert.Zero(t, client.BufferedBytes())
	})
}
