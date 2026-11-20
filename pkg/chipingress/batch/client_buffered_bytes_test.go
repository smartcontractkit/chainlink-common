package batch

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	cepb "github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
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
		require.ErrorIs(t, client.QueueMessage(bufferedBytesEvent("big", 4096), nil), ErrBufferedBytesExceeded)
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

func TestClient_BufferPressureMetrics(t *testing.T) {
	reader, restore := useTestMeterProvider(t)
	defer restore()

	gaugePoint := func(rm metricdata.ResourceMetrics, name, client string) *metricdata.DataPoint[int64] {
		t.Helper()
		g := mustMetric(t, rm, name).Data.(metricdata.Gauge[int64])
		for i := range g.DataPoints {
			if hasStringAttr(g.DataPoints[i].Attributes, "client_name", client) {
				return &g.DataPoints[i]
			}
		}
		t.Fatalf("%s datapoint for client %q not found", name, client)
		return nil
	}

	mockClient := mocks.NewClient(t)
	mockClient.EXPECT().PublishBatch(mock.Anything, mock.Anything).
		RunAndReturn(func(_ context.Context, in *chipingress.CloudEventBatch, _ ...grpc.CallOption) (*chipingress.PublishResponse, error) {
			return &chipingress.PublishResponse{Results: make([]*chipingress.PublishResult, len(in.Events))}, nil
		}).Maybe()
	mockClient.EXPECT().Close().Return(nil).Maybe()

	client, err := NewBatchClient(mockClient,
		WithClientName("beholder_data-feeds"),
		WithMessageBuffer(100),
		WithMaxBufferedBytes(1<<30),
	)
	require.NoError(t, err)
	// A second client without a byte budget: its bytes gauge must stay zero.
	noBudgetMock := mocks.NewClient(t)
	noBudgetMock.EXPECT().Close().Return(nil).Maybe()
	noBudget, err := NewBatchClient(noBudgetMock, WithClientName("no-budget"))
	require.NoError(t, err)

	// Not started: the batcher is not draining, so the queued messages stay in the buffer.
	const queued = 5
	for i := range queued {
		require.NoError(t, client.QueueMessage(bufferedBytesEvent(strconv.Itoa(i), 1024), nil))
	}
	require.NoError(t, noBudget.QueueMessage(bufferedBytesEvent("nb", 1024), nil))

	rm := collectResourceMetrics(t, reader)
	assert.Equal(t, int64(queued), gaugePoint(rm, "chip_ingress.batch.buffered_messages", "beholder_data-feeds").Value)
	assert.Equal(t, int64(1), gaugePoint(rm, "chip_ingress.batch.buffered_messages", "no-budget").Value)
	bytesPoint := gaugePoint(rm, "chip_ingress.batch.buffered_bytes", "beholder_data-feeds")
	assert.Equal(t, client.BufferedBytes(), bytesPoint.Value)
	assert.Positive(t, bytesPoint.Value, "reserved bytes are reported while a byte budget is set")
	assert.Zero(t, gaugePoint(rm, "chip_ingress.batch.buffered_bytes", "no-budget").Value, "no accounting without the byte budget")

	// config.info reports the configured byte budget; recorded on Start.
	client.Start(t.Context())
	rm = collectResourceMetrics(t, reader)
	cfg := gaugePoint(rm, "chip_ingress.batch.config.info", "beholder_data-feeds")
	assert.True(t, hasIntAttr(cfg.Attributes, "max_buffered_bytes", 1<<30))

	// Once the batcher drains, the message gauge returns to zero.
	require.Eventually(t, func() bool { return len(client.messageBuffer) == 0 }, time.Second, 5*time.Millisecond)
	rm = collectResourceMetrics(t, reader)
	assert.Zero(t, gaugePoint(rm, "chip_ingress.batch.buffered_messages", "beholder_data-feeds").Value)

	client.Stop()
	noBudget.Stop()
}
