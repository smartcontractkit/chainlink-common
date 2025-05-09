package relayer

import (
	"context"
	"errors"
	"math/big"
	"testing"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	evmpb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/chains/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	txId         = "txid1"
	res          = big.NewInt(12345)
	balance      = big.NewInt(1222345)
	abi          = []byte("data")
	respAbi      = []byte("response")
	address      = evm.Address([]byte("0xabc"))
	address1     = evm.Address([]byte("0xabca"))
	blockHash    = evm.Hash([]byte("somehash"))
	fromBlock    = big.NewInt(10)
	blockNum     = big.NewInt(101)
	toBlock      = big.NewInt(145)
	topic        = evm.Hash([]byte("topic"))
	topic2       = evm.Hash([]byte("topic2"))
	topic3       = evm.Hash([]byte("topic3"))
	gas          = uint64(10)
	txHash       = evm.Hash([]byte("0xaaa"))
	eventSigHash = evm.Hash([]byte("0x654"))
	filterName   = "f name 1"
	maxLogKept   = uint64(10)
	logsPerBlock = uint64(1)
	retention    = time.Second
)

type staticEVMClient struct {
	t *testing.T

	FilterLogsFunc func(ctx context.Context, in *evmpb.FilterLogsRequest, opts ...grpc.CallOption) (*evmpb.FilterLogsReply, error)
}

func (s *staticEVMClient) GetTransactionFee(ctx context.Context, in *evmpb.GetTransactionFeeRequest, opts ...grpc.CallOption) (*evmpb.GetTransactionFeeReply, error) {
	require.Equal(s.t, txId, in.TransactionId)

	return &evmpb.GetTransactionFeeReply{
		TransationFee: pb.NewBigIntFromInt(res),
	}, nil
}
func (s *staticEVMClient) CallContract(ctx context.Context, in *evmpb.CallContractRequest, opts ...grpc.CallOption) (*evmpb.CallContractReply, error) {
	require.Equal(s.t, address, in.Call.To.Address)
	require.Equal(s.t, string(abi), string(in.Call.Data.Abi))
	require.Equal(s.t, in.BlockNumber.Int(), blockNum)
	return &evmpb.CallContractReply{
		Data: &evmpb.ABIPayload{Abi: respAbi},
	}, nil
}

func (s *staticEVMClient) FilterLogs(ctx context.Context, in *evmpb.FilterLogsRequest, opts ...grpc.CallOption) (*evmpb.FilterLogsReply, error) {
	return s.FilterLogsFunc(ctx, in, opts...)
}
func (s *staticEVMClient) BalanceAt(ctx context.Context, in *evmpb.BalanceAtRequest, opts ...grpc.CallOption) (*evmpb.BalanceAtReply, error) {
	require.Equal(s.t, address, in.Account.Address)
	require.Equal(s.t, 0, in.BlockNumber.Int().Cmp(blockNum))
	return &evmpb.BalanceAtReply{
		Balance: pb.NewBigIntFromInt(balance),
	}, nil
}
func (s *staticEVMClient) EstimateGas(ctx context.Context, in *evmpb.EstimateGasRequest, opts ...grpc.CallOption) (*evmpb.EstimateGasReply, error) {
	require.Equal(s.t, address1, in.Msg.From.Address)
	require.Equal(s.t, address, in.Msg.To.Address)
	require.Equal(s.t, string(abi), string(in.Msg.Data.Abi))
	return &evmpb.EstimateGasReply{Gas: gas}, nil
}
func (s *staticEVMClient) GetTransactionByHash(ctx context.Context, in *evmpb.GetTransactionByHashRequest, opts ...grpc.CallOption) (*evmpb.GetTransactionByHashReply, error) {

	require.Equal(s.t, txHash, in.Hash.Hash)
	return &evmpb.GetTransactionByHashReply{
		Transaction: &evmpb.Transaction{
			Hash: &evmpb.Hash{Hash: txHash},
			Data: &evmpb.ABIPayload{Abi: respAbi},
			To:   &evmpb.Address{Address: address},
		},
	}, nil
}
func (s *staticEVMClient) GetTransactionReceipt(ctx context.Context, in *evmpb.GetReceiptRequest, opts ...grpc.CallOption) (*evmpb.GetReceiptReply, error) {
	require.Equal(s.t, txHash, in.Hash.Hash)
	return &evmpb.GetReceiptReply{
		Receipt: &evmpb.Receipt{TxHash: &evmpb.Hash{Hash: txHash}},
	}, nil
}
func (s *staticEVMClient) LatestAndFinalizedHead(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*evmpb.LatestAndFinalizedHeadReply, error) {
	return nil, errors.New("unimplemented")
}

func (s *staticEVMClient) QueryLogsFromCache(ctx context.Context, in *evmpb.QueryLogsFromCacheRequest, opts ...grpc.CallOption) (*evmpb.QueryLogsFromCacheReply, error) {
	return nil, errors.New("unimplemented")
}
func (s *staticEVMClient) RegisterLogTracking(ctx context.Context, req *evmpb.RegisterLogTrackingRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	require.Equal(s.t, filterName, req.Filter.Name)
	require.Equal(s.t, 1, len(req.Filter.Addresses))
	require.Equal(s.t, address, req.Filter.Addresses[0].Address)
	require.Equal(s.t, 1, len(req.Filter.EventSigs))
	require.Equal(s.t, eventSigHash, req.Filter.EventSigs[0].Hash)
	require.Equal(s.t, 1, len(req.Filter.Topic2))
	require.Equal(s.t, topic, req.Filter.Topic2[0].Hash)
	return nil, nil
}

func (s *staticEVMClient) UnregisterLogTracking(ctx context.Context, in *evmpb.UnregisterLogTrackingRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return nil, errors.New("unimplemented")
}
func (s *staticEVMClient) GetTransactionStatus(ctx context.Context, in *pb.GetTransactionStatusRequest, opts ...grpc.CallOption) (*pb.GetTransactionStatusReply, error) {
	return nil, errors.New("unimplemented")
}

func TestEVMClient_GetTransactionFee(t *testing.T) {
	t.Parallel()
	client := &evmClient{cl: &staticEVMClient{t: t}}
	ctx := context.Background()

	resp, err := client.GetTransactionFee(ctx, txId)
	require.NoError(t, err)
	require.Equal(t, res, resp.TransactionFee)
}

func TestEVMClient_CallContract(t *testing.T) {
	t.Parallel()
	client := &evmClient{cl: &staticEVMClient{t: t}}
	ctx := context.Background()

	msg := &evm.CallMsg{To: "0xabc", From: "0xdef", Data: abi}
	resp, err := client.CallContract(ctx, msg, primitives.Finalized)
	require.NoError(t, err)
	require.Equal(t, respAbi, resp)
}

func TestEVMClient_Getlogs(t *testing.T) {
	t.Parallel()
	static := &staticEVMClient{t: t}
	client := &evmClient{cl: static}
	ctx := context.Background()
	t.Run("fields can be nil", func(t *testing.T) {
		static.FilterLogsFunc = func(ctx context.Context, in *evmpb.FilterLogsRequest, opts ...grpc.CallOption) (*evmpb.FilterLogsReply, error) {
			require.NotNil(t, in.FilterQuery)
			require.Nil(t, in.FilterQuery.FromBlock)
			require.Nil(t, in.FilterQuery.ToBlock)
			require.Nil(t, in.FilterQuery.BlockHash)
			return &evmpb.FilterLogsReply{
				Logs: []*evmpb.Log{{Address: &evmpb.Address{Address: address}, Data: &evmpb.ABIPayload{Abi: respAbi}}},
			}, nil
		}

		logs, err := client.FilterLogs(ctx, evm.FilterQuery{})
		require.NoError(t, err)
		require.Len(t, logs, 1)
		require.Equal(t, address, logs[0].Address)
		require.Equal(t, respAbi, logs[0].Data)
	})
	t.Run("full fields match", func(t *testing.T) {
		fq := evm.FilterQuery{
			BlockHash: blockHash,
			FromBlock: fromBlock,
			ToBlock:   toBlock,
			Addresses: []string{address},
			Topics:    [][]string{{topic, topic2}, {topic3}},
		}

		static.FilterLogsFunc = func(ctx context.Context, in *evmpb.FilterLogsRequest, opts ...grpc.CallOption) (*evmpb.FilterLogsReply, error) {
			require.NotNil(t, in.FilterQuery)
			require.Equal(t, fromBlock, in.FilterQuery.FromBlock.Int())
			require.Equal(t, toBlock, in.FilterQuery.ToBlock.Int())
			require.Equal(t, blockHash, in.FilterQuery.BlockHash.Hash)
			require.Equal(t, topic, in.FilterQuery.Topics[0].Topic[0].Hash)
			require.Equal(t, topic2, in.FilterQuery.Topics[0].Topic[1].Hash)
			require.Equal(t, topic3, in.FilterQuery.Topics[1].Topic[0].Hash)
			return &evmpb.FilterLogsReply{
				Logs: []*evmpb.Log{{Address: &evmpb.Address{Address: address}, Data: &evmpb.ABIPayload{Abi: respAbi}}},
			}, nil
		}

		logs, err := client.FilterLogs(ctx, fq)
		require.NoError(t, err)
		require.Len(t, logs, 1)
		require.Equal(t, address, logs[0].Address)
	})
}

func TestEVMClient_BalanceAt(t *testing.T) {
	t.Parallel()
	client := &evmClient{cl: &staticEVMClient{t: t}}
	ctx := context.Background()

	resp, err := client.BalanceAt(ctx, address, blockNum)
	require.NoError(t, err)
	require.Equal(t, balance, resp)
}

func TestEVMClient_EstimateGas(t *testing.T) {
	t.Parallel()
	client := &evmClient{cl: &staticEVMClient{t: t}}
	ctx := context.Background()

	msg := &evm.CallMsg{To: address, From: address1, Data: abi}

	resp, err := client.EstimateGas(ctx, msg)
	require.NoError(t, err)
	require.Equal(t, gas, resp)
}

func TestEVMClient_TransactionByHash(t *testing.T) {
	t.Parallel()
	client := &evmClient{cl: &staticEVMClient{t: t}}
	ctx := context.Background()

	tx, err := client.TransactionByHash(ctx, txHash)
	require.NoError(t, err)
	require.Equal(t, txHash, tx.Hash)
	require.Equal(t, address, tx.To)
	require.Equal(t, string(respAbi), string(tx.Data))
}

func TestEVMClient_TransactionReceipt(t *testing.T) {
	t.Parallel()
	client := &evmClient{cl: &staticEVMClient{t: t}}
	ctx := context.Background()

	r, err := client.TransactionReceipt(ctx, txHash)
	require.NoError(t, err)
	require.Equal(t, txHash, r.TxHash)
}

func TestEVMClient_RegisterLogTracking(t *testing.T) {
	t.Parallel()
	client := &evmClient{cl: &staticEVMClient{t: t}}
	ctx := context.Background()

	name := filterName
	addresses := []string{address}
	eventSigs := []string{eventSigHash}
	topic2 := []string{topic}
	filter := evm.LPFilterQuery{
		Name:      name,
		Addresses: addresses,
		EventSigs: eventSigs,
		Topic2:    topic2,
	}

	err := client.RegisterLogTracking(ctx, filter)
	require.NoError(t, err)
}

type staticEVMService struct {
	t *testing.T
}

func (ss *staticEVMService) CallContract(ctx context.Context, msg *evm.CallMsg, confidence primitives.ConfidenceLevel) ([]byte, error) {
	require.Equal(ss.t, address, msg.From)
	require.Equal(ss.t, address1, msg.To)
	require.Equal(ss.t, abi, msg.Data)
	require.Equal(ss.t, primitives.Finalized, confidence)

	return respAbi, nil
}

func (ss *staticEVMService) GetTransactionFee(ctx context.Context, transactionID string) (*evm.TransactionFee, error) {
	require.Equal(ss.t, transactionID, txId)
	return &evm.TransactionFee{TransactionFee: res}, nil
}

func (ss *staticEVMService) BalanceAt(ctx context.Context, account evm.Address, blockNumber *big.Int) (*big.Int, error) {
	require.Equal(ss.t, address, account)
	require.Equal(ss.t, blockNum, blockNumber)
	return balance, nil
}

func (ss *staticEVMService) RegisterLogTracking(ctx context.Context, filter evm.LPFilterQuery) error {
	require.Equal(ss.t, filterName, filter.Name)
	require.Equal(ss.t, retention, filter.Retention)
	require.Equal(ss.t, address, filter.Addresses[0])
	require.Equal(ss.t, eventSigHash, filter.EventSigs[0])
	require.Equal(ss.t, topic, filter.Topic2[0])
	require.Equal(ss.t, topic2, filter.Topic3[0])
	require.Equal(ss.t, topic3, filter.Topic4[0])
	require.Equal(ss.t, maxLogKept, filter.MaxLogsKept)
	require.Equal(ss.t, logsPerBlock, filter.LogsPerBlock)
	return nil
}

func (ss *staticEVMService) FilterLogs(ctx context.Context, filterQuery evm.FilterQuery) ([]*evm.Log, error) {
	return nil, errors.New("unimplemented")
}
func (ss *staticEVMService) EstimateGas(ctx context.Context, call *evm.CallMsg) (uint64, error) {
	return 0, errors.New("unimplemented")
}
func (ss *staticEVMService) TransactionByHash(ctx context.Context, hash evm.Hash) (*evm.Transaction, error) {
	return nil, errors.New("unimplemented")
}
func (ss *staticEVMService) TransactionReceipt(ctx context.Context, txHash evm.Hash) (*evm.Receipt, error) {
	return nil, errors.New("unimplemented")
}
func (ss *staticEVMService) LatestAndFinalizedHead(ctx context.Context) (latest evm.Head, finalized evm.Head, err error) {
	err = errors.New("unimplemented")
	return
}
func (ss *staticEVMService) QueryLogsFromCache(ctx context.Context, filterQuery []query.Expression,
	limitAndSort query.LimitAndSort, confidenceLevel primitives.ConfidenceLevel) ([]*evm.Log, error) {
	return nil, nil
}
func (ss *staticEVMService) UnregisterLogTracking(ctx context.Context, filterName string) error {
	return errors.New("unimplemented")
}
func (ss *staticEVMService) GetTransactionStatus(ctx context.Context, transactionID string) (types.TransactionStatus, error) {
	return types.Unknown, errors.New("unimplemented")
}

func TestEVMServer_CallContract(t *testing.T) {
	t.Parallel()
	server := evmServer{impl: &staticEVMService{t: t}}
	ctx := t.Context()

	req := &evmpb.CallContractRequest{
		Call: &evmpb.CallMsg{
			From: &evmpb.Address{Address: address},
			To:   &evmpb.Address{Address: address1},
			Data: &evmpb.ABIPayload{Abi: abi},
		},
		ConfidenceLevel: 1,
	}

	resp, err := server.CallContract(ctx, req)
	require.NoError(t, err)
	require.Equal(t, respAbi, resp.Data.Abi)
}

func TestEVMServer_GetTransactionFee(t *testing.T) {
	t.Parallel()
	server := evmServer{impl: &staticEVMService{t: t}}
	ctx := context.Background()

	req := &evmpb.GetTransactionFeeRequest{TransactionId: txId}
	resp, err := server.GetTransactionFee(ctx, req)
	require.NoError(t, err)
	require.Equal(t, res, resp.TransationFee.Int())
}

func TestEVMServer_BalanceAt(t *testing.T) {
	t.Parallel()
	server := evmServer{impl: &staticEVMService{t: t}}
	ctx := context.Background()

	req := &evmpb.BalanceAtRequest{
		Account:     &evmpb.Address{Address: address},
		BlockNumber: pb.NewBigIntFromInt(blockNum),
	}
	resp, err := server.BalanceAt(ctx, req)
	require.NoError(t, err)
	require.Equal(t, balance, resp.Balance.Int())
}

func TestEVMServer_RegisterLogTracking(t *testing.T) {
	t.Parallel()
	server := evmServer{impl: &staticEVMService{t: t}}
	ctx := context.Background()

	filter := evm.LPFilterQuery{
		Name:         filterName,
		Retention:    retention,
		Addresses:    []string{address},
		EventSigs:    []string{eventSigHash},
		Topic2:       []string{topic},
		Topic3:       []string{topic2},
		Topic4:       []string{topic3},
		MaxLogsKept:  maxLogKept,
		LogsPerBlock: logsPerBlock,
	}
	req := &evmpb.RegisterLogTrackingRequest{
		Filter: &evmpb.LPFilter{
			Name:          filter.Name,
			RetentionTime: int64(filter.Retention),
			Addresses:     []*evmpb.Address{{Address: filter.Addresses[0]}},
			EventSigs:     []*evmpb.Hash{{Hash: filter.EventSigs[0]}},
			Topic2:        []*evmpb.Hash{{Hash: filter.Topic2[0]}},
			Topic3:        []*evmpb.Hash{{Hash: filter.Topic3[0]}},
			Topic4:        []*evmpb.Hash{{Hash: filter.Topic4[0]}},
			MaxLogsKept:   filter.MaxLogsKept,
			LogsPerBlock:  filter.LogsPerBlock,
		},
	}
	_, err := server.RegisterLogTracking(ctx, req)
	require.NoError(t, err)
}
