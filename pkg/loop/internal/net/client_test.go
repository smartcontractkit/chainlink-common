package net

import (
	"context"
	"errors"
	"fmt"
	stdnet "net"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

type testBroker struct {
	dial func(id uint32, opts ...grpc.DialOption) (*grpc.ClientConn, error)
}

func (b testBroker) Accept(uint32) (stdnet.Listener, error) {
	return nil, errors.New("unused in test")
}

func (b testBroker) DialWithOptions(id uint32, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return b.dial(id, opts...)
}

func (b testBroker) NextId() uint32 { return 1 }

func TestWrappedErrorError(t *testing.T) {
	t.Parallel()
	t.Run("Is returns false for different error code", func(t *testing.T) {
		// silly, but to verify that it's only looking at error code here, we need to make the message the same
		err := types.NotFoundError(types.ErrInvalidType.Error())
		assert.False(t, errors.Is(WrapRPCErr(types.ErrInvalidType), err))
	})

	t.Run("Is returns false for different message", func(t *testing.T) {
		// Both are InvalidArgumentError
		assert.False(t, errors.Is(WrapRPCErr(types.ErrInvalidType), types.ErrInvalidEncoding))
	})

	t.Run("Is returns true if the message and code are the same", func(t *testing.T) {
		assert.True(t, errors.Is(WrapRPCErr(types.ErrInvalidType), types.ErrInvalidType))
	})

	t.Run("Is returns true if the message is contained and the code is the same", func(t *testing.T) {
		wrapped := WrapRPCErr(fmt.Errorf("%w: %w", types.ErrInvalidType, errors.New("some other error")))
		assert.True(t, errors.Is(wrapped, types.ErrInvalidType))
	})
}

func TestClientConnRefresh_PreservesPublishedDepsUntilReplacementSucceeds(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	oldConn := newBufconnClientConn(t)
	newConn := newBufconnClientConn(t)

	var oldDepClosed atomic.Int32
	var failedAttemptDepClosed atomic.Int32
	var newDepClosed atomic.Int32
	var dialAttempts atomic.Uint32

	client := &clientConn{
		BrokerExt: &BrokerExt{Broker: testBroker{dial: func(id uint32, _ ...grpc.DialOption) (*grpc.ClientConn, error) {
			switch id {
			case 1:
				assert.Equal(t, int32(0), oldDepClosed.Load(), "published deps must remain live while a replacement attempt fails")
				return nil, errors.New("boom")
			case 2:
				return newConn, nil
			default:
				return nil, fmt.Errorf("unexpected dial id %d", id)
			}
		}}, BrokerConfig: BrokerConfig{Logger: logger.Test(t)}},
		name: "Relayer",
		cc:   oldConn,
		deps: Resources{{
			Closer: fnCloser(func() { oldDepClosed.Add(1) }),
			Name:   "old",
		}},
	}

	client.newClient = func(context.Context) (uint32, Resources, error) {
		switch dialAttempts.Add(1) {
		case 1:
			return 1, Resources{{
				Closer: fnCloser(func() { failedAttemptDepClosed.Add(1) }),
				Name:   "failed-attempt",
			}}, nil
		case 2:
			return 2, Resources{{
				Closer: fnCloser(func() { newDepClosed.Add(1) }),
				Name:   "new",
			}}, nil
		default:
			return 0, nil, errors.New("unexpected extra refresh attempt")
		}
	}

	cc, err := client.refresh(ctx, oldConn)
	require.NoError(t, err)
	require.NotNil(t, cc)
	assert.Equal(t, int32(1), failedAttemptDepClosed.Load(), "failed attempt deps should be cleaned up immediately")
	assert.Equal(t, int32(1), oldDepClosed.Load(), "old deps should be closed only after the replacement is published")
	assert.Equal(t, int32(0), newDepClosed.Load(), "replacement deps must remain live after cutover")
	assert.Len(t, client.deps, 1)
	assert.Equal(t, "new", client.deps[0].Name)
	assert.Same(t, newConn, cc)
	assert.Same(t, newConn, client.cc)
	assert.Equal(t, uint32(2), dialAttempts.Load())
}

func newBufconnClientConn(t *testing.T) *grpc.ClientConn {
	t.Helper()

	lis := bufconn.Listen(1024)
	server := grpc.NewServer()
	go func() {
		_ = server.Serve(lis)
	}()
	t.Cleanup(func() {
		server.Stop()
	})

	conn, err := grpc.DialContext(context.Background(), "bufnet",
		grpc.WithContextDialer(func(context.Context, string) (stdnet.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = conn.Close()
	})
	return conn
}
