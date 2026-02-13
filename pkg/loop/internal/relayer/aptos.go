package relayer

import (
	"context"
	"fmt"

	aptospb "github.com/smartcontractkit/chainlink-common/pkg/chains/aptos"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/chains/aptos"
)

var _ types.AptosService = (*AptosClient)(nil)

type AptosClient struct {
	grpcClient aptospb.AptosClient
}

func NewAptosClient(client aptospb.AptosClient) *AptosClient {
	return &AptosClient{
		grpcClient: client,
	}
}

func (ac *AptosClient) AccountAPTBalance(ctx context.Context, req aptos.AccountAPTBalanceRequest) (*aptos.AccountAPTBalanceReply, error) {
	reply, err := ac.grpcClient.AccountAPTBalance(ctx, &aptospb.AccountAPTBalanceRequest{
		Address: req.Address[:],
	})
	if err != nil {
		return nil, err
	}
	return &aptos.AccountAPTBalanceReply{
		Value: reply.Value,
	}, nil
}

func (ac *AptosClient) View(ctx context.Context, req aptos.ViewRequest) (*aptos.ViewReply, error) {
	// Convert Go types to proto types
	protoPayload, err := aptospb.ConvertViewPayloadToProto(req.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to convert view payload: %w", err)
	}

	protoReq := &aptospb.ViewRequest{
		Payload: protoPayload,
	}

	reply, err := ac.grpcClient.View(ctx, protoReq)
	if err != nil {
		return nil, err
	}

	// Convert proto types back to Go types
	return aptospb.ConvertViewReplyFromProto(reply)
}

func (ac *AptosClient) EventsByHandle(ctx context.Context, req aptos.EventsByHandleRequest) (*aptos.EventsByHandleReply, error) {
	// Convert Go types to proto types
	protoReq, err := aptospb.ConvertEventsByHandleRequestToProto(&req)
	if err != nil {
		return nil, fmt.Errorf("failed to convert request: %w", err)
	}

	reply, err := ac.grpcClient.EventsByHandle(ctx, protoReq)
	if err != nil {
		return nil, err
	}

	// Convert proto types back to Go types
	goReply, err := aptospb.ConvertEventsByHandleReplyFromProto(reply)
	if err != nil {
		return nil, fmt.Errorf("failed to convert reply: %w", err)
	}

	return goReply, nil
}

func (ac *AptosClient) TransactionByHash(ctx context.Context, req aptos.TransactionByHashRequest) (*aptos.TransactionByHashReply, error) {
	protoReq := aptospb.ConvertTransactionByHashRequestToProto(req)
	protoResp, err := ac.grpcClient.TransactionByHash(ctx, protoReq)
	if err != nil {
		return nil, err
	}
	return aptospb.ConvertTransactionByHashReplyFromProto(protoResp)
}

func (ac *AptosClient) SubmitTransaction(ctx context.Context, req aptos.SubmitTransactionRequest) (*aptos.SubmitTransactionReply, error) {
	protoReq, err := aptospb.ConvertSubmitTransactionRequestToProto(req)
	if err != nil {
		return nil, fmt.Errorf("failed to convert request: %w", err)
	}

	protoResp, err := ac.grpcClient.SubmitTransaction(ctx, protoReq)
	if err != nil {
		return nil, err
	}

	return aptospb.ConvertSubmitTransactionReplyFromProto(protoResp)
}

type aptosServer struct {
	aptospb.UnimplementedAptosServer

	*net.BrokerExt

	impl types.AptosService
}

var _ aptospb.AptosServer = (*aptosServer)(nil)

func newAptosServer(impl types.AptosService, b *net.BrokerExt) *aptosServer {
	return &aptosServer{impl: impl, BrokerExt: b.WithName("AptosServer")}
}

func (s *aptosServer) AccountAPTBalance(ctx context.Context, req *aptospb.AccountAPTBalanceRequest) (*aptospb.AccountAPTBalanceReply, error) {
	reply, err := s.impl.AccountAPTBalance(ctx, aptos.AccountAPTBalanceRequest{
		Address: aptos.AccountAddress(req.Address),
	})
	if err != nil {
		return nil, err
	}
	return &aptospb.AccountAPTBalanceReply{
		Value: reply.Value,
	}, nil
}

func (s *aptosServer) View(ctx context.Context, req *aptospb.ViewRequest) (*aptospb.ViewReply, error) {
	// Convert proto types to Go types
	goPayload, err := aptospb.ConvertViewPayloadFromProto(req.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to convert proto payload: %w", err)
	}

	goReq := aptos.ViewRequest{
		Payload: goPayload,
	}

	reply, err := s.impl.View(ctx, goReq)
	if err != nil {
		return nil, err
	}

	// Convert Go types back to proto types
	return aptospb.ConvertViewReplyToProto(reply)
}

func (s *aptosServer) EventsByHandle(ctx context.Context, req *aptospb.EventsByHandleRequest) (*aptospb.EventsByHandleReply, error) {
	// Convert proto types to Go types
	goReq, err := aptospb.ConvertEventsByHandleRequestFromProto(req)
	if err != nil {
		return nil, fmt.Errorf("failed to convert request: %w", err)
	}

	reply, err := s.impl.EventsByHandle(ctx, *goReq)
	if err != nil {
		return nil, err
	}

	// Convert Go types back to proto types
	protoReply, err := aptospb.ConvertEventsByHandleReplyToProto(reply)
	if err != nil {
		return nil, fmt.Errorf("failed to convert reply: %w", err)
	}

	return protoReply, nil
}

func (s *aptosServer) TransactionByHash(ctx context.Context, req *aptospb.TransactionByHashRequest) (*aptospb.TransactionByHashReply, error) {
	// Convert proto to Go types
	goReq := aptospb.ConvertTransactionByHashRequestFromProto(req)

	reply, err := s.impl.TransactionByHash(ctx, goReq)
	if err != nil {
		return nil, err
	}

	// Convert Go types back to proto types
	return aptospb.ConvertTransactionByHashReplyToProto(reply), nil
}

func (s *aptosServer) SubmitTransaction(ctx context.Context, req *aptospb.SubmitTransactionRequest) (*aptospb.SubmitTransactionReply, error) {
	goReq, err := aptospb.ConvertSubmitTransactionRequestFromProto(req)
	if err != nil {
		return nil, fmt.Errorf("failed to convert request: %w", err)
	}

	reply, err := s.impl.SubmitTransaction(ctx, *goReq)
	if err != nil {
		return nil, err
	}

	protoReply, err := aptospb.ConvertSubmitTransactionReplyToProto(reply)
	if err != nil {
		return nil, fmt.Errorf("failed to convert reply: %w", err)
	}

	return protoReply, nil
}
