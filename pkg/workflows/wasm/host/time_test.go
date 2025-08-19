package host

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
)

func TestTimeFetcher_GetTime_NODE(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockExec := NewMockExecutionHelper(t)
	expected := time.Now()
	mockExec.EXPECT().GetNodeTime().Return(expected)

	tf := newTimeFetcher(ctx, mockExec)
	tf.Start()

	actual, err := tf.GetTime(sdk.Mode_MODE_NODE)
	require.NoError(t, err)
	require.WithinDuration(t, expected, actual, time.Millisecond)
}

func TestTimeFetcher_GetTime_DON(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockExec := NewMockExecutionHelper(t)
	expected := time.Now()
	mockExec.EXPECT().GetDONTime().Return(expected, nil)

	tf := newTimeFetcher(ctx, mockExec)
	tf.Start()

	actual, err := tf.GetTime(sdk.Mode_MODE_DON)
	require.NoError(t, err)
	require.WithinDuration(t, expected, actual, time.Millisecond)
}

func TestTimeFetcher_GetTime_DON_Error(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockExec := NewMockExecutionHelper(t)
	mockExec.EXPECT().GetDONTime().Return(time.Time{}, errors.New("don error"))

	tf := newTimeFetcher(ctx, mockExec)
	tf.Start()

	_, err := tf.GetTime(sdk.Mode_MODE_DON)
	require.ErrorContains(t, err, "don error")
}

func TestTimeFetcher_ContextCancelledBeforeRequest(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	mockExec := NewMockExecutionHelper(t)
	mockExec.EXPECT().GetDONTime().Return(time.Time{}, context.Canceled).Maybe()

	tf := newTimeFetcher(ctx, mockExec)

	done := make(chan struct{})
	go func() {
		defer close(done)
		tf.runLoop()
	}()

	t.Cleanup(func() {
		<-done // wait for runLoop to exit to avoid mock usage after test ends
	})

	_, err := tf.GetTime(sdk.Mode_MODE_DON)
	require.ErrorIs(t, err, context.Canceled)
}

func TestTimeFetcher_ContextCancelledDuringResponse(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	mockExec := NewMockExecutionHelper(t)
	mockExec.EXPECT().GetDONTime().Run(func() {
		time.Sleep(20 * time.Millisecond) // force timeout
	}).Return(time.Time{}, nil)

	tf := newTimeFetcher(ctx, mockExec)
	tf.Start()

	_, err := tf.GetTime(sdk.Mode_MODE_DON)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}
