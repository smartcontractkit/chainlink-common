package solana

import (
	"context"
	"errors"
)

// UnimplementedSolanaClient provides default stubs for Client methods.
// Embed this type to implement Client and receive default behavior for new RPC methods.
type UnimplementedSolanaClient struct{}

var _ Client = UnimplementedSolanaClient{}

func (UnimplementedSolanaClient) mustEmbedUnimplementedClient() {}

// ClientMustEmbed satisfies Client.mustEmbedUnimplementedClient without stub RPC methods.
// Embed alongside another Client implementation (e.g. a testify mock) when the mock cannot live in this package.
type ClientMustEmbed struct{}

func (ClientMustEmbed) mustEmbedUnimplementedClient() {}

func (UnimplementedSolanaClient) GetBalance(context.Context, GetBalanceRequest) (*GetBalanceReply, error) {
	return nil, errors.New("method GetBalance not implemented")
}

func (UnimplementedSolanaClient) GetAccountInfoWithOpts(context.Context, GetAccountInfoRequest) (*GetAccountInfoReply, error) {
	return nil, errors.New("method GetAccountInfoWithOpts not implemented")
}

func (UnimplementedSolanaClient) GetMultipleAccountsWithOpts(context.Context, GetMultipleAccountsRequest) (*GetMultipleAccountsReply, error) {
	return nil, errors.New("method GetMultipleAccountsWithOpts not implemented")
}

func (UnimplementedSolanaClient) GetBlock(context.Context, GetBlockRequest) (*GetBlockReply, error) {
	return nil, errors.New("method GetBlock not implemented")
}

func (UnimplementedSolanaClient) GetSlotHeight(context.Context, GetSlotHeightRequest) (*GetSlotHeightReply, error) {
	return nil, errors.New("method GetSlotHeight not implemented")
}

func (UnimplementedSolanaClient) GetTransaction(context.Context, GetTransactionRequest) (*GetTransactionReply, error) {
	return nil, errors.New("method GetTransaction not implemented")
}

func (UnimplementedSolanaClient) GetFeeForMessage(context.Context, GetFeeForMessageRequest) (*GetFeeForMessageReply, error) {
	return nil, errors.New("method GetFeeForMessage not implemented")
}

func (UnimplementedSolanaClient) GetSignatureStatuses(context.Context, GetSignatureStatusesRequest) (*GetSignatureStatusesReply, error) {
	return nil, errors.New("method GetSignatureStatuses not implemented")
}

func (UnimplementedSolanaClient) SimulateTX(context.Context, SimulateTXRequest) (*SimulateTXReply, error) {
	return nil, errors.New("method SimulateTX not implemented")
}

func (UnimplementedSolanaClient) GetProgramAccounts(context.Context, GetProgramAccountsRequest) (*GetProgramAccountsReply, error) {
	return nil, errors.New("method GetProgramAccounts not implemented")
}
