package contractwriter

import (
	"context"
	"fmt"
	"math/big"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	evmpb "github.com/smartcontractkit/chainlink-common/pkg/chains/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/goplugin"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	loopjson "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/json"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

var _ types.ContractWriter = (*Client)(nil)

type ClientOpt func(*Client)

type Client struct {
	*goplugin.ServiceClient
	grpc pb.ContractWriterClient
}

func NewClient(b *net.BrokerExt, cc grpc.ClientConnInterface, opts ...ClientOpt) *Client {
	client := &Client{
		ServiceClient: goplugin.NewServiceClient(b, cc),
		grpc:          pb.NewContractWriterClient(cc),
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

func (c *Client) SubmitTransaction(ctx context.Context, contractName, method string, params any, transactionID, toAddress string, meta *types.TxMeta, value *big.Int) error {
	// Encode params as JSON bytes with type hint
	encodedParams, paramsTypeHint, err := loopjson.MarshalWithHint(params)
	if err != nil {
		return err
	}

	req := pb.SubmitTransactionRequest{
		ContractName:   contractName,
		Method:         method,
		Params:         encodedParams,
		TransactionId:  transactionID,
		ToAddress:      toAddress,
		Meta:           TxMetaToProto(meta),
		Value:          pb.NewBigIntFromInt(value),
		ParamsTypeHint: paramsTypeHint,
	}

	_, err = c.grpc.SubmitTransaction(ctx, &req)
	if err != nil {
		return net.WrapRPCErr(err)
	}

	return nil
}

func (c *Client) GetTransactionStatus(ctx context.Context, transactionID string) (types.TransactionStatus, error) {
	reply, err := c.grpc.GetTransactionStatus(ctx, &evmpb.GetTransactionStatusRequest{TransactionId: transactionID})
	if err != nil {
		return types.Unknown, net.WrapRPCErr(err)
	}

	return types.TransactionStatus(reply.TransactionStatus), nil
}

func (c *Client) GetFeeComponents(ctx context.Context) (*types.ChainFeeComponents, error) {
	reply, err := c.grpc.GetFeeComponents(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, net.WrapRPCErr(err)
	}

	return &types.ChainFeeComponents{
		ExecutionFee:        reply.ExecutionFee.Int(),
		DataAvailabilityFee: reply.DataAvailabilityFee.Int(),
	}, nil
}

func (c *Client) GetEstimateFee(ctx context.Context, contract, method string, params any, toAddress string, meta *types.TxMeta, val *big.Int) (types.EstimateFee, error) {
	// Encode params as JSON bytes with type hint
	encodedParams, paramsTypeHint, err := loopjson.MarshalWithHint(params)
	if err != nil {
		return types.EstimateFee{}, err
	}

	req := &pb.GetEstimateFeeRequest{
		ContractName:   contract,
		Method:         method,
		Params:         encodedParams,
		ToAddress:      toAddress,
		Meta:           TxMetaToProto(meta),
		Value:          pb.NewBigIntFromInt(val),
		ParamsTypeHint: paramsTypeHint,
	}

	reply, err := c.grpc.GetEstimateFee(ctx, req)
	if err != nil {
		return types.EstimateFee{}, net.WrapRPCErr(err)
	}

	return types.EstimateFee{
		Fee:      reply.Fee.Int(),
		Decimals: reply.Decimals,
	}, nil
}

// Server.

var _ pb.ContractWriterServer = (*Server)(nil)

type ServerOpt func(*Server)

type Server struct {
	pb.UnimplementedContractWriterServer
	impl types.ContractWriter
}

func NewServer(impl types.ContractWriter, opts ...ServerOpt) pb.ContractWriterServer {
	server := &Server{
		impl: impl,
	}

	for _, opt := range opts {
		opt(server)
	}

	return server
}

func (s *Server) SubmitTransaction(ctx context.Context, req *pb.SubmitTransactionRequest) (*emptypb.Empty, error) {
	// Use type hint to properly unmarshal params
	var params any
	if req.ParamsTypeHint != "" {
		var err error
		params, err = loopjson.UnmarshalWithHint(req.Params, req.ParamsTypeHint)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal params with type hint: %w", err)
		}
	} else {
		// Fallback to generic unmarshal if no type hint provided
		if err := loopjson.UnmarshalJson(req.Params, &params); err != nil {
			return nil, err
		}
	}

	err := s.impl.SubmitTransaction(ctx, req.ContractName, req.Method, params, req.TransactionId, req.ToAddress, TxMetaFromProto(req.Meta), req.Value.Int())
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) GetTransactionStatus(ctx context.Context, req *evmpb.GetTransactionStatusRequest) (*evmpb.GetTransactionStatusReply, error) {
	status, err := s.impl.GetTransactionStatus(ctx, req.TransactionId)
	if err != nil {
		return nil, err
	}

	//nolint: gosec // G115
	return &evmpb.GetTransactionStatusReply{TransactionStatus: evmpb.TransactionStatus(status)}, nil
}

func (s *Server) GetFeeComponents(ctx context.Context, _ *emptypb.Empty) (*pb.GetFeeComponentsReply, error) {
	feeComponents, err := s.impl.GetFeeComponents(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.GetFeeComponentsReply{
		ExecutionFee:        pb.NewBigIntFromInt(feeComponents.ExecutionFee),
		DataAvailabilityFee: pb.NewBigIntFromInt(feeComponents.DataAvailabilityFee),
	}, nil
}

func (s *Server) GetEstimateFee(ctx context.Context, req *pb.GetEstimateFeeRequest) (*pb.GetEstimateFeeReply, error) {
	// Use type hint to properly unmarshal params
	var params any
	if req.ParamsTypeHint != "" {
		var err error
		params, err = loopjson.UnmarshalWithHint(req.Params, req.ParamsTypeHint)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal params with type hint: %w", err)
		}
	} else {
		// Fallback to generic unmarshal if no type hint provided
		if err := loopjson.UnmarshalJson(req.Params, &params); err != nil {
			return nil, err
		}
	}

	estimateFee, err := s.impl.GetEstimateFee(ctx, req.ContractName, req.Method, params, req.ToAddress, TxMetaFromProto(req.Meta), req.Value.Int())
	if err != nil {
		return nil, err
	}

	return &pb.GetEstimateFeeReply{
		Fee:      pb.NewBigIntFromInt(estimateFee.Fee),
		Decimals: estimateFee.Decimals,
	}, nil
}

func RegisterContractWriterService(s *grpc.Server, contractWriter types.ContractWriter) {
	pb.RegisterServiceServer(s, &goplugin.ServiceServer{Srv: contractWriter})
	pb.RegisterContractWriterServer(s, NewServer(contractWriter))
}
