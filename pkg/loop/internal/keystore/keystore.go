package keystore

import (
	"context"

	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/goplugin"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	keystorepb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/keystore"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/types/keystore"
)

var _ keystore.Keystore = (*Client)(nil)

type Client struct {
	services.Service
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

func (c *Client) ListKeys(ctx context.Context, tags []string) ([][]byte, error) {
	reply, err := c.grpc.ListKeys(ctx, &keystorepb.ListKeysRequest{
		Tags: tags,
	})

	if err != nil {
		return nil, err
	}
	return reply.KeyIDs, nil
}

func (c *Client) RunUDF(ctx context.Context, name string, keyID []byte, data []byte) ([]byte, error) {
	reply, err := c.grpc.RunUDF(ctx, &keystorepb.RunUDFRequest{
		Name:  name,
		KeyID: keyID,
		Data:  data,
	})

	if err != nil {
		return nil, err
	}
	return reply.Data, nil
}

func (c *Client) ImportKey(ctx context.Context, keyType string, data []byte, tags []string) ([]byte, error) {
	reply, err := c.grpc.ImportKey(ctx, &keystorepb.ImportKeyRequest{
		KeyType: keyType,
		Data:    data,
		Tags:    tags,
	})

	if err != nil {
		return nil, err
	}
	return reply.KeyID, nil
}

func (c *Client) ExportKey(ctx context.Context, keyID []byte) ([]byte, error) {
	reply, err := c.grpc.ExportKey(ctx, &keystorepb.ExportKeyRequest{
		KeyID: keyID,
	})

	if err != nil {
		return nil, err
	}
	return reply.Data, nil
}

func (c *Client) CreateKey(ctx context.Context, keyType string, tags []string) ([]byte, error) {
	reply, err := c.grpc.CreateKey(ctx, &keystorepb.CreateKeyRequest{
		KeyType: keyType,
		Tags:    tags,
	})

	if err != nil {
		return nil, err
	}
	return reply.KeyID, nil
}

func (c *Client) DeleteKey(ctx context.Context, keyID []byte) error {
	_, err := c.grpc.DeleteKey(ctx, &keystorepb.DeleteKeyRequest{
		KeyID: keyID,
	})

	if err != nil {
		return err
	}
	return nil
}

func (c *Client) AddTag(ctx context.Context, keyID []byte, tag string) error {
	_, err := c.grpc.AddTag(ctx, &keystorepb.AddTagRequest{
		KeyID: keyID,
		Tag:   tag,
	})

	if err != nil {
		return err
	}
	return nil
}

func (c *Client) RemoveTag(ctx context.Context, keyID []byte, tag string) error {
	_, err := c.grpc.RemoveTag(ctx, &keystorepb.RemoveTagRequest{
		KeyID: keyID,
		Tag:   tag,
	})

	if err != nil {
		return err
	}
	return nil
}

func (c *Client) ListTags(ctx context.Context, keyID []byte) ([]string, error) {
	reply, err := c.grpc.ListTags(ctx, &keystorepb.ListTagsRequest{
		KeyID: keyID,
	})

	if err != nil {
		return nil, err
	}
	return reply.Tags, nil
}

var _ keystorepb.KeystoreServer = (*server)(nil)

type server struct {
	*net.BrokerExt
	keystorepb.UnimplementedKeystoreServer

	impl GRPCService
}

func RegisterKeystoreServer(server *grpc.Server, broker net.Broker, brokerCfg net.BrokerConfig, impl GRPCService) error {
	keystorepb.RegisterKeystoreServer(server, newKeystoreServer(broker, brokerCfg, impl))
	return nil
}

func newKeystoreServer(broker net.Broker, brokerCfg net.BrokerConfig, impl GRPCService) *server {
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

func (s *server) ListKeys(ctx context.Context, request *keystorepb.ListKeysRequest) (*keystorepb.ListKeysResponse, error) {
	keyIDs, err := s.impl.ListKeys(ctx, request.Tags)
	if err != nil {
		return nil, err
	}
	return &keystorepb.ListKeysResponse{KeyIDs: keyIDs}, err
}

func (s *server) RunUDF(ctx context.Context, request *keystorepb.RunUDFRequest) (*keystorepb.RunUDFResponse, error) {
	data, err := s.impl.RunUDF(ctx, request.Name, request.KeyID, request.Data)
	if err != nil {
		return nil, err
	}
	return &keystorepb.RunUDFResponse{Data: data}, err
}

func (s *server) ImportKey(ctx context.Context, request *keystorepb.ImportKeyRequest) (*keystorepb.ImportKeyResponse, error) {
	keyIDs, err := s.impl.ImportKey(ctx, request.KeyType, request.Data, request.Tags)
	if err != nil {
		return nil, err
	}
	return &keystorepb.ImportKeyResponse{KeyID: keyIDs}, err
}

func (s *server) ExportKey(ctx context.Context, request *keystorepb.ExportKeyRequest) (*keystorepb.ExportKeyResponse, error) {
	data, err := s.impl.ExportKey(ctx, request.KeyID)
	if err != nil {
		return nil, err
	}
	return &keystorepb.ExportKeyResponse{Data: data}, err
}

func (s *server) CreateKey(ctx context.Context, request *keystorepb.CreateKeyRequest) (*keystorepb.CreateKeyResponse, error) {
	keyIDs, err := s.impl.CreateKey(ctx, request.KeyType, request.Tags)
	if err != nil {
		return nil, err
	}
	return &keystorepb.CreateKeyResponse{KeyID: keyIDs}, err
}

func (s *server) DeleteKey(ctx context.Context, request *keystorepb.DeleteKeyRequest) (*keystorepb.DeleteKeyResponse, error) {
	err := s.impl.DeleteKey(ctx, request.KeyID)
	if err != nil {
		return nil, err
	}
	return &keystorepb.DeleteKeyResponse{}, err
}

func (s *server) AddTag(ctx context.Context, request *keystorepb.AddTagRequest) (*keystorepb.AddTagResponse, error) {
	err := s.impl.AddTag(ctx, request.KeyID, request.Tag)
	if err != nil {
		return nil, err
	}
	return &keystorepb.AddTagResponse{}, err
}

func (s *server) RemoveTag(ctx context.Context, request *keystorepb.RemoveTagRequest) (*keystorepb.RemoveTagResponse, error) {
	err := s.impl.RemoveTag(ctx, request.KeyID, request.Tag)
	if err != nil {
		return nil, err
	}
	return &keystorepb.RemoveTagResponse{}, err
}

func (s *server) ListTags(ctx context.Context, request *keystorepb.ListTagsRequest) (*keystorepb.ListTagsResponse, error) {
	tags, err := s.impl.ListTags(ctx, request.KeyID)
	if err != nil {
		return nil, err
	}
	return &keystorepb.ListTagsResponse{Tags: tags}, nil
}
