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

	evmpb "github.com/smartcontractkit/chainlink-common/pkg/loop/chain-capabilities/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/chains/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
	evmprimitives "github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives/evm"
)

var (
	txId         = "txid1"
	txIndex      = 10
	txFee        = big.NewInt(12345)
	balance      = big.NewInt(1222345)
	gasPrice     = big.NewInt(12344)
	abi          = []byte("data")
	respAbi      = []byte("response")
	address      = evm.Address{1, 2, 3}
	address1     = evm.Address{10, 11, 14}
	blockHash    = evm.Hash{22, 33, 44}
	parentHash   = evm.Hash{01, 33, 44}
	fromBlock    = big.NewInt(10)
	blockNum     = big.NewInt(101)
	toBlock      = big.NewInt(145)
	topic        = evm.Hash{21, 3, 4}
	topic2       = evm.Hash{33, 1, 33}
	topic3       = evm.Hash{20, 19, 17}
	gas          = uint64(10)
	txHash       = evm.Hash{5, 3, 44}
	eventSigHash = evm.Hash{14, 16, 29}
	filterName   = "f name 1"
	maxLogKept   = uint64(10)
	logsPerBlock = uint64(1)
	retention    = time.Second
)

func Test_EVMDomainRoundTripThroughGRPC(t *testing.T) {
	t.Parallel()

	lis := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer()
	evmService := &staticEVMService{}
	evmpb.RegisterEVMServer(s, &EvmServer{impl: evmService})

	go func() {
		_ = s.Serve(lis)
	}()
	defer s.Stop()

	ctx := t.Context()

	//nolint:staticcheck
	conn, err := grpc.DialContext(
		ctx,
		"bufnet",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			return lis.Dial()
		}),
	)
	require.NoError(t, err)

	defer conn.Close()
	client := &EVMClient{
		grpcClient: evmpb.NewEVMClient(conn),
	}
	t.Run("BalanceAt", func(t *testing.T) {
		evmService.staticBalanceAt = func(ctx context.Context, account evm.Address, blockNumber *big.Int) (*big.Int, error) {
			require.Equal(t, account, address)
			require.Equal(t, blockNumber, blockNum)
			return balance, nil
		}

		resp, err := client.BalanceAt(ctx, address, blockNum)
		require.NoError(t, err)
		require.Equal(t, resp, balance)
	})

	t.Run("CallContract", func(t *testing.T) {
		expMsg := &evm.CallMsg{
			To:   address,
			From: address1,
			Data: abi,
		}
		evmService.staticCallContract = func(ctx context.Context, msg *evm.CallMsg, blockNumber *big.Int) ([]byte, error) {
			require.Equal(t, expMsg, msg)
			return respAbi, nil
		}

		resp, err := client.CallContract(ctx, expMsg, blockNum)
		require.NoError(t, err)
		require.Equal(t, respAbi, resp)
	})

	t.Run("RegisterLogTracking", func(t *testing.T) {
		name := filterName
		addresses := []evm.Address{address}
		eventSigs := []evm.Hash{eventSigHash}
		topic2 := []evm.Hash{topic}
		topic3 := []evm.Hash{topic, topic3}
		expFilter := evm.LPFilterQuery{
			Name:      name,
			Addresses: addresses,
			EventSigs: eventSigs,
			Topic2:    topic2,
			Retention: retention,
			Topic3:    topic3,
		}
		evmService.staticRegisterLogTracking = func(ctx context.Context, filter evm.LPFilterQuery) error {
			require.Equal(t, expFilter.Name, filter.Name)
			require.Equal(t, expFilter.Addresses, filter.Addresses)
			require.Equal(t, expFilter.EventSigs, filter.EventSigs)
			require.Equal(t, expFilter.Topic2, filter.Topic2)
			require.Equal(t, expFilter.Topic3, filter.Topic3)
			require.Equal(t, expFilter.Retention, filter.Retention)

			return nil
		}

		err := client.RegisterLogTracking(ctx, expFilter)
		require.NoError(t, err)
	})

	t.Run("GetTransactionFee", func(t *testing.T) {
		evmService.staticGetTransactionFee = func(ctx context.Context, transactionID types.IdempotencyKey) (*evm.TransactionFee, error) {
			require.Equal(t, txId, transactionID)
			return &evm.TransactionFee{
				TransactionFee: txFee,
			}, nil
		}

		fee, err := client.GetTransactionFee(ctx, txId)
		require.NoError(t, err)
		require.Equal(t, txFee, fee.TransactionFee)
	})

	t.Run("GetTransactionStatus", func(t *testing.T) {
		evmService.staticGetTransactionStatus = func(ctx context.Context, transactionID types.IdempotencyKey) (types.TransactionStatus, error) {
			require.Equal(t, txId, transactionID)
			return types.Finalized, nil
		}

		got, err := client.GetTransactionStatus(ctx, txId)
		require.NoError(t, err)
		require.Equal(t, got, types.Finalized)
	})

	t.Run("FilterLogs", func(t *testing.T) {
		expFQ := evm.FilterQuery{
			BlockHash: blockHash,
			FromBlock: fromBlock,
			ToBlock:   toBlock,
			Addresses: []evm.Address{address},
			Topics:    [][]evm.Hash{{topic, topic2}, {topic3}},
		}
		expLog := []*evm.Log{
			{
				LogIndex:    1,
				BlockHash:   blockHash,
				BlockNumber: blockNum,
				Topics:      []evm.Hash{topic, topic2},
				EventSig:    eventSigHash,
				Address:     address,
				TxHash:      txHash,
				Data:        abi,
				Removed:     false,
			},
		}
		evmService.staticFilterLogs = func(ctx context.Context, fq evm.FilterQuery) ([]*evm.Log, error) {
			require.Equal(t, expFQ, fq)
			return expLog, nil
		}

		logs, err := client.FilterLogs(ctx, expFQ)
		require.NoError(t, err)
		require.Equal(t, expLog, logs)

	})

	t.Run("EstimateGas", func(t *testing.T) {
		expMsg := &evm.CallMsg{
			To:   address,
			From: address1,
			Data: abi,
		}
		evmService.staticEstimateGas = func(ctx context.Context, call *evm.CallMsg) (uint64, error) {
			require.Equal(t, expMsg, call)
			return gas, nil
		}

		resp, err := client.EstimateGas(ctx, expMsg)
		require.NoError(t, err)
		require.Equal(t, gas, resp)
	})

	t.Run("TransactionReceipt", func(t *testing.T) {
		expReceipt := &evm.Receipt{
			Status: 1,
			Logs: []*evm.Log{
				{
					LogIndex:    1,
					BlockHash:   blockHash,
					BlockNumber: blockNum,
					Topics:      []evm.Hash{topic, topic2},
					EventSig:    eventSigHash,
					Address:     address,
					TxHash:      txHash,
					Data:        abi,
					Removed:     false,
				},
			},
			TxHash:            txHash,
			EffectiveGasPrice: gasPrice,
			GasUsed:           gas,
			BlockNumber:       blockNum,
			TransactionIndex:  uint64(txIndex),
		}
		evmService.staticTransactionReceipt = func(ctx context.Context, got evm.Hash) (*evm.Receipt, error) {
			require.Equal(t, txHash, got)
			return expReceipt, nil
		}

		got, err := client.TransactionReceipt(ctx, txHash)
		require.NoError(t, err)
		require.Equal(t, expReceipt, got)

	})

	t.Run("TransactionByHash", func(t *testing.T) {
		expTx := &evm.Transaction{
			To:       address,
			Data:     abi,
			Hash:     txHash,
			Nonce:    1,
			Gas:      gas,
			GasPrice: gasPrice,
		}
		evmService.staticTransactionByHash = func(ctx context.Context, hash evm.Hash) (*evm.Transaction, error) {
			require.Equal(t, txHash, hash)
			return expTx, nil

		}

		got, err := client.TransactionByHash(ctx, txHash)
		require.NoError(t, err)
		require.Equal(t, expTx, got)
	})

	t.Run("LatestAndFinalizedHead", func(t *testing.T) {
		expHead := evm.Head{
			Hash:       blockHash,
			ParentHash: parentHash,
			Number:     blockNum,
			Timestamp:  10,
		}
		evmService.staticLatestAndFinalizedHead = func(ctx context.Context) (latest evm.Head, finalized evm.Head, err error) {
			return expHead, expHead, nil
		}

		got1, _, err := client.LatestAndFinalizedHead(ctx)
		require.NoError(t, err)
		require.Equal(t, expHead, got1)
	})

	t.Run("QueryTrackedLogs", func(t *testing.T) {
		expQuery := generateFixtureQuery()
		expLimitAndSort := query.NewLimitAndSort(query.CountLimit(10), query.SortByTimestamp{})
		expConfidence := primitives.Finalized
		expLog := []*evm.Log{
			{
				LogIndex:    2,
				BlockHash:   blockHash,
				BlockNumber: blockNum,
				Topics:      []evm.Hash{topic, topic2},
				EventSig:    eventSigHash,
				Address:     address,
				TxHash:      txHash,
				Data:        abi,
				Removed:     false,
			},
		}

		evmService.staticQueryTrackedLogs = func(ctx context.Context, filterQuery []query.Expression, limitAndSort query.LimitAndSort,
			confidenceLevel primitives.ConfidenceLevel) ([]*evm.Log, error) {
			require.Equal(t, expQuery, filterQuery)
			require.Equal(t, expLimitAndSort, limitAndSort)
			require.Equal(t, expConfidence, confidenceLevel)
			return expLog, nil
		}

		got, err := client.QueryTrackedLogs(ctx, expQuery, expLimitAndSort, expConfidence)
		require.NoError(t, err)
		require.Equal(t, expLog, got)

	})
}

type staticEVMService struct {
	staticCallContract           func(ctx context.Context, msg *evm.CallMsg, blockNumber *big.Int) ([]byte, error)
	staticFilterLogs             func(ctx context.Context, filterQuery evm.FilterQuery) ([]*evm.Log, error)
	staticBalanceAt              func(ctx context.Context, account evm.Address, blockNumber *big.Int) (*big.Int, error)
	staticEstimateGas            func(ctx context.Context, call *evm.CallMsg) (uint64, error)
	staticTransactionByHash      func(ctx context.Context, hash evm.Hash) (*evm.Transaction, error)
	staticTransactionReceipt     func(ctx context.Context, txHash evm.Hash) (*evm.Receipt, error)
	staticGetTransactionFee      func(ctx context.Context, transactionID types.IdempotencyKey) (*evm.TransactionFee, error)
	staticQueryTrackedLogs       func(ctx context.Context, filterQuery []query.Expression, limitAndSort query.LimitAndSort, confidenceLevel primitives.ConfidenceLevel) ([]*evm.Log, error)
	staticLatestAndFinalizedHead func(ctx context.Context) (latest evm.Head, finalized evm.Head, err error)
	staticRegisterLogTracking    func(ctx context.Context, filter evm.LPFilterQuery) error
	staticUnregisterLogTracking  func(ctx context.Context, filterName string) error
	staticGetTransactionStatus   func(ctx context.Context, transactionID types.IdempotencyKey) (types.TransactionStatus, error)
}

func (s *staticEVMService) CallContract(ctx context.Context, msg *evm.CallMsg, blockNumber *big.Int) ([]byte, error) {
	return s.staticCallContract(ctx, msg, blockNumber)
}

func (s *staticEVMService) FilterLogs(ctx context.Context, filterQuery evm.FilterQuery) ([]*evm.Log, error) {
	return s.staticFilterLogs(ctx, filterQuery)
}

func (s *staticEVMService) BalanceAt(ctx context.Context, account evm.Address, blockNumber *big.Int) (*big.Int, error) {
	return s.staticBalanceAt(ctx, account, blockNumber)
}

func (s *staticEVMService) EstimateGas(ctx context.Context, call *evm.CallMsg) (uint64, error) {
	return s.staticEstimateGas(ctx, call)
}

func (s *staticEVMService) TransactionByHash(ctx context.Context, hash evm.Hash) (*evm.Transaction, error) {
	return s.staticTransactionByHash(ctx, hash)
}

func (s *staticEVMService) TransactionReceipt(ctx context.Context, txHash evm.Hash) (*evm.Receipt, error) {
	return s.staticTransactionReceipt(ctx, txHash)
}

func (s *staticEVMService) GetTransactionFee(ctx context.Context, transactionID types.IdempotencyKey) (*evm.TransactionFee, error) {
	return s.staticGetTransactionFee(ctx, transactionID)
}

func (s *staticEVMService) QueryTrackedLogs(ctx context.Context, filterQuery []query.Expression, limitAndSort query.LimitAndSort, confidenceLevel primitives.ConfidenceLevel) ([]*evm.Log, error) {
	return s.staticQueryTrackedLogs(ctx, filterQuery, limitAndSort, confidenceLevel)
}

func (s *staticEVMService) LatestAndFinalizedHead(ctx context.Context) (evm.Head, evm.Head, error) {
	return s.staticLatestAndFinalizedHead(ctx)
}

func (s *staticEVMService) RegisterLogTracking(ctx context.Context, filter evm.LPFilterQuery) error {
	return s.staticRegisterLogTracking(ctx, filter)
}

func (s *staticEVMService) UnregisterLogTracking(ctx context.Context, filterName string) error {
	return s.staticUnregisterLogTracking(ctx, filterName)
}

func (s *staticEVMService) GetTransactionStatus(ctx context.Context, transactionID types.IdempotencyKey) (types.TransactionStatus, error) {
	return s.staticGetTransactionStatus(ctx, transactionID)
}

func generateFixtureQuery() []query.Expression {
	exprs := make([]query.Expression, 0)

	confirmationsValues := []primitives.ConfidenceLevel{primitives.Finalized, primitives.Unconfirmed}
	operatorValues := []primitives.ComparisonOperator{primitives.Eq, primitives.Neq, primitives.Gt, primitives.Lt, primitives.Gte, primitives.Lte}

	primitiveExpressions := []query.Expression{query.TxHash("txHash")}
	values := []evm.Hash{topic3, topic2}
	for _, op := range operatorValues {
		primitiveExpressions = append(primitiveExpressions, query.Block("123", op))
		primitiveExpressions = append(primitiveExpressions, query.Timestamp(123, op))
		primitiveExpressions = append(primitiveExpressions, evmprimitives.NewAddressFilter(address))
		primitiveExpressions = append(primitiveExpressions, evmprimitives.NewEventSigFilter(topic2))
		primitiveExpressions = append(primitiveExpressions, evmprimitives.NewEventByWordFilter(10, []evmprimitives.HashedValueComparator{{
			Values:   values,
			Operator: op,
		}}))
		primitiveExpressions = append(primitiveExpressions, evmprimitives.NewEventByTopicFilter(10, []evmprimitives.HashedValueComparator{{
			Values:   values,
			Operator: op,
		}}))
	}

	for _, conf := range confirmationsValues {
		primitiveExpressions = append(primitiveExpressions, query.Confidence(conf))
	}
	exprs = append(exprs, primitiveExpressions...)

	andOverPrimitivesBoolExpr := query.And(primitiveExpressions...)
	orOverPrimitivesBoolExpr := query.Or(primitiveExpressions...)

	nestedBoolExpr := query.And(
		query.TxHash("txHash"),
		andOverPrimitivesBoolExpr,
		orOverPrimitivesBoolExpr,
		query.TxHash("txHash"),
	)

	exprs = append(exprs, nestedBoolExpr)
	exprs = append(exprs, andOverPrimitivesBoolExpr)
	exprs = append(exprs, orOverPrimitivesBoolExpr)

	return exprs
}
