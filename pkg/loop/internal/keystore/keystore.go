package keystore

import (
	"context"

	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/goplugin"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	keystorepb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/keystore"
	"github.com/smartcontractkit/chainlink-common/pkg/types/keystore"
)

var _ keystore.Keystore = (*Client)(nil)

type Client struct {
	*goplugin.PluginClient

	grpc keystorepb.KeystoreClient
}

func NewKeystoreClient(broker net.Broker, brokerCfg net.BrokerConfig, conn *grpc.ClientConn) *Client {
	brokerCfg.Logger = logger.Named(brokerCfg.Logger, "KeystoreClient")
	pc := goplugin.NewPluginClient(broker, brokerCfg, conn)
	return &Client{PluginClient: pc, grpc: keystorepb.NewKeystoreClient(pc)}
}

func (c *Client) Sign(ctx context.Context, keyID []byte, data []byte) ([]byte, error) {
	reply, err := c.grpc.Sign(ctx, &keystorepb.SignRequest{
		KeyID: keyID,
		Data:  data,
	})

	if err != nil {
		return nil, err
	}
	return reply.Data, nil
}

func (c *Client) SignBatch(ctx context.Context, keyID []byte, data [][]byte) ([][]byte, error) {
	reply, err := c.grpc.SignBatch(ctx, &keystorepb.SignBatchRequest{
		KeyID: keyID,
		Data:  data,
	})

	if err != nil {
		return nil, err
	}
	return reply.Data, nil
}

func (c *Client) Verify(ctx context.Context, keyID []byte, data []byte) (bool, error) {
	reply, err := c.grpc.Verify(ctx, &keystorepb.VerifyRequest{
		KeyID: keyID,
		Data:  data,
	})

	if err != nil {
		return false, err
	}
	return reply.Valid, nil
}

func (c *Client) VerifyBatch(ctx context.Context, keyID []byte, data [][]byte) ([]bool, error) {
	reply, err := c.grpc.VerifyBatch(ctx, &keystorepb.VerifyBatchRequest{
		KeyID: keyID,
		Data:  data,
	})

	if err != nil {
		return nil, err
	}
	return reply.Valid, nil
}

func (c *Client) Get(ctx context.Context, tags []string) ([][]byte, error) {
	reply, err := c.grpc.Get(ctx, &keystorepb.GetRequest{
		Tags: tags,
	})

	if err != nil {
		return nil, err
	}
	return reply.KeyIDs, nil
}

func (c *Client) RunUDF(ctx context.Context, udfName string, keyID []byte, data []byte) ([]byte, error) {
	reply, err := c.grpc.RunUDF(ctx, &keystorepb.RunUDFRequest{
		UdfName: udfName,
		KeyID:   keyID,
		Data:    data,
	})

	if err != nil {
		return nil, err
	}
	return reply.Data, nil
}

var _ keystorepb.KeystoreServer = (*server)(nil)

type server struct {
	*net.BrokerExt
	keystorepb.UnimplementedKeystoreServer

	impl keystore.Keystore
}

func RegisterKeystoreServer(server *grpc.Server, broker net.Broker, brokerCfg net.BrokerConfig, impl keystore.Keystore) error {
	keystorepb.RegisterKeystoreServer(server, newKeystoreServer(broker, brokerCfg, impl))
	return nil
}

func newKeystoreServer(broker net.Broker, brokerCfg net.BrokerConfig, impl keystore.Keystore) *server {
	brokerCfg.Logger = logger.Named(brokerCfg.Logger, "KeystoreServer")
	return &server{BrokerExt: &net.BrokerExt{Broker: broker, BrokerConfig: brokerCfg}, impl: impl}
}

func (s *server) Sign(ctx context.Context, request *keystorepb.SignRequest) (*keystorepb.SignResponse, error) {
	data, err := s.impl.Sign(ctx, request.KeyID, request.Data)
	if err != nil {
		return nil, err
	}
	return &keystorepb.SignResponse{Data: data}, err
}

func (s *server) SignBatch(ctx context.Context, request *keystorepb.SignBatchRequest) (*keystorepb.SignBatchResponse, error) {
	data, err := s.impl.SignBatch(ctx, request.KeyID, request.Data)
	if err != nil {
		return nil, err
	}
	return &keystorepb.SignBatchResponse{Data: data}, err
}

func (s *server) Verify(ctx context.Context, request *keystorepb.VerifyRequest) (*keystorepb.VerifyResponse, error) {
	valid, err := s.impl.Verify(ctx, request.KeyID, request.Data)
	if err != nil {
		return nil, err
	}
	return &keystorepb.VerifyResponse{Valid: valid}, err
}

func (s *server) VerifyBatch(ctx context.Context, request *keystorepb.VerifyBatchRequest) (*keystorepb.VerifyBatchResponse, error) {
	valid, err := s.impl.VerifyBatch(ctx, request.KeyID, request.Data)
	if err != nil {
		return nil, err
	}
	return &keystorepb.VerifyBatchResponse{Valid: valid}, err
}

func (s *server) Get(ctx context.Context, request *keystorepb.GetRequest) (*keystorepb.GetResponse, error) {
	keyIDs, err := s.impl.Get(ctx, request.Tags)
	if err != nil {
		return nil, err
	}
	return &keystorepb.GetResponse{KeyIDs: keyIDs}, err
}

func (s *server) RunUDF(ctx context.Context, request *keystorepb.RunUDFRequest) (*keystorepb.RunUDFResponse, error) {
	data, err := s.impl.RunUDF(ctx, request.UdfName, request.KeyID, request.Data)
	if err != nil {
		return nil, err
	}
	return &keystorepb.RunUDFResponse{Data: data}, err
}
