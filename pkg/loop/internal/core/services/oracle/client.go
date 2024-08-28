package oracle

import (
	"context"

	"github.com/smartcontractkit/libocr/offchainreporting2plus"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	oraclepb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/oracle"
)

var _ offchainreporting2plus.Oracle = (*client)(nil)

type client struct {
	grpc oraclepb.OracleClient
}

func NewClient(cc grpc.ClientConnInterface) *client {
	return &client{grpc: oraclepb.NewOracleClient(cc)}
}

func (c *client) Close() error {
	_, err := c.grpc.CloseOracle(context.Background(), &emptypb.Empty{})
	return err
}

func (c *client) Start() error {
	_, err := c.grpc.StartOracle(context.Background(), &emptypb.Empty{})
	return err
}
