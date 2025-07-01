package relayer

import (
	"context"
	"math/big"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	tonpb "github.com/smartcontractkit/chainlink-common/pkg/chains/ton"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	tontypes "github.com/smartcontractkit/chainlink-common/pkg/types/chains/ton"
)

func Test_TONDomainRoundTripThroughGRPC(t *testing.T) {
	t.Parallel()

	lis := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer()
	tonService := &staticTONService{}
	tonpb.RegisterTONServer(s, &tonServer{impl: tonService})

	go func() { _ = s.Serve(lis) }()
	defer s.Stop()

	ctx := t.Context()

	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			return lis.Dial()
		}),
	)
	require.NoError(t, err)
	defer conn.Close()

	client := &TONClient{grpcClient: tonpb.NewTONClient(conn)}

	blockID := &tontypes.BlockIDExt{Workchain: -1, Shard: 0x8000000000000000, SeqNo: 12345}
	balance := &tontypes.Balance{Balance: big.NewInt(987654321)}
	txStatus := types.Finalized
	t.Run("GetMasterchainInfo", func(t *testing.T) {
		expected := blockID
		tonService.getMasterchainInfo = func(ctx context.Context) (*tontypes.BlockIDExt, error) {
			return expected, nil
		}
		got, err := client.GetMasterchainInfo(ctx)
		require.NoError(t, err)
		require.Equal(t, expected, got)
	})

	t.Run("GetBlockData", func(t *testing.T) {
		exp := &tontypes.Block{GlobalID: 101}
		tonService.getBlockData = func(ctx context.Context, b *tontypes.BlockIDExt) (*tontypes.Block, error) {
			require.Equal(t, blockID, b)
			return exp, nil
		}
		got, err := client.GetBlockData(ctx, blockID)
		require.NoError(t, err)
		require.Equal(t, exp, got)
	})

	t.Run("GetAccountBalance", func(t *testing.T) {
		addr := "0:abc123"
		tonService.getAccountBalance = func(ctx context.Context, address string, block *tontypes.BlockIDExt) (*tontypes.Balance, error) {
			require.Equal(t, addr, address)
			require.Equal(t, blockID, block)
			return balance, nil
		}
		got, err := client.GetAccountBalance(ctx, addr, blockID)
		require.NoError(t, err)
		require.Equal(t, balance, got)
	})

	t.Run("SendTx", func(t *testing.T) {
		msg := tontypes.Message{Mode: 1, ToAddress: "0:abc", AmountNano: "1000"}
		tonService.sendTx = func(ctx context.Context, m tontypes.Message) error {
			require.Equal(t, msg, m)
			return nil
		}
		err := client.SendTx(ctx, msg)
		require.NoError(t, err)
	})

	t.Run("GetTxStatus", func(t *testing.T) {
		exitCode := tontypes.ExitCode(0)
		tonService.getTxStatus = func(ctx context.Context, lt uint64) (types.TransactionStatus, tontypes.ExitCode, error) {
			return txStatus, exitCode, nil
		}
		status, ec, err := client.GetTxStatus(ctx, 999)
		require.NoError(t, err)
		require.Equal(t, txStatus, status)
		require.Equal(t, tontypes.ExitCode(0), ec)
	})

	t.Run("GetTxExecutionFees", func(t *testing.T) {
		tonService.getTxFees = func(ctx context.Context, lt uint64) (*tontypes.TransactionFee, error) {
			return &tontypes.TransactionFee{TransactionFee: big.NewInt(8888)}, nil
		}
		fee, err := client.GetTxExecutionFees(ctx, 123)
		require.NoError(t, err)
		require.Equal(t, &tontypes.TransactionFee{TransactionFee: big.NewInt(8888)}, fee)
	})

	t.Run("HasFilter", func(t *testing.T) {
		name := "testFilter"
		tonService.hasFilter = func(ctx context.Context, s string) bool {
			return s == name
		}
		ok := client.HasFilter(ctx, name)
		require.True(t, ok)
	})

	t.Run("RegisterFilter", func(t *testing.T) {
		filter := tontypes.LPFilterQuery{Name: "filter1", Address: "0:addr", Retention: time.Second}
		tonService.registerFilter = func(ctx context.Context, f tontypes.LPFilterQuery) error {
			require.Equal(t, filter, f)
			return nil
		}
		require.NoError(t, client.RegisterFilter(ctx, filter))
	})

	t.Run("UnregisterFilter", func(t *testing.T) {
		filterName := "filter2"
		tonService.unregisterFilter = func(ctx context.Context, s string) error {
			require.Equal(t, filterName, s)
			return nil
		}
		require.NoError(t, client.UnregisterFilter(ctx, filterName))
	})
}

type staticTONService struct {
	getMasterchainInfo func(ctx context.Context) (*tontypes.BlockIDExt, error)
	getBlockData       func(ctx context.Context, b *tontypes.BlockIDExt) (*tontypes.Block, error)
	getAccountBalance  func(ctx context.Context, addr string, b *tontypes.BlockIDExt) (*tontypes.Balance, error)
	sendTx             func(ctx context.Context, msg tontypes.Message) error
	getTxStatus        func(ctx context.Context, lt uint64) (types.TransactionStatus, tontypes.ExitCode, error)
	getTxFees          func(ctx context.Context, lt uint64) (*tontypes.TransactionFee, error)
	hasFilter          func(ctx context.Context, name string) bool
	registerFilter     func(ctx context.Context, filter tontypes.LPFilterQuery) error
	unregisterFilter   func(ctx context.Context, name string) error
}

func (s *staticTONService) GetMasterchainInfo(ctx context.Context) (*tontypes.BlockIDExt, error) {
	return s.getMasterchainInfo(ctx)
}
func (s *staticTONService) GetBlockData(ctx context.Context, b *tontypes.BlockIDExt) (*tontypes.Block, error) {
	return s.getBlockData(ctx, b)
}
func (s *staticTONService) GetAccountBalance(ctx context.Context, addr string, b *tontypes.BlockIDExt) (*tontypes.Balance, error) {
	return s.getAccountBalance(ctx, addr, b)
}
func (s *staticTONService) SendTx(ctx context.Context, msg tontypes.Message) error {
	return s.sendTx(ctx, msg)
}
func (s *staticTONService) GetTxStatus(ctx context.Context, lt uint64) (types.TransactionStatus, tontypes.ExitCode, error) {
	return s.getTxStatus(ctx, lt)
}
func (s *staticTONService) GetTxExecutionFees(ctx context.Context, lt uint64) (*tontypes.TransactionFee, error) {
	return s.getTxFees(ctx, lt)
}
func (s *staticTONService) HasFilter(ctx context.Context, name string) bool {
	return s.hasFilter(ctx, name)
}
func (s *staticTONService) RegisterFilter(ctx context.Context, filter tontypes.LPFilterQuery) error {
	return s.registerFilter(ctx, filter)
}
func (s *staticTONService) UnregisterFilter(ctx context.Context, name string) error {
	return s.unregisterFilter(ctx, name)
}
