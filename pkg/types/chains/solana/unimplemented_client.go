package solana

import (
	"context"
	"fmt"
)

// UnimplementedSolanaClient provides default stubs for Client methods.
// Embed this type to implement Client and receive default behavior for new RPC methods.
type UnimplementedSolanaClient struct{}

var _ Client = UnimplementedSolanaClient{}

func (UnimplementedSolanaClient) mustEmbedUnimplementedClient() {}

func (UnimplementedSolanaClient) GetBalance(context.Context, GetBalanceRequest) (*GetBalanceReply, error) {
	return nil, fmt.Errorf("method GetBalance not implemented")
}

func (UnimplementedSolanaClient) GetAccountInfoWithOpts(context.Context, GetAccountInfoRequest) (*GetAccountInfoReply, error) {
	return nil, fmt.Errorf("method GetAccountInfoWithOpts not implemented")
}

func (UnimplementedSolanaClient) GetMultipleAccountsWithOpts(context.Context, GetMultipleAccountsRequest) (*GetMultipleAccountsReply, error) {
	return nil, fmt.Errorf("method GetMultipleAccountsWithOpts not implemented")
}

func (UnimplementedSolanaClient) GetBlock(context.Context, GetBlockRequest) (*GetBlockReply, error) {
	return nil, fmt.Errorf("method GetBlock not implemented")
}

func (UnimplementedSolanaClient) GetSlotHeight(context.Context, GetSlotHeightRequest) (*GetSlotHeightReply, error) {
	return nil, fmt.Errorf("method GetSlotHeight not implemented")
}

func (UnimplementedSolanaClient) GetTransaction(context.Context, GetTransactionRequest) (*GetTransactionReply, error) {
	return nil, fmt.Errorf("method GetTransaction not implemented")
}

func (UnimplementedSolanaClient) GetFeeForMessage(context.Context, GetFeeForMessageRequest) (*GetFeeForMessageReply, error) {
	return nil, fmt.Errorf("method GetFeeForMessage not implemented")
}

func (UnimplementedSolanaClient) GetSignatureStatuses(context.Context, GetSignatureStatusesRequest) (*GetSignatureStatusesReply, error) {
	return nil, fmt.Errorf("method GetSignatureStatuses not implemented")
}

func (UnimplementedSolanaClient) SimulateTX(context.Context, SimulateTXRequest) (*SimulateTXReply, error) {
	return nil, fmt.Errorf("method SimulateTX not implemented")
}

func (UnimplementedSolanaClient) GetProgramAccounts(context.Context, GetProgramAccountsRequest) (*GetProgramAccountsReply, error) {
	return nil, fmt.Errorf("method GetProgramAccounts not implemented")
}
