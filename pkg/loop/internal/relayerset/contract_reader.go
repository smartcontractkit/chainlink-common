package relayerset

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/relayerset"

	"google.golang.org/protobuf/types/known/emptypb"
)

type contractReader struct {
	contractReaderID string
	client           *Client
}

func (c *contractReader) GetLatestValue(ctx context.Context, in *pb.GetLatestValueRequest, opts ...grpc.CallOption) (*pb.GetLatestValueReply, error) {
	return c.client.relayerSetClient.ContractReaderGetLatestValue(ctx, &relayerset.ContractReaderGetLatestValueRequest{
		ContractReaderId: c.contractReaderID,
		Request:          in,
	}, opts...)
}

func (c *contractReader) GetLatestValueWithHeadData(ctx context.Context, in *pb.GetLatestValueRequest, opts ...grpc.CallOption) (*pb.GetLatestValueWithHeadDataReply, error) {
	return c.client.relayerSetClient.ContractReaderGetLatestValueWithHeadData(ctx, &relayerset.ContractReaderGetLatestValueRequest{
		ContractReaderId: c.contractReaderID,
		Request:          in,
	}, opts...)
}

func (c *contractReader) BatchGetLatestValues(ctx context.Context, in *pb.BatchGetLatestValuesRequest, opts ...grpc.CallOption) (*pb.BatchGetLatestValuesReply, error) {
	return c.client.relayerSetClient.ContractReaderBatchGetLatestValues(ctx, &relayerset.ContractReaderBatchGetLatestValuesRequest{
		ContractReaderId: c.contractReaderID,
		Request:          in,
	}, opts...)
}

func (c *contractReader) QueryKey(ctx context.Context, in *pb.QueryKeyRequest, opts ...grpc.CallOption) (*pb.QueryKeyReply, error) {
	return c.client.relayerSetClient.ContractReaderQueryKey(ctx, &relayerset.ContractReaderQueryKeyRequest{
		ContractReaderId: c.contractReaderID,
		Request:          in,
	}, opts...)
}

func (c *contractReader) QueryKeys(ctx context.Context, in *pb.QueryKeysRequest, opts ...grpc.CallOption) (*pb.QueryKeysReply, error) {
	return c.client.relayerSetClient.ContractReaderQueryKeys(ctx, &relayerset.ContractReaderQueryKeysRequest{
		ContractReaderId: c.contractReaderID,
		Request:          in,
	}, opts...)
}

func (c *contractReader) Bind(ctx context.Context, in *pb.BindRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return c.client.relayerSetClient.ContractReaderBind(ctx, &relayerset.ContractReaderBindRequest{
		ContractReaderId: c.contractReaderID,
		Request:          in,
	}, opts...)
}

func (c *contractReader) Unbind(ctx context.Context, in *pb.UnbindRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return c.client.relayerSetClient.ContractReaderUnbind(ctx, &relayerset.ContractReaderUnbindRequest{
		ContractReaderId: c.contractReaderID,
		Request:          in,
	}, opts...)
}

type contractReaderServiceClient struct {
	contractReaderID string
	client           *Client
}

func (s *contractReaderServiceClient) ClientConn() grpc.ClientConnInterface {
	return s.client.ClientConn()
}

func (s *contractReaderServiceClient) Start(ctx context.Context) error {
	_, err := s.client.relayerSetClient.ContractReaderStart(ctx, &relayerset.ContractReaderStartRequest{
		ContractReaderId: s.contractReaderID,
	})
	if err != nil {
		return fmt.Errorf("error starting contract reader: %w", err)
	}
	return nil
}

func (s *contractReaderServiceClient) Close() error {
	_, err := s.client.relayerSetClient.ContractReaderClose(context.Background(), &relayerset.ContractReaderCloseRequest{
		ContractReaderId: s.contractReaderID,
	})
	if err != nil {
		return fmt.Errorf("error closing contract reader: %w", err)
	}
	return nil
}

func (s *contractReaderServiceClient) HealthReport() map[string]error {
	return map[string]error{}
}
func (s *contractReaderServiceClient) Name() string {
	return "RelayerSetContractReader"
}

func (s *contractReaderServiceClient) Ready() error {
	return nil
}
