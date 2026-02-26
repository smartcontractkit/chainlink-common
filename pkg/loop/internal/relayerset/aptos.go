package relayerset

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	aptospb "github.com/smartcontractkit/chainlink-common/pkg/chains/aptos"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/relayerset"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/chains/aptos"
)

// aptosClient wraps the AptosRelayerSetClient by attaching a RelayerID to AptosClient requests.
// The attached RelayerID is stored in the context metadata.
type aptosClient struct {
	relayID types.RelayID
	client  aptospb.AptosClient
}

var _ aptospb.AptosClient = (*aptosClient)(nil)

func (ac *aptosClient) AccountAPTBalance(ctx context.Context, in *aptospb.AccountAPTBalanceRequest, opts ...grpc.CallOption) (*aptospb.AccountAPTBalanceReply, error) {
	return ac.client.AccountAPTBalance(appendRelayID(ctx, ac.relayID), in, opts...)
}

func (ac *aptosClient) View(ctx context.Context, in *aptospb.ViewRequest, opts ...grpc.CallOption) (*aptospb.ViewReply, error) {
	return ac.client.View(appendRelayID(ctx, ac.relayID), in, opts...)
}

func (ac *aptosClient) TransactionByHash(ctx context.Context, in *aptospb.TransactionByHashRequest, opts ...grpc.CallOption) (*aptospb.TransactionByHashReply, error) {
	return ac.client.TransactionByHash(appendRelayID(ctx, ac.relayID), in, opts...)
}

func (ac *aptosClient) AccountTransactions(ctx context.Context, in *aptospb.AccountTransactionsRequest, opts ...grpc.CallOption) (*aptospb.AccountTransactionsReply, error) {
	return ac.client.AccountTransactions(appendRelayID(ctx, ac.relayID), in, opts...)
}

func (ac *aptosClient) SubmitTransaction(ctx context.Context, in *aptospb.SubmitTransactionRequest, opts ...grpc.CallOption) (*aptospb.SubmitTransactionReply, error) {
	return ac.client.SubmitTransaction(appendRelayID(ctx, ac.relayID), in, opts...)
}

type aptosServer struct {
	aptospb.UnimplementedAptosServer
	parent *Server
}

var _ aptospb.AptosServer = (*aptosServer)(nil)

func (as *aptosServer) AccountAPTBalance(ctx context.Context, req *aptospb.AccountAPTBalanceRequest) (*aptospb.AccountAPTBalanceReply, error) {
	aptosService, err := as.parent.getAptosService(ctx)
	if err != nil {
		return nil, err
	}

	reply, err := aptosService.AccountAPTBalance(ctx, aptos.AccountAPTBalanceRequest{
		Address: aptos.AccountAddress(req.Address),
	})
	if err != nil {
		return nil, err
	}
	return &aptospb.AccountAPTBalanceReply{
		Value: reply.Value,
	}, nil
}

func (as *aptosServer) View(ctx context.Context, req *aptospb.ViewRequest) (*aptospb.ViewReply, error) {
	aptosService, err := as.parent.getAptosService(ctx)
	if err != nil {
		return nil, err
	}

	// Convert proto types to Go types
	goPayload, err := aptospb.ConvertViewPayloadFromProto(req.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to convert proto payload: %w", err)
	}

	goReq := aptos.ViewRequest{
		Payload: goPayload,
	}

	reply, err := aptosService.View(ctx, goReq)
	if err != nil {
		return nil, err
	}

	// Convert Go types back to proto types
	return aptospb.ConvertViewReplyToProto(reply)
}

func (as *aptosServer) TransactionByHash(ctx context.Context, req *aptospb.TransactionByHashRequest) (*aptospb.TransactionByHashReply, error) {
	aptosService, err := as.parent.getAptosService(ctx)
	if err != nil {
		return nil, err
	}

	// Convert proto to Go types
	goReq := aptospb.ConvertTransactionByHashRequestFromProto(req)

	reply, err := aptosService.TransactionByHash(ctx, goReq)
	if err != nil {
		return nil, err
	}

	// Convert Go types back to proto types
	return aptospb.ConvertTransactionByHashReplyToProto(reply), nil
}

func (as *aptosServer) AccountTransactions(ctx context.Context, req *aptospb.AccountTransactionsRequest) (*aptospb.AccountTransactionsReply, error) {
	aptosService, err := as.parent.getAptosService(ctx)
	if err != nil {
		return nil, err
	}

	goReq, err := aptospb.ConvertAccountTransactionsRequestFromProto(req)
	if err != nil {
		return nil, fmt.Errorf("failed to convert request: %w", err)
	}

	reply, err := aptosService.AccountTransactions(ctx, goReq)
	if err != nil {
		return nil, err
	}

	return aptospb.ConvertAccountTransactionsReplyToProto(reply), nil
}

func (as *aptosServer) SubmitTransaction(ctx context.Context, req *aptospb.SubmitTransactionRequest) (*aptospb.SubmitTransactionReply, error) {
	aptosService, err := as.parent.getAptosService(ctx)
	if err != nil {
		return nil, err
	}

	goReq, err := aptospb.ConvertSubmitTransactionRequestFromProto(req)
	if err != nil {
		return nil, fmt.Errorf("failed to convert request: %w", err)
	}

	reply, err := aptosService.SubmitTransaction(ctx, *goReq)
	if err != nil {
		return nil, err
	}

	protoReply, err := aptospb.ConvertSubmitTransactionReplyToProto(reply)
	if err != nil {
		return nil, fmt.Errorf("failed to convert reply: %w", err)
	}

	return protoReply, nil
}

func (s *Server) getAptosService(ctx context.Context) (types.AptosService, error) {
	id, err := readRelayID(ctx)
	if err != nil {
		return nil, err
	}
	idT := relayerset.RelayerId{Network: id.Network, ChainId: id.ChainID}
	r, err := s.getRelayer(ctx, &idT)
	if err != nil {
		return nil, err
	}

	return r.Aptos()
}
