package relayerset

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/relayerset"
)

const contractHead = "contractReaderID"

type contractReader struct {
	contractReaderID string
	client           *Client
}

func (c *contractReader) GetLatestValue(ctx context.Context, in *pb.GetLatestValueRequest, opts ...grpc.CallOption) (*pb.GetLatestValueReply, error) {
	return c.client.contractReaderClient.GetLatestValue(metadata.AppendToOutgoingContext(ctx, contractHead, c.contractReaderID), in, opts...)
}

func (c *contractReader) GetLatestValueWithHeadData(ctx context.Context, in *pb.GetLatestValueRequest, opts ...grpc.CallOption) (*pb.GetLatestValueWithHeadDataReply, error) {
	return c.client.contractReaderClient.GetLatestValueWithHeadData(metadata.AppendToOutgoingContext(ctx, contractHead, c.contractReaderID), in, opts...)
}

func (c *contractReader) BatchGetLatestValues(ctx context.Context, in *pb.BatchGetLatestValuesRequest, opts ...grpc.CallOption) (*pb.BatchGetLatestValuesReply, error) {
	return c.client.contractReaderClient.BatchGetLatestValues(metadata.AppendToOutgoingContext(ctx, contractHead, c.contractReaderID), in, opts...)
}

func (c *contractReader) QueryKey(ctx context.Context, in *pb.QueryKeyRequest, opts ...grpc.CallOption) (*pb.QueryKeyReply, error) {
	return c.client.contractReaderClient.QueryKey(metadata.AppendToOutgoingContext(ctx, contractHead, c.contractReaderID), in, opts...)
}

func (c *contractReader) QueryKeys(ctx context.Context, in *pb.QueryKeysRequest, opts ...grpc.CallOption) (*pb.QueryKeysReply, error) {
	return c.client.contractReaderClient.QueryKeys(metadata.AppendToOutgoingContext(ctx, contractHead, c.contractReaderID), in, opts...)
}

func (c *contractReader) Bind(ctx context.Context, in *pb.BindRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return c.client.contractReaderClient.Bind(metadata.AppendToOutgoingContext(ctx, contractHead, c.contractReaderID), in, opts...)
}

func (c *contractReader) Unbind(ctx context.Context, in *pb.UnbindRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return c.client.contractReaderClient.Unbind(metadata.AppendToOutgoingContext(ctx, contractHead, c.contractReaderID), in, opts...)
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

func (s *Server) ContractReaderStart(ctx context.Context, req *relayerset.ContractReaderStartRequest) (*emptypb.Empty, error) {
	reader, err := s.getReader(req.ContractReaderId)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, reader.reader.Start(ctx)
}

func (s *Server) ContractReaderClose(_ context.Context, req *relayerset.ContractReaderCloseRequest) (*emptypb.Empty, error) {
	reader, err := s.getReader(req.ContractReaderId)
	if err != nil {
		return nil, err
	}

	s.removeReader(req.ContractReaderId)
	return &emptypb.Empty{}, reader.reader.Close()
}

func (s *Server) GetLatestValue(ctx context.Context, in *pb.GetLatestValueRequest) (*pb.GetLatestValueReply, error) {
	readerId, err := getContractReaderId(ctx)
	if err != nil {
		return nil, err
	}
	reader, err := s.getReader(readerId)
	if err != nil {
		return nil, err
	}

	return reader.server.GetLatestValue(ctx, in)
}

func (s *Server) GetLatestValueWithHeadData(ctx context.Context, in *pb.GetLatestValueRequest) (*pb.GetLatestValueWithHeadDataReply, error) {
	readerId, err := getContractReaderId(ctx)
	if err != nil {
		return nil, err
	}
	reader, err := s.getReader(readerId)
	if err != nil {
		return nil, err
	}

	return reader.server.GetLatestValueWithHeadData(ctx, in)
}

func (s *Server) GetLatestValues(ctx context.Context, in *pb.BatchGetLatestValuesRequest) (*pb.BatchGetLatestValuesReply, error) {
	readerId, err := getContractReaderId(ctx)
	if err != nil {
		return nil, err
	}
	reader, err := s.getReader(readerId)
	if err != nil {
		return nil, err
	}

	return reader.server.BatchGetLatestValues(ctx, in)
}

func (s *Server) QueryKeys(ctx context.Context, in *pb.QueryKeysRequest) (*pb.QueryKeysReply, error) {
	readerId, err := getContractReaderId(ctx)
	if err != nil {
		return nil, err
	}
	reader, err := s.getReader(readerId)
	if err != nil {
		return nil, err
	}

	return reader.server.QueryKeys(ctx, in)
}

func (s *Server) QueryKey(ctx context.Context, in *pb.QueryKeyRequest) (*pb.QueryKeyReply, error) {
	readerId, err := getContractReaderId(ctx)
	if err != nil {
		return nil, err
	}
	reader, err := s.getReader(readerId)
	if err != nil {
		return nil, err
	}

	return reader.server.QueryKey(ctx, in)
}

func (s *Server) Bind(ctx context.Context, in *pb.BindRequest) (*emptypb.Empty, error) {
	readerId, err := getContractReaderId(ctx)
	if err != nil {
		return nil, err
	}
	reader, err := s.getReader(readerId)
	if err != nil {
		return nil, err
	}

	return reader.server.Bind(ctx, in)
}

func (s *Server) Unbind(ctx context.Context, in *pb.UnbindRequest) (*emptypb.Empty, error) {
	readerId, err := getContractReaderId(ctx)
	if err != nil {
		return nil, err
	}
	reader, err := s.getReader(readerId)
	if err != nil {
		return nil, err
	}

	return reader.server.Unbind(ctx, in)
}

func getContractReaderId(ctx context.Context) (string, error) {
	return readValue(ctx, contractHead)
}

func readValue(ctx context.Context, key string) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		contractReaderIds := md.Get(key)
		if len(contractReaderIds) == 1 {
			return contractReaderIds[0], nil
		}
		return "", fmt.Errorf("num values is not 1 but %d", len(contractReaderIds))
	}
	return "", errors.New("could not read ctx metadata")
}
