package oracle

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	oraclepb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/oracle"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

var _ oraclepb.OracleServer = (*server)(nil)

type server struct {
	oraclepb.UnimplementedOracleServer

	broker *net.BrokerExt
	impl   core.Oracle
	log    logger.Logger

	Name string
}

func NewServer(log logger.Logger, impl core.Oracle, broker *net.BrokerExt) (*server, net.Resource) {
	name := "OracleServer"
	newServer := &server{
		log:    log,
		impl:   impl,
		broker: broker.WithName(name),
	}

	return newServer, net.Resource{
		Name:   name,
		Closer: newServer,
	}
}

func (s *server) Close() error {
	return nil
}

func (s *server) OracleClose(ctx context.Context, e *emptypb.Empty) (*emptypb.Empty, error) {
	return e, s.impl.Close(ctx)
}

func (s *server) OracleStart(ctx context.Context, e *emptypb.Empty) (*emptypb.Empty, error) {
	return e, s.impl.Start(ctx)
}
