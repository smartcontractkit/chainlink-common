package net_test

import (
	"context"
	"testing"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net/test"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// fnCloser implements io.Closer with a func().
type fnCloser func()

func (s fnCloser) Close() error {
	s()
	return nil
}

type lifecycleClient interface {
	grpc.ClientConnInterface
	Close() error
}

type lifecycleTestSetup struct {
	lifecycleClient
	newClientCalls int
}

func setupLifecycleTest(t *testing.T, name string) *lifecycleTestSetup {
	broker := &test.Broker{T: t}
	brokerExt := &net.BrokerExt{
		Broker: broker,
		BrokerConfig: net.BrokerConfig{
			Logger: logger.Test(t),
		},
	}

	setup := &lifecycleTestSetup{}
	newClientFn := func(ctx context.Context) (uint32, net.Resources, error) {
		setup.newClientCalls++
		id := broker.NextId()
		lis, err := broker.Accept(id)
		if err != nil {
			return 0, nil, err
		}
		s := grpc.NewServer()
		go func() {
			if err := s.Serve(lis); err != nil {
				// ignore
			}
		}()

		res := net.Resource{Closer: fnCloser(func() { s.Stop() }), Name: "Server"}
		var resources net.Resources
		resources.Add(res)

		return id, resources, nil
	}

	c := brokerExt.NewClientConn(name, newClientFn)
	t.Cleanup(func() { c.Close() })
	setup.lifecycleClient = c

	return setup
}

func TestClient_Lifecycle(t *testing.T) {
	t.Parallel()

	t.Run("Invoke Context Cancellation", func(t *testing.T) {
		setup := setupLifecycleTest(t, "TestClient")

		// 3. Establish connection by calling Invoke() on a non-existent method.
		ctx := context.Background()
		// Method "Foo" doesn't exist, should return Unimplemented (code 12), which is not a terminal error.
		err := setup.Invoke(ctx, "/Service/Foo", nil, nil)
		require.Error(t, err)
		require.Equal(t, codes.Unimplemented, status.Code(err))
		require.Equal(t, 1, setup.newClientCalls, "Should have called newClientFn once to establish connection")

		// 4. Now call with CANCELLED context.
		ctxCancelled, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		// This call should return Canceled immediately and NOT trigger refresh.
		err = setup.Invoke(ctxCancelled, "/Service/Foo", nil, nil)
		require.ErrorIs(t, err, context.Canceled)

		// If the bug exists, newClientFn will be called AGAIN (incrementing to 2).
		// If fixed, it should remain 1.
		require.Equal(t, 1, setup.newClientCalls, "Should NOT refresh connection on context cancellation")
	})

	t.Run("Invoke Context Deadline Exceeded", func(t *testing.T) {
		setup := setupLifecycleTest(t, "TestClient_Deadline")

		// 3. Establish connection first
		ctx := context.Background()
		err := setup.Invoke(ctx, "/Service/Foo", nil, nil)
		require.Error(t, err)
		require.Equal(t, codes.Unimplemented, status.Code(err))
		require.Equal(t, 1, setup.newClientCalls, "Should have called newClientFn once to establish connection")

		// 4. Now call with TIMED OUT context
		// We use a very short timeout and wait for it to expire
		ctxTimeout, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()
		<-ctxTimeout.Done() // Wait for timeout

		// This call should return DeadlineExceeded immediately and NOT trigger refresh.
		err = setup.Invoke(ctxTimeout, "/Service/Foo", nil, nil)

		require.Equal(t, codes.DeadlineExceeded, status.Code(err))
		require.Equal(t, 1, setup.newClientCalls, "Should NOT refresh connection on context deadline exceeded")
	})

	t.Run("NewStream Context Cancellation", func(t *testing.T) {
		setup := setupLifecycleTest(t, "TestClient_Stream")

		// 3. Establish connection first
		ctx := context.Background()
		err := setup.Invoke(ctx, "/Service/Probe", nil, nil) // Force connection
		require.Error(t, err)                                // Unimplemented

		// 4. Now call with CANCELLED context
		ctxCancelled, cancel := context.WithCancel(context.Background())
		cancel()

		stream, err := setup.NewStream(ctxCancelled, &grpc.StreamDesc{StreamName: "Foo", ClientStreams: true, ServerStreams: true}, "/Service/Foo")

		require.Error(t, err)
		require.ErrorIs(t, err, context.Canceled)
		require.Nil(t, stream)

		require.Equal(t, 1, setup.newClientCalls, "Should NOT refresh connection on NewStream context cancellation")
	})

	t.Run("NewStream Context Deadline Exceeded", func(t *testing.T) {
		setup := setupLifecycleTest(t, "TestClient_Stream_Deadline")

		// 3. Establish connection first
		ctx := context.Background()
		err := setup.Invoke(ctx, "/Service/Probe", nil, nil) // Force connection
		require.Error(t, err)                                // Unimplemented

		// 4. Now call with TIMED OUT context
		ctxTimeout, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()
		<-ctxTimeout.Done()

		stream, err := setup.NewStream(ctxTimeout, &grpc.StreamDesc{StreamName: "Foo", ClientStreams: true, ServerStreams: true}, "/Service/Foo")

		require.Error(t, err)
		require.Equal(t, codes.DeadlineExceeded, status.Code(err))
		require.Nil(t, stream)

		require.Equal(t, 1, setup.newClientCalls, "Should NOT refresh connection on NewStream context deadline exceeded")
	})
}
