package relayer

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	evmpb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/evm"
	pbmocks "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/evm/mocks"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/chains/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/types/mocks"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
	mock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestEVMClient_GetTransactionFee(t *testing.T) {
	t.Parallel()
	mockGRPC := pbmocks.NewEVMClient(t)
	client := &evmClient{cl: mockGRPC}
	ctx := context.Background()
	res := big.NewInt(12345)
	id := "tx1"
	mockGRPC.On("GetTransactionFee", ctx, mock.MatchedBy(func(req *evmpb.GetTransactionFeeRequest) bool {
		return req.TransactionId == id
	})).Return(&evmpb.GetTransactionFeeReply{
		TransationFee: pb.NewBigIntFromInt(res),
	}, nil)

	resp, err := client.GetTransactionFee(ctx, id)
	require.NoError(t, err)
	require.Equal(t, res, resp.TransactionFee)
}

func TestEVMClient_CallContract(t *testing.T) {
	t.Parallel()
	mockGRPC := pbmocks.NewEVMClient(t)
	client := &evmClient{cl: mockGRPC}
	ctx := context.Background()

	t.Run("fields are matching", func(t *testing.T) {
		expectedData := []byte("data")
		mockGRPC.On("CallContract", ctx, mock.MatchedBy(func(req *evmpb.CallContractRequest) bool {
			return req.Call.To.Address == "0xabc" && string(req.Call.Data.Abi) == string(expectedData) && req.ConfidenceLevel == pb.Confidence_Finalized
		})).Return(&evmpb.CallContractReply{
			Data: &evmpb.ABIPayload{Abi: []byte("result")},
		}, nil)

		msg := &evm.CallMsg{To: "0xabc", From: "0xdef", Data: expectedData}
		resp, err := client.CallContract(ctx, msg, primitives.Finalized)
		require.NoError(t, err)
		require.Equal(t, []byte("result"), resp)
	})

	t.Run("from can be nil", func(t *testing.T) {
		mockGRPC.On("CallContract", ctx, mock.MatchedBy(func(req *evmpb.CallContractRequest) bool {
			return req.Call.From == nil
		})).Return(&evmpb.CallContractReply{
			Data: &evmpb.ABIPayload{Abi: []byte("result")},
		}, nil)

		msg := &evm.CallMsg{To: "123", From: "", Data: []byte{1, 2, 3}}
		resp, err := client.CallContract(ctx, msg, primitives.Finalized)
		require.NoError(t, err)
		require.Equal(t, []byte("result"), resp)
	})
}

func TestEVMClient_Getlogs(t *testing.T) {
	t.Parallel()
	mockGRPC := pbmocks.NewEVMClient(t)
	client := &evmClient{cl: mockGRPC}
	ctx := context.Background()
	addr := "0xabc"
	data := []byte("abi")
	t.Run("fields can be nil", func(t *testing.T) {
		mockGRPC.On("GetLogs", ctx, mock.MatchedBy(func(req *evmpb.GetLogsRequest) bool {
			return req.FilterQuery != nil &&
				req.FilterQuery.BlockHash == nil &&
				req.FilterQuery.FromBlock == nil &&
				req.FilterQuery.ToBlock == nil
		})).Return(&evmpb.GetLogsReply{
			Logs: []*evmpb.Log{{Address: &evmpb.Address{Address: addr}, Data: &evmpb.ABIPayload{Abi: data}}},
		}, nil)

		logs, err := client.GetLogs(ctx, evm.FilterQuery{})
		require.NoError(t, err)
		require.Len(t, logs, 1)
		require.Equal(t, addr, logs[0].Address)
		require.Equal(t, data, logs[0].Data)
	})
	t.Run("full fields match", func(t *testing.T) {
		fq := evm.FilterQuery{
			BlockHash: "0xblock",
			FromBlock: big.NewInt(100),
			ToBlock:   big.NewInt(200),
			Addresses: []string{"0xabc"},
			Topics:    []string{"0xtopic1"},
		}

		mockGRPC.On("GetLogs", ctx, mock.MatchedBy(func(req *evmpb.GetLogsRequest) bool {
			return req.FilterQuery.BlockHash.Hash == fq.BlockHash &&
				req.FilterQuery.FromBlock.Int().Cmp(fq.FromBlock) == 0 &&
				req.FilterQuery.ToBlock.Int().Cmp(fq.ToBlock) == 0 &&
				len(req.FilterQuery.Addresses) == 1 && req.FilterQuery.Addresses[0].Address == fq.Addresses[0] &&
				len(req.FilterQuery.Topics) == 1 && req.FilterQuery.Topics[0].Hash == fq.Topics[0]
		})).Return(&evmpb.GetLogsReply{
			Logs: []*evmpb.Log{{Address: &evmpb.Address{Address: "0xabc"}}},
		}, nil)

		logs, err := client.GetLogs(ctx, fq)
		require.NoError(t, err)
		require.Len(t, logs, 1)
		require.Equal(t, "0xabc", logs[0].Address)
	})
}

func TestEVMClient_BalanceAt(t *testing.T) {
	t.Parallel()
	mockGRPC := pbmocks.NewEVMClient(t)
	client := &evmClient{cl: mockGRPC}
	ctx := context.Background()

	addr := "0xabc"
	bal := big.NewInt(123)
	blockNum := big.NewInt(123)
	mockGRPC.On("BalanceAt", ctx, mock.MatchedBy(func(req *evmpb.BalanceAtRequest) bool {
		return req.Account.Address == addr && req.BlockNumber.Int().Cmp(blockNum) == 0
	})).Return(&evmpb.BalanceAtReply{
		Balance: pb.NewBigIntFromInt(bal),
	}, nil)

	resp, err := client.BalanceAt(ctx, addr, blockNum)
	require.NoError(t, err)
	require.Equal(t, bal, resp)
}

func TestEVMClient_EstimateGas(t *testing.T) {
	t.Parallel()
	mockGRPC := pbmocks.NewEVMClient(t)
	client := &evmClient{cl: mockGRPC}
	ctx := context.Background()

	from := "0xbbb"
	to := "0xaaa"
	data := []byte("foo")
	gas := uint64(21000)
	msg := &evm.CallMsg{To: to, From: from, Data: data}

	mockGRPC.On("EstimateGas", ctx, mock.MatchedBy(func(req *evmpb.EstimateGasRequest) bool {
		return req.Msg.To.Address == to && req.Msg.From.Address == from && string(req.Msg.Data.Abi) == string(data)
	})).Return(&evmpb.EstimateGasReply{Gas: gas}, nil)

	resp, err := client.EstimateGas(ctx, msg)
	require.NoError(t, err)
	require.Equal(t, gas, resp)
}

func TestEVMClient_TransactionByHash(t *testing.T) {
	t.Parallel()
	mockGRPC := pbmocks.NewEVMClient(t)
	client := &evmClient{cl: mockGRPC}
	ctx := context.Background()
	hash := "0xx"
	mockGRPC.On("GetTransactionByHash", ctx, mock.MatchedBy(func(req *evmpb.GetTransactionByHashRequest) bool {
		return req.Hash.Hash == hash
	})).Return(&evmpb.GetTransactionByHashReply{
		Transaction: &evmpb.Transaction{Hash: &evmpb.Hash{Hash: hash}},
	}, nil)

	tx, err := client.TransactionByHash(ctx, hash)
	require.NoError(t, err)
	require.Equal(t, hash, tx.Hash)
}

func TestEVMClient_TransactionReceipt(t *testing.T) {
	t.Parallel()
	mockGRPC := pbmocks.NewEVMClient(t)
	client := &evmClient{cl: mockGRPC}
	ctx := context.Background()
	hash := "0xhash"
	mockGRPC.On("GetTransactionReceipt", ctx, mock.MatchedBy(func(req *evmpb.GetReceiptRequest) bool {
		return req.Hash.Hash == hash
	})).Return(&evmpb.GetReceiptReply{
		Receipt: &evmpb.Receipt{TxHash: &evmpb.Hash{Hash: hash}},
	}, nil)

	r, err := client.TransactionReceipt(ctx, hash)
	require.NoError(t, err)
	require.Equal(t, hash, r.TxHash)
}

func TestEVMClient_RegisterLogTracking(t *testing.T) {
	t.Parallel()
	mockGRPC := pbmocks.NewEVMClient(t)
	client := &evmClient{cl: mockGRPC}
	ctx := context.Background()

	name := "testFilter"
	addresses := []string{"0xabc"}
	eventSigs := []string{"0xevt"}
	topic2 := []string{"0x02"}
	filter := evm.LPFilterQuery{
		Name:      name,
		Addresses: addresses,
		EventSigs: eventSigs,
		Topic2:    topic2,
	}

	mockGRPC.On("RegisterLogTracking", ctx, mock.MatchedBy(func(req *evmpb.RegisterLogTrackingRequest) bool {
		return req.Filter.Name == name &&
			len(req.Filter.Addresses) == 1 && req.Filter.Addresses[0].Address == addresses[0] &&
			len(req.Filter.EventSigs) == 1 && req.Filter.EventSigs[0].Hash == eventSigs[0] &&
			len(req.Filter.Topic2) == 1 && req.Filter.Topic2[0].Hash == topic2[0]
	})).Return(nil, nil)

	err := client.RegisterLogTracking(ctx, filter)
	require.NoError(t, err)
}

func TestEVMServer_CallContract(t *testing.T) {
	t.Parallel()
	mockImpl := mocks.NewEVMService(t)
	server := evmServer{impl: mockImpl}
	ctx := t.Context()
	from := "0xabc"
	to := "0xdef"
	data := []byte("abcd")
	confidence := primitives.Finalized

	mockImpl.On("CallContract", ctx, mock.MatchedBy(func(m *evm.CallMsg) bool {
		return m.From == from && m.To == to && string(m.Data) == string(data)
	}), confidence).Return([]byte("response"), nil)

	req := &evmpb.CallContractRequest{
		Call: &evmpb.CallMsg{
			From: &evmpb.Address{Address: from},
			To:   &evmpb.Address{Address: to},
			Data: &evmpb.ABIPayload{Abi: data},
		},
		ConfidenceLevel: 1,
	}

	resp, err := server.CallContract(ctx, req)
	require.NoError(t, err)
	require.Equal(t, []byte("response"), resp.Data.Abi)
}

func TestEVMServer_GetTransactionFee(t *testing.T) {
	mockImpl := mocks.NewEVMService(t)
	server := evmServer{impl: mockImpl}
	ctx := context.Background()

	txID := "tx123"
	txFee := big.NewInt(9999)
	mockImpl.On("GetTransactionFee", ctx, txID).Return(&types.TransactionFee{TransactionFee: txFee}, nil)

	req := &evmpb.GetTransactionFeeRequest{TransactionId: txID}
	resp, err := server.GetTransactionFee(ctx, req)
	require.NoError(t, err)
	require.Equal(t, txFee, resp.TransationFee.Int())
}

func TestEVMServer_BalanceAt(t *testing.T) {
	mockImpl := mocks.NewEVMService(t)
	server := evmServer{impl: mockImpl}
	ctx := context.Background()

	addr := "0xabc"
	block := big.NewInt(12345)
	balance := big.NewInt(5000)
	mockImpl.On("BalanceAt", ctx, addr, mock.MatchedBy(func(bn *big.Int) bool {
		return bn.Cmp(block) == 0 &&
			addr == addr
	})).Return(balance, nil)

	req := &evmpb.BalanceAtRequest{
		Account:     &evmpb.Address{Address: addr},
		BlockNumber: pb.NewBigIntFromInt(block),
	}
	resp, err := server.BalanceAt(ctx, req)
	require.NoError(t, err)
	require.Equal(t, balance, resp.Balance.Int())
}

func TestEVMServer_RegisterLogTracking(t *testing.T) {
	mockImpl := mocks.NewEVMService(t)
	server := evmServer{impl: mockImpl}
	ctx := context.Background()

	filter := evm.LPFilterQuery{
		Name:         "filter-1",
		Retention:    time.Second,
		Addresses:    []string{"0xaaa"},
		EventSigs:    []string{"0x111"},
		Topic2:       []string{"0x222"},
		Topic3:       []string{"0x333"},
		Topic4:       []string{"0x444"},
		MaxLogsKept:  100,
		LogsPerBlock: 10,
	}
	mockImpl.On("RegisterLogTracking", ctx, mock.MatchedBy(func(f evm.LPFilterQuery) bool {
		return f.Name == filter.Name && f.Retention == filter.Retention &&
			len(f.Addresses) == 1 && f.Addresses[0] == filter.Addresses[0] &&
			len(f.EventSigs) == 1 && f.EventSigs[0] == filter.EventSigs[0] &&
			len(f.Topic2) == 1 && f.Topic2[0] == filter.Topic2[0] &&
			len(f.Topic3) == 1 && f.Topic3[0] == filter.Topic3[0] &&
			len(f.Topic4) == 1 && f.Topic4[0] == filter.Topic4[0] &&
			f.MaxLogsKept == filter.MaxLogsKept && f.LogsPerBlock == filter.LogsPerBlock
	})).Return(nil)

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
