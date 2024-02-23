package test

import (
	"context"
	"net"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"github.com/smartcontractkit/chainlink-common/pkg/types/interfacetests"
)

type loopServerTester[T interfacetests.TestingT[T]] struct {
	lis          *bufconn.Listener
	server       *grpc.Server
	conn         *grpc.ClientConn
	registerHook func(server *grpc.Server)
}

func (lst *loopServerTester[T]) Setup(t T) {
	lis := bufconn.Listen(1024 * 1024)
	lst.lis = lis
	s := grpc.NewServer()
	lst.registerHook(s)
	go func() {
		if err := s.Serve(lis); err != nil {
			require.NoError(t, err)
		}
	}()

	t.Cleanup(func() {
		if lst.server != nil {
			lst.server.Stop()
		}

		if lst.conn != nil {
			require.NoError(t, lst.conn.Close())
		}

		lst.lis = nil
		lst.server = nil
		lst.conn = nil
	})
}

func (lst *loopServerTester[T]) GetConn(t T) *grpc.ClientConn {
	if lst.conn != nil {
		return lst.conn
	}

	conn, err := grpc.Dial("bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return lst.lis.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	require.NoError(t, err)
	lst.conn = conn
	return conn
}
