package relayer

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	aptospb "github.com/smartcontractkit/chainlink-common/pkg/chains/aptos"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	loopnet "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	aptostypes "github.com/smartcontractkit/chainlink-common/pkg/types/chains/aptos"
)

func Test_AptosDomainRoundTripThroughGRPC(t *testing.T) {
	t.Parallel()

	lis := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer()

	aptosService := &staticAptosService{}
	aptospb.RegisterAptosServer(s, newAptosServer(aptosService, &loopnet.BrokerExt{
		BrokerConfig: loopnet.BrokerConfig{
			Logger: logger.Test(t),
		},
	}))

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

	client := &AptosClient{grpcClient: aptospb.NewAptosClient(conn)}

	t.Run("LedgerVersion", func(t *testing.T) {
		const expectedLedgerVersion = uint64(12345)
		aptosService.ledgerVersion = func(context.Context) (uint64, error) {
			return expectedLedgerVersion, nil
		}

		got, err := client.LedgerVersion(ctx)
		require.NoError(t, err)
		require.Equal(t, expectedLedgerVersion, got)
	})

	t.Run("View_WithLedgerVersion", func(t *testing.T) {
		ledgerVersion := uint64(77)
		expectedPayload := &aptostypes.ViewPayload{
			Module: aptostypes.ModuleID{
				Address: aptostypes.AccountAddress{},
				Name:    "module_name",
			},
			Function: "function_name",
			Args:     [][]byte{[]byte{0x01, 0x02}},
		}
		expectedData := []byte(`{"value":"ok"}`)

		aptosService.view = func(_ context.Context, req aptostypes.ViewRequest) (*aptostypes.ViewReply, error) {
			require.NotNil(t, req.Payload)
			require.Equal(t, expectedPayload.Function, req.Payload.Function)
			require.Equal(t, expectedPayload.Module.Name, req.Payload.Module.Name)
			require.Equal(t, expectedPayload.Args, req.Payload.Args)
			require.NotNil(t, req.LedgerVersion)
			require.Equal(t, ledgerVersion, *req.LedgerVersion)
			return &aptostypes.ViewReply{Data: expectedData}, nil
		}

		got, err := client.View(ctx, aptostypes.ViewRequest{
			Payload:       expectedPayload,
			LedgerVersion: &ledgerVersion,
		})
		require.NoError(t, err)
		require.Equal(t, expectedData, got.Data)
	})

	t.Run("View_WithoutLedgerVersion", func(t *testing.T) {
		expectedPayload := &aptostypes.ViewPayload{
			Module: aptostypes.ModuleID{
				Address: aptostypes.AccountAddress{},
				Name:    "module_name",
			},
			Function: "function_name",
		}
		expectedData := []byte(`{"value":"latest"}`)

		aptosService.view = func(_ context.Context, req aptostypes.ViewRequest) (*aptostypes.ViewReply, error) {
			require.NotNil(t, req.Payload)
			require.Equal(t, expectedPayload.Function, req.Payload.Function)
			require.Equal(t, expectedPayload.Module.Name, req.Payload.Module.Name)
			require.Nil(t, req.LedgerVersion)
			return &aptostypes.ViewReply{Data: expectedData}, nil
		}

		got, err := client.View(ctx, aptostypes.ViewRequest{
			Payload: expectedPayload,
		})
		require.NoError(t, err)
		require.Equal(t, expectedData, got.Data)
	})
}

type staticAptosService struct {
	types.UnimplementedAptosService
	ledgerVersion func(ctx context.Context) (uint64, error)
	view          func(ctx context.Context, req aptostypes.ViewRequest) (*aptostypes.ViewReply, error)
}

func (s *staticAptosService) LedgerVersion(ctx context.Context) (uint64, error) {
	return s.ledgerVersion(ctx)
}

func (s *staticAptosService) View(ctx context.Context, req aptostypes.ViewRequest) (*aptostypes.ViewReply, error) {
	return s.view(ctx, req)
}
