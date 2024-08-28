package oracle

import (
	"context"

	"github.com/smartcontractkit/libocr/offchainreporting2plus"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	oraclepb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/oracle"
)

var _ oraclepb.OracleServer = (*server)(nil)

type server struct {
	oraclepb.UnimplementedOracleServer

	broker *net.BrokerExt
	impl   offchainreporting2plus.Oracle
	log    logger.Logger

	Name string
}

func NewServer(log logger.Logger, impl offchainreporting2plus.Oracle, broker *net.BrokerExt) (*server, net.Resource) {
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

func (s *server) OracleClose(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	err := s.impl.Close()
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *server) OracleStart(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	err := s.impl.Start()
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
