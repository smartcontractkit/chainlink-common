package types_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
	stellartypes "github.com/smartcontractkit/chainlink-common/pkg/types/chains/stellar"
)

func TestUnimplementedStellarService_SubmitTransaction(t *testing.T) {
	t.Parallel()

	var svc types.UnimplementedStellarService
	_, err := svc.SubmitTransaction(context.Background(), stellartypes.SubmitTransactionRequest{})
	require.Error(t, err)
	require.Equal(t, codes.Unimplemented, status.Code(err))
}

func TestUnimplementedStellarService_GetTransaction(t *testing.T) {
	t.Parallel()

	var svc types.UnimplementedStellarService
	_, err := svc.GetTransaction(context.Background(), stellartypes.GetTransactionRequest{})
	require.Error(t, err)
	require.Equal(t, codes.Unimplemented, status.Code(err))
}

func TestUnimplementedStellarService_GetSigningAccount(t *testing.T) {
	t.Parallel()

	var svc types.UnimplementedStellarService
	_, err := svc.GetSigningAccount(context.Background())
	require.Error(t, err)
	require.Equal(t, codes.Unimplemented, status.Code(err))
}

func TestUnimplementedStellarService_GetLedgers(t *testing.T) {
	t.Parallel()

	var svc types.UnimplementedStellarService
	_, err := svc.GetLedgers(context.Background(), stellartypes.GetLedgersRequest{})
	require.Error(t, err)
	require.Equal(t, codes.Unimplemented, status.Code(err))
}
