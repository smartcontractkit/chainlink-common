package batch

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/mocks"
)

func noAckTestMessages(ids ...string) ([]*messageWithCallback, chan error) {
	callbackErrs := make(chan error, len(ids))
	msgs := make([]*messageWithCallback, len(ids))
	for i, id := range ids {
		msgs[i] = &messageWithCallback{
			event:    &chipingress.CloudEventPb{Id: id, Source: "test-source", Type: "test.event.type"},
			callback: func(err error) { callbackErrs <- err },
		}
	}
	return msgs, callbackErrs
}

// okResponse answers a batch with one successful positional result per event, as the
// server does for both PublishBatch and PublishBatchNoAck.
func okResponse(_ context.Context, b *chipingress.CloudEventBatch, _ ...grpc.CallOption) (*chipingress.PublishResponse, error) {
	results := make([]*chipingress.PublishResult, len(b.Events))
	for i, e := range b.Events {
		results[i] = &chipingress.PublishResult{EventId: e.Id}
	}
	return &chipingress.PublishResponse{Results: results}, nil
}

func collectCallbackErrs(t *testing.T, ch chan error, n int) []error {
	t.Helper()
	errs := make([]error, 0, n)
	for range n {
		select {
		case err := <-ch:
			errs = append(errs, err)
		case <-time.After(time.Second):
			t.Fatalf("timeout waiting for callback %d of %d", len(errs)+1, n)
		}
	}
	return errs
}

func TestWithNoAck(t *testing.T) {
	t.Run("option sets noAck, default is false", func(t *testing.T) {
		client, err := NewBatchClient(mocks.NewClient(t))
		require.NoError(t, err)
		assert.False(t, client.noAck)

		client, err = NewBatchClient(mocks.NewClient(t), WithNoAck())
		require.NoError(t, err)
		assert.True(t, client.noAck)
	})

	t.Run("routes to PublishBatchNoAck and never PublishBatch", func(t *testing.T) {
		mockClient := mocks.NewClient(t)
		mockClient.EXPECT().PublishBatchNoAck(mock.Anything, mock.Anything).
			RunAndReturn(okResponse).Once()

		client, err := NewBatchClient(mockClient, WithNoAck())
		require.NoError(t, err)

		msgs, cbs := noAckTestMessages("a", "b")
		client.sendBatch(t.Context(), msgs)
		for _, err := range collectCallbackErrs(t, cbs, 2) {
			assert.NoError(t, err)
		}

		mockClient.AssertNotCalled(t, "PublishBatch", mock.Anything, mock.Anything)
		mockClient.AssertNumberOfCalls(t, "PublishBatchNoAck", 1)
	})

	t.Run("default client routes to PublishBatch and never PublishBatchNoAck", func(t *testing.T) {
		mockClient := mocks.NewClient(t)
		mockClient.EXPECT().PublishBatch(mock.Anything, mock.Anything).
			RunAndReturn(okResponse).Once()

		client, err := NewBatchClient(mockClient)
		require.NoError(t, err)

		msgs, cbs := noAckTestMessages("a")
		client.sendBatch(t.Context(), msgs)
		collectCallbackErrs(t, cbs, 1)

		mockClient.AssertNotCalled(t, "PublishBatchNoAck", mock.Anything, mock.Anything)
		mockClient.AssertNumberOfCalls(t, "PublishBatch", 1)
	})

	t.Run("sends the same request as PublishBatch", func(t *testing.T) {
		for _, txn := range []bool{false, true} {
			mockClient := mocks.NewClient(t)
			var got *chipingress.CloudEventBatch
			mockClient.EXPECT().PublishBatchNoAck(mock.Anything, mock.Anything).
				RunAndReturn(func(_ context.Context, b *chipingress.CloudEventBatch, _ ...grpc.CallOption) (*chipingress.PublishResponse, error) {
					got = b
					return &chipingress.PublishResponse{}, nil
				}).Once()

			client, err := NewBatchClient(mockClient, WithNoAck(), WithTransactionEnabled(txn))
			require.NoError(t, err)

			msgs, cbs := noAckTestMessages("a", "b")
			client.sendBatch(t.Context(), msgs)
			collectCallbackErrs(t, cbs, 2)

			require.NotNil(t, got)
			assert.Len(t, got.Events, 2)
			require.NotNil(t, got.Options)
			require.NotNil(t, got.Options.TransactionEnabled)
			assert.Equal(t, txn, *got.Options.TransactionEnabled)
		}
	})

	t.Run("per-event results reach callbacks by position with the same error codes", func(t *testing.T) {
		mockClient := mocks.NewClient(t)
		mockClient.EXPECT().PublishBatchNoAck(mock.Anything, mock.Anything).
			Return(&chipingress.PublishResponse{Results: []*chipingress.PublishResult{
				{EventId: "ok"},
				{EventId: "bad", Error: &chipingress.PublishError{
					ErrorCode: chipingress.PublishErrorCode(ErrCodeValidationFailed),
					Reason:    "invalid event",
				}},
			}}, nil).Once()

		client, err := NewBatchClient(mockClient, WithNoAck())
		require.NoError(t, err)

		type outcome struct {
			id  string
			err error
		}
		outcomes := make(chan outcome, 2)
		msgs := []*messageWithCallback{
			{event: &chipingress.CloudEventPb{Id: "ok"}, callback: func(err error) { outcomes <- outcome{"ok", err} }},
			{event: &chipingress.CloudEventPb{Id: "bad"}, callback: func(err error) { outcomes <- outcome{"bad", err} }},
		}
		client.sendBatch(t.Context(), msgs)

		// callbacks complete concurrently, so key the outcomes by event rather than order
		byID := map[string]error{}
		for range 2 {
			select {
			case o := <-outcomes:
				byID[o.id] = o.err
			case <-time.After(time.Second):
				t.Fatal("timeout waiting for callbacks")
			}
		}
		results := []error{byID["ok"], byID["bad"]}

		assert.NoError(t, results[0])
		var pubErr *PublishError
		require.ErrorAs(t, results[1], &pubErr)
		assert.Equal(t, ErrCodeValidationFailed, pubErr.Code)
		assert.Equal(t, "invalid event", pubErr.Reason)
	})

	t.Run("a result count mismatch is still reported", func(t *testing.T) {
		mockClient := mocks.NewClient(t)
		mockClient.EXPECT().PublishBatchNoAck(mock.Anything, mock.Anything).
			Return(&chipingress.PublishResponse{}, nil).Once()

		client, err := NewBatchClient(mockClient, WithNoAck())
		require.NoError(t, err)

		msgs, cbs := noAckTestMessages("a")
		client.sendBatch(t.Context(), msgs)

		var pubErr *PublishError
		require.ErrorAs(t, collectCallbackErrs(t, cbs, 1)[0], &pubErr)
		assert.Equal(t, ErrCodeResultsMismatch, pubErr.Code)
	})

	t.Run("a transport error fails every callback, Unimplemented included", func(t *testing.T) {
		mockClient := mocks.NewClient(t)
		mockClient.EXPECT().PublishBatchNoAck(mock.Anything, mock.Anything).
			Return(nil, status.Error(codes.Unimplemented, "method PublishBatchNoAck not implemented")).Once()

		client, err := NewBatchClient(mockClient, WithNoAck())
		require.NoError(t, err)

		msgs, cbs := noAckTestMessages("a", "b")
		client.sendBatch(t.Context(), msgs)

		for _, err := range collectCallbackErrs(t, cbs, 2) {
			require.Error(t, err)
			assert.Equal(t, codes.Unimplemented.String(), ErrorCodeFor(err))
		}
		// no fallback to the acked RPC
		mockClient.AssertNotCalled(t, "PublishBatch", mock.Anything, mock.Anything)
	})

	t.Run("oversized batches are split across multiple PublishBatchNoAck calls", func(t *testing.T) {
		events := []*chipingress.CloudEventPb{
			largeTestEvent("noack-split-1"),
			largeTestEvent("noack-split-2"),
			largeTestEvent("noack-split-3"),
		}
		// 2 events fit, 3 do not.
		_, maxRequestSize := newBatchRequest([]*messageWithCallback{{event: events[0]}, {event: events[1]}}, false)

		mockClient := mocks.NewClient(t)
		mockClient.EXPECT().PublishBatchNoAck(mock.Anything, mock.Anything).
			RunAndReturn(okResponse).Twice()

		client, err := NewBatchClient(mockClient, WithNoAck())
		require.NoError(t, err)
		client.maxGRPCRequestSize = maxRequestSize
		client.effectiveMaxRequestSize = maxRequestSize

		cbs := make(chan error, len(events))
		msgs := make([]*messageWithCallback, len(events))
		for i, e := range events {
			msgs[i] = &messageWithCallback{event: e, callback: func(err error) { cbs <- err }}
		}
		client.sendBatch(t.Context(), msgs)
		collectCallbackErrs(t, cbs, len(events))

		mockClient.AssertNumberOfCalls(t, "PublishBatchNoAck", 2)
		mockClient.AssertNotCalled(t, "PublishBatch", mock.Anything, mock.Anything)
	})

	t.Run("no_ack attribute is emitted only for NoAck clients", func(t *testing.T) {
		for _, noAck := range []bool{false, true} {
			reader, restore := useTestMeterProvider(t)

			mockClient := mocks.NewClient(t)
			opts := []Opt{}
			if noAck {
				mockClient.EXPECT().PublishBatchNoAck(mock.Anything, mock.Anything).
					RunAndReturn(okResponse).Once()
				opts = append(opts, WithNoAck())
			} else {
				mockClient.EXPECT().PublishBatch(mock.Anything, mock.Anything).
					RunAndReturn(okResponse).Once()
			}

			client, err := NewBatchClient(mockClient, opts...)
			require.NoError(t, err)
			client.metrics.recordConfig(t.Context(), client)

			msgs, cbs := noAckTestMessages("a")
			client.sendBatch(t.Context(), msgs)
			collectCallbackErrs(t, cbs, 1)

			rm := collectResourceMetrics(t, reader)
			restore()

			reqSum, ok := mustMetric(t, rm, "chip_ingress.batch.send_requests_total").Data.(metricdata.Sum[int64])
			require.True(t, ok)
			point := mustInt64SumPointWithAttr(t, reqSum, "status", "success")
			assertNoAckAttr(t, point.Attributes, noAck)

			cfg, ok := mustMetric(t, rm, "chip_ingress.batch.config.info").Data.(metricdata.Gauge[int64])
			require.True(t, ok)
			require.Len(t, cfg.DataPoints, 1)
			assertNoAckAttr(t, cfg.DataPoints[0].Attributes, noAck)
		}
	})
}

// assertNoAckAttr checks no_ack=true is present when want, and that the key is absent
// otherwise (existing series must keep their label set).
func assertNoAckAttr(t *testing.T, set attribute.Set, want bool) {
	t.Helper()
	v, found := set.Value("no_ack")
	assert.Equal(t, want, found)
	if want {
		assert.True(t, v.AsBool())
	}
}
