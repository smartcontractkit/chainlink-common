package relayerset

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
)

const metadataContractReader = "contractReaderID"

// contractReader wraps the ContractReaderClient by attaching a contractReaderId to its requests.
// The attached contractReaderId is stored in the context metadata.
type contractReader struct {
	contractReaderID string
	client           *Client
}

func (c *contractReader) GetLatestValue(ctx context.Context, in *pb.GetLatestValueRequest, opts ...grpc.CallOption) (*pb.GetLatestValueReply, error) {
	return c.client.contractReaderClient.GetLatestValue(appendContractReaderID(ctx, c.contractReaderID), in, opts...)
}

func (c *contractReader) GetLatestValueWithHeadData(ctx context.Context, in *pb.GetLatestValueRequest, opts ...grpc.CallOption) (*pb.GetLatestValueWithHeadDataReply, error) {
	return c.client.contractReaderClient.GetLatestValueWithHeadData(appendContractReaderID(ctx, c.contractReaderID), in, opts...)
}

func (c *contractReader) BatchGetLatestValues(ctx context.Context, in *pb.BatchGetLatestValuesRequest, opts ...grpc.CallOption) (*pb.BatchGetLatestValuesReply, error) {
	return c.client.contractReaderClient.BatchGetLatestValues(appendContractReaderID(ctx, c.contractReaderID), in, opts...)
}

func (c *contractReader) QueryKey(ctx context.Context, in *pb.QueryKeyRequest, opts ...grpc.CallOption) (*pb.QueryKeyReply, error) {
	return c.client.contractReaderClient.QueryKey(appendContractReaderID(ctx, c.contractReaderID), in, opts...)
}

func (c *contractReader) QueryKeys(ctx context.Context, in *pb.QueryKeysRequest, opts ...grpc.CallOption) (*pb.QueryKeysReply, error) {
	return c.client.contractReaderClient.QueryKeys(appendContractReaderID(ctx, c.contractReaderID), in, opts...)
}

func (c *contractReader) Bind(ctx context.Context, in *pb.BindRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return c.client.contractReaderClient.Bind(appendContractReaderID(ctx, c.contractReaderID), in, opts...)
}

func (c *contractReader) Unbind(ctx context.Context, in *pb.UnbindRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return c.client.contractReaderClient.Unbind(appendContractReaderID(ctx, c.contractReaderID), in, opts...)
}

func appendContractReaderID(ctx context.Context, id string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, metadataContractReader, id)
}

// contractReaderServiceClient wraps the RelayerSetClient ContractReader{Close/Start} methods by attaching a contractReaderId to its requests.
// The attached contractReaderId is stored in the context metadata.
type contractReaderServiceClient struct {
	contractReaderID string
	client           *Client
}

func (s *contractReaderServiceClient) ClientConn() grpc.ClientConnInterface {
	return s.client.ClientConn()
}

func (s *contractReaderServiceClient) Start(ctx context.Context) error {
	_, err := s.client.relayerSetClient.ContractReaderStart(appendContractReaderID(ctx, s.contractReaderID), nil)
	return err
}

func (s *contractReaderServiceClient) Close() error {
	_, err := s.client.relayerSetClient.ContractReaderClose(appendContractReaderID(context.Background(), s.contractReaderID), &emptypb.Empty{})
	return err
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

func (s *Server) ContractReaderStart(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	reader, err := s.getReader(ctx)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, reader.reader.Start(ctx)
}

func (s *Server) ContractReaderClose(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	reader, err := s.getReader(ctx)
	if err != nil {
		return nil, err
	}
	id, err := readContextValue(ctx, metadataContractReader)
	if err != nil {
		return nil, err
	}
	s.removeReader(id)
	return &emptypb.Empty{}, reader.reader.Close()
}

func (s *Server) GetLatestValue(ctx context.Context, in *pb.GetLatestValueRequest) (*pb.GetLatestValueReply, error) {
	reader, err := s.getReader(ctx)
	if err != nil {
		return nil, err
	}

	return reader.server.GetLatestValue(ctx, in)
}

func (s *Server) GetLatestValueWithHeadData(ctx context.Context, in *pb.GetLatestValueRequest) (*pb.GetLatestValueWithHeadDataReply, error) {
	reader, err := s.getReader(ctx)
	if err != nil {
		return nil, err
	}

	return reader.server.GetLatestValueWithHeadData(ctx, in)
}

func (s *Server) GetLatestValues(ctx context.Context, in *pb.BatchGetLatestValuesRequest) (*pb.BatchGetLatestValuesReply, error) {
	reader, err := s.getReader(ctx)
	if err != nil {
		return nil, err
	}

	return reader.server.BatchGetLatestValues(ctx, in)
}

func (s *Server) QueryKeys(ctx context.Context, in *pb.QueryKeysRequest) (*pb.QueryKeysReply, error) {
	reader, err := s.getReader(ctx)
	if err != nil {
		return nil, err
	}

	return reader.server.QueryKeys(ctx, in)
}

func (s *Server) QueryKey(ctx context.Context, in *pb.QueryKeyRequest) (*pb.QueryKeyReply, error) {
	reader, err := s.getReader(ctx)
	if err != nil {
		return nil, err
	}

	return reader.server.QueryKey(ctx, in)
}

func (s *Server) Bind(ctx context.Context, in *pb.BindRequest) (*emptypb.Empty, error) {
	reader, err := s.getReader(ctx)
	if err != nil {
		return nil, err
	}

	return reader.server.Bind(ctx, in)
}

func (s *Server) Unbind(ctx context.Context, in *pb.UnbindRequest) (*emptypb.Empty, error) {
	reader, err := s.getReader(ctx)
	if err != nil {
		return nil, err
	}

	return reader.server.Unbind(ctx, in)
}
