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

	evmpb "github.com/smartcontractkit/chainlink-common/pkg/chains/evm"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/chains/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
	evmprimitives "github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives/evm"
)

var (
	txId             = "txid1"
	txIndex          = 10
	txFee            = big.NewInt(12345)
	balance          = big.NewInt(1222345)
	gasPrice         = big.NewInt(12344)
	abi              = []byte("data")
	respAbi          = []byte("response")
	address          = evm.Address{1, 2, 3}
	address1         = evm.Address{10, 11, 14}
	address2         = evm.Address{13, 15, 16}
	blockHash        = evm.Hash{22, 33, 44}
	parentHash       = evm.Hash{01, 33, 44}
	fromBlock        = big.NewInt(10)
	blockNum         = big.NewInt(101)
	toBlock          = big.NewInt(145)
	topic            = evm.Hash{21, 3, 4}
	topic2           = evm.Hash{33, 1, 33}
	topic3           = evm.Hash{20, 19, 17}
	gas              = uint64(10)
	txHash           = evm.Hash{5, 3, 44}
	eventSigHash     = evm.Hash{14, 16, 29}
	filterName       = "f name 1"
	retention        = time.Second
	medianPluginType = string(types.Median)
	confidenceLevel  = primitives.Finalized
)

func Test_EVMDomainRoundTripThroughGRPC(t *testing.T) {
	t.Parallel()

	lis := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer()
	evmService := &staticEVMService{}
	evmpb.RegisterEVMServer(s, &evmServer{impl: evmService})

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
		evmService.staticBalanceAt = func(ctx context.Context, account evm.Address, blockNumber *big.Int, conf primitives.ConfidenceLevel) (*big.Int, error) {
			require.Equal(t, account, address)
			require.Equal(t, blockNumber, blockNum)
			require.Equal(t, conf, confidenceLevel)
			return balance, nil
		}

		resp, err := client.BalanceAtWithConfidence(ctx, address, blockNum, confidenceLevel)
		require.NoError(t, err)
		require.Equal(t, resp, balance)
	})

	t.Run("CallContract", func(t *testing.T) {
		expMsg := &evm.CallMsg{
			To:   address,
			From: address1,
			Data: abi,
		}
		evmService.staticCallContract = func(ctx context.Context, msg *evm.CallMsg, blockNumber *big.Int, conf primitives.ConfidenceLevel) ([]byte, error) {
			require.Equal(t, expMsg, msg)
			require.Equal(t, blockNumber, blockNum)
			require.Equal(t, conf, confidenceLevel)
			return respAbi, nil
		}

		resp, err := client.CallContractWithConfidence(ctx, expMsg, blockNum, confidenceLevel)
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
		evmService.staticFilterLogs = func(ctx context.Context, fq evm.FilterQuery, conf primitives.ConfidenceLevel) ([]*evm.Log, error) {
			require.Equal(t, expFQ, fq)
			require.Equal(t, conf, confidenceLevel)
			return expLog, nil
		}

		logs, err := client.FilterLogsWithConfidence(ctx, expFQ, confidenceLevel)
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

	t.Run("GetTransactionReceipt", func(t *testing.T) {
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
		evmService.staticGetTransactionReceipt = func(ctx context.Context, got evm.Hash) (*evm.Receipt, error) {
			require.Equal(t, txHash, got)
			return expReceipt, nil
		}

		got, err := client.GetTransactionReceipt(ctx, txHash)
		require.NoError(t, err)
		require.Equal(t, expReceipt, got)

	})

	t.Run("GetTransactionByHash", func(t *testing.T) {
		expTx := &evm.Transaction{
			To:       address,
			Data:     abi,
			Hash:     txHash,
			Nonce:    1,
			Gas:      gas,
			GasPrice: gasPrice,
		}
		evmService.staticGetTransactionByHash = func(ctx context.Context, hash evm.Hash) (*evm.Transaction, error) {
			require.Equal(t, txHash, hash)
			return expTx, nil

		}

		got, err := client.GetTransactionByHash(ctx, txHash)
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
		evmService.staticHeaderByNumber = func(ctx context.Context, blockNumber *big.Int, conf primitives.ConfidenceLevel) (evm.Head, error) {
			require.Equal(t, blockNum, blockNumber)
			require.Equal(t, conf, confidenceLevel)
			return expHead, nil
		}

		h, err := client.HeaderByNumberWithConfidence(ctx, blockNum, confidenceLevel)
		require.NoError(t, err)
		require.Equal(t, expHead, h)
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

	t.Run("GetForwarderForEOA", func(t *testing.T) {
		evmService.staticGetForwarderForEOA = func(ctx context.Context, eoa, ocr2AggregatorID evm.Address, pluginType string) (evm.Address, error) {
			require.Equal(t, address, eoa)
			require.Equal(t, address2, ocr2AggregatorID)
			require.Equal(t, pluginType, medianPluginType)
			return address1, nil
		}
		got, err := client.GetForwarderForEOA(ctx, address, address2, medianPluginType)
		require.NoError(t, err)
		require.Equal(t, address1, got)
	})
}

type staticEVMService struct {
	staticCallContract            func(ctx context.Context, msg *evm.CallMsg, blockNumber *big.Int, confidenceLevel primitives.ConfidenceLevel) ([]byte, error)
	staticFilterLogs              func(ctx context.Context, filterQuery evm.FilterQuery, confidenceLevel primitives.ConfidenceLevel) ([]*evm.Log, error)
	staticBalanceAt               func(ctx context.Context, account evm.Address, blockNumber *big.Int, confidenceLevel primitives.ConfidenceLevel) (*big.Int, error)
	staticEstimateGas             func(ctx context.Context, call *evm.CallMsg) (uint64, error)
	staticGetTransactionByHash    func(ctx context.Context, hash evm.Hash) (*evm.Transaction, error)
	staticGetTransactionReceipt   func(ctx context.Context, txHash evm.Hash) (*evm.Receipt, error)
	staticGetTransactionFee       func(ctx context.Context, transactionID types.IdempotencyKey) (*evm.TransactionFee, error)
	staticQueryTrackedLogs        func(ctx context.Context, filterQuery []query.Expression, limitAndSort query.LimitAndSort, confidenceLevel primitives.ConfidenceLevel) ([]*evm.Log, error)
	staticHeaderByNumber          func(ctx context.Context, blockNumber *big.Int, confidenceLevel primitives.ConfidenceLevel) (evm.Head, error)
	staticRegisterLogTracking     func(ctx context.Context, filter evm.LPFilterQuery) error
	staticUnregisterLogTracking   func(ctx context.Context, filterName string) error
	staticGetTransactionStatus    func(ctx context.Context, transactionID types.IdempotencyKey) (types.TransactionStatus, error)
	staticSubmitTransaction       func(ctx context.Context, submitTransactionRequest evm.SubmitTransactionRequest) (*evm.TransactionResult, error)
	staticCalculateTransactionFee func(ctx context.Context, gasInfo evm.ReceiptGasInfo) (*evm.TransactionFee, error)
	staticGetForwarderForEOA      func(ctx context.Context, eoa, ocr2AggregatorID evm.Address, pluginType string) (forwarder evm.Address, err error)
}

func (s *staticEVMService) CallContractWithConfidence(ctx context.Context, msg *evm.CallMsg, blockNumber *big.Int, conf primitives.ConfidenceLevel) ([]byte, error) {
	return s.staticCallContract(ctx, msg, blockNumber, conf)
}

func (s *staticEVMService) FilterLogsWithConfidence(ctx context.Context, filterQuery evm.FilterQuery, conf primitives.ConfidenceLevel) ([]*evm.Log, error) {
	return s.staticFilterLogs(ctx, filterQuery, conf)
}

func (s *staticEVMService) BalanceAtWithConfidence(ctx context.Context, account evm.Address, blockNumber *big.Int, conf primitives.ConfidenceLevel) (*big.Int, error) {
	return s.staticBalanceAt(ctx, account, blockNumber, conf)
}

func (s *staticEVMService) EstimateGas(ctx context.Context, call *evm.CallMsg) (uint64, error) {
	return s.staticEstimateGas(ctx, call)
}

func (s *staticEVMService) GetTransactionByHash(ctx context.Context, hash evm.Hash) (*evm.Transaction, error) {
	return s.staticGetTransactionByHash(ctx, hash)
}

func (s *staticEVMService) GetTransactionReceipt(ctx context.Context, txHash evm.Hash) (*evm.Receipt, error) {
	return s.staticGetTransactionReceipt(ctx, txHash)
}

func (s *staticEVMService) GetTransactionFee(ctx context.Context, transactionID types.IdempotencyKey) (*evm.TransactionFee, error) {
	return s.staticGetTransactionFee(ctx, transactionID)
}

func (s *staticEVMService) QueryTrackedLogs(ctx context.Context, filterQuery []query.Expression, limitAndSort query.LimitAndSort, confidenceLevel primitives.ConfidenceLevel) ([]*evm.Log, error) {
	return s.staticQueryTrackedLogs(ctx, filterQuery, limitAndSort, confidenceLevel)
}

func (s *staticEVMService) HeaderByNumberWithConfidence(ctx context.Context, blockNumber *big.Int, confidenceLevel primitives.ConfidenceLevel) (evm.Head, error) {
	return s.staticHeaderByNumber(ctx, blockNumber, confidenceLevel)
}

func (s *staticEVMService) RegisterLogTracking(ctx context.Context, filter evm.LPFilterQuery) error {
	return s.staticRegisterLogTracking(ctx, filter)
}

func (s *staticEVMService) GetTransactionStatus(ctx context.Context, transactionID types.IdempotencyKey) (types.TransactionStatus, error) {
	return s.staticGetTransactionStatus(ctx, transactionID)
}

func (s *staticEVMService) CalculateTransactionFee(ctx context.Context, gasInfo evm.ReceiptGasInfo) (*evm.TransactionFee, error) {
	return s.staticCalculateTransactionFee(ctx, gasInfo)
}

func (s *staticEVMService) SubmitTransaction(ctx context.Context, submitTransactionRequest evm.SubmitTransactionRequest) (*evm.TransactionResult, error) {
	return s.staticSubmitTransaction(ctx, submitTransactionRequest)
}

func (s *staticEVMService) UnregisterLogTracking(ctx context.Context, filterName string) error {
	return s.staticUnregisterLogTracking(ctx, filterName)
}

func (s *staticEVMService) GetForwarderForEOA(ctx context.Context, eoa, ocr2AggregatorID evm.Address, pluginType string) (forwarder evm.Address, err error) {
	return s.staticGetForwarderForEOA(ctx, eoa, ocr2AggregatorID, pluginType)
}

func generateFixtureQuery() []query.Expression {
	exprs := make([]query.Expression, 0)

	confirmationsValues := []primitives.ConfidenceLevel{primitives.Finalized, primitives.Unconfirmed, primitives.Safe}
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
