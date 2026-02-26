package relayerset

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"math/big"
	"testing"
	"time"

	"github.com/hashicorp/go-plugin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/relayerset"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/chains/aptos"
	evmtypes "github.com/smartcontractkit/chainlink-common/pkg/types/chains/evm"
	soltypes "github.com/smartcontractkit/chainlink-common/pkg/types/chains/solana"
	"github.com/smartcontractkit/chainlink-common/pkg/types/chains/ton"
	tontypes "github.com/smartcontractkit/chainlink-common/pkg/types/chains/ton"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core/mocks"
	mocks2 "github.com/smartcontractkit/chainlink-common/pkg/types/mocks"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
	evmprimitives "github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives/evm"
	solprimitives "github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives/solana"
)

func Test_RelayerSet(t *testing.T) {
	ctx := t.Context()
	stopCh := make(chan struct{})
	log := logger.Test(t)

	relayer1 := mocks.NewRelayer(t)
	relayer2 := mocks.NewRelayer(t)

	relayers := map[types.RelayID]core.Relayer{
		{
			Network: "N1",
			ChainID: "C1",
		}: relayer1,
		{
			Network: "N2",
			ChainID: "C2",
		}: relayer2,
	}

	pluginName := "relayerset-test"
	client, server := plugin.TestPluginGRPCConn(
		t,
		true,
		map[string]plugin.Plugin{
			pluginName: &testRelaySetPlugin{
				log:  log,
				impl: &TestRelayerSet{relayers: relayers},
				brokerExt: &net.BrokerExt{
					BrokerConfig: net.BrokerConfig{
						StopCh: stopCh,
						Logger: log,
					},
				},
			},
		},
	)

	defer client.Close()
	defer server.Stop()

	relayerSetClient, err := client.Dispense(pluginName)
	require.NoError(t, err)

	rc, ok := relayerSetClient.(*Client)
	require.True(t, ok)

	relayerClient, err := rc.Get(ctx, types.RelayID{
		Network: "N1",
		ChainID: "C1",
	})

	require.NoError(t, err)

	relayer1.On("Start", mock.Anything).Return(nil)
	err = relayerClient.Start(ctx)
	require.NoError(t, err)
	relayer1.AssertCalled(t, "Start", mock.Anything)

	relayer1.On("Close").Return(nil)
	err = relayerClient.Close()
	require.NoError(t, err)
	relayer1.AssertCalled(t, "Close")

	relayer1.On("Ready").Return(nil)
	err = relayerClient.Ready()
	require.NoError(t, err)
	relayer1.AssertCalled(t, "Ready")

	relayer1.On("HealthReport").Return(map[string]error{"stat1": errors.New("error1")})
	healthReport := relayerClient.HealthReport()
	require.Len(t, healthReport, 1)
	require.Equal(t, "error1", healthReport["stat1"].Error())
	relayer1.AssertCalled(t, "HealthReport")

	chainInfo := types.ChainInfo{
		FamilyName:      "familyName",
		ChainID:         "123",
		NetworkName:     "someNetwork",
		NetworkNameFull: "someNetwork-full",
	}
	relayer1.On("GetChainInfo", mock.Anything).Return(chainInfo, nil)
	chainInfoReply, err := relayerClient.GetChainInfo(ctx)
	require.NoError(t, err)
	require.Equal(t, chainInfo, chainInfoReply)

	relayer1.On("Name").Return("test-relayer")
	name := relayerClient.Name()
	require.Equal(t, "test-relayer", name)
	relayer1.AssertCalled(t, "Name")
}

func Test_RelayerSet_ContractReader(t *testing.T) {
	ctx := t.Context()
	stopCh := make(chan struct{})
	log := logger.Test(t)

	relayer1 := mocks.NewRelayer(t)
	relayer2 := mocks.NewRelayer(t)

	relayers := map[types.RelayID]core.Relayer{
		{
			Network: "N1",
			ChainID: "C1",
		}: relayer1,
		{
			Network: "N2",
			ChainID: "C2",
		}: relayer2,
	}

	pluginName := "relayerset-test"
	client, server := plugin.TestPluginGRPCConn(
		t,
		true,
		map[string]plugin.Plugin{
			pluginName: &testRelaySetPlugin{
				log:  log,
				impl: &TestRelayerSet{relayers: relayers},
				brokerExt: &net.BrokerExt{
					BrokerConfig: net.BrokerConfig{
						StopCh: stopCh,
						Logger: log,
					},
				},
			},
		},
	)

	defer client.Close()
	defer server.Stop()

	relayerSetClient, err := client.Dispense(pluginName)
	require.NoError(t, err)

	rc, ok := relayerSetClient.(*Client)
	require.True(t, ok)

	retrievedRelayer, err := rc.Get(ctx, types.RelayID{
		Network: "N1",
		ChainID: "C1",
	})

	require.NoError(t, err)

	reader := &TestContractReader{mockedContractReader: mocks2.NewContractReader(t)}
	relayer1.On("NewContractReader", mock.Anything, mock.Anything).Return(reader, nil)

	fetchedReader, err := retrievedRelayer.NewContractReader(ctx, []byte("config"))
	require.NoError(t, err)

	reader.mockedContractReader.EXPECT().Start(mock.Anything).Return(nil)
	err = fetchedReader.Start(ctx)
	require.NoError(t, err)

	returnVal := map[any]any{}
	reader.mockedContractReader.EXPECT().GetLatestValue(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	err = fetchedReader.GetLatestValue(ctx, "readIdentifier", primitives.Finalized, nil, &returnVal)
	require.NoError(t, err)

	reader.mockedContractReader.EXPECT().GetLatestValueWithHeadData(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
	_, err = fetchedReader.GetLatestValueWithHeadData(ctx, "readIdentifier", primitives.Finalized, nil, &returnVal)
	require.NoError(t, err)

	reader.mockedContractReader.EXPECT().QueryKey(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
	_, err = fetchedReader.QueryKey(ctx, types.BoundContract{}, query.KeyFilter{}, query.LimitAndSort{}, &returnVal)
	require.NoError(t, err)

	reader.mockedContractReader.EXPECT().QueryKeys(mock.Anything, mock.Anything, mock.Anything).Return(func(yield func(string, types.Sequence) bool) {}, nil)
	_, err = fetchedReader.QueryKeys(ctx, []types.ContractKeyFilter{}, query.LimitAndSort{})
	require.NoError(t, err)

	reader.mockedContractReader.EXPECT().Close().Return(nil)
	err = fetchedReader.Close()
	require.NoError(t, err)

	reader.mockedContractReader.EXPECT().GetLatestValue(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	err = fetchedReader.GetLatestValue(ctx, "readIdentifier", primitives.Finalized, nil, &returnVal)
	require.ErrorContains(t, err, "contract reader not found")
}

var (
	address  = evmtypes.Address{1, 2, 3}
	address1 = evmtypes.Address{10, 11, 14}
	topic    = evmtypes.Hash{21, 3, 4}
	topic2   = evmtypes.Hash{33, 1, 33}
	topic3   = evmtypes.Hash{20, 19, 17}
	txHash   = evmtypes.Hash{5, 3, 44}

	msg = evmtypes.CallMsg{
		From: address,
		To:   address1,
		Data: []byte("data"),
	}

	evmLog = evmtypes.Log{
		Address:     address,
		Topics:      [][32]byte{topic, topic2},
		Data:        []byte("data"),
		BlockNumber: big.NewInt(101),
		TxHash:      txHash,
	}
)

func Test_RelayerSet_EVMService(t *testing.T) {
	ctx := t.Context()
	stopCh := make(chan struct{})
	log := logger.Test(t)

	relayer1 := mocks.NewRelayer(t)
	relayers := map[types.RelayID]core.Relayer{
		{Network: "N1", ChainID: "C1"}: relayer1,
	}

	pluginName := "evm-relayerset-test"
	client, server := plugin.TestPluginGRPCConn(
		t,
		true,
		map[string]plugin.Plugin{
			pluginName: &testRelaySetPlugin{
				log:  log,
				impl: &TestRelayerSet{relayers: relayers},
				brokerExt: &net.BrokerExt{
					BrokerConfig: net.BrokerConfig{
						StopCh: stopCh,
						Logger: log,
					},
				},
			},
		},
	)
	defer client.Close()
	defer server.Stop()

	relayerSetClient, err := client.Dispense(pluginName)
	require.NoError(t, err)
	rc, ok := relayerSetClient.(*Client)
	require.True(t, ok)

	retrievedRelayer, err := rc.Get(ctx, types.RelayID{Network: "N1", ChainID: "C1"})
	require.NoError(t, err)

	zero := evmtypes.Address{}
	tests := []struct {
		name string
		run  func(t *testing.T, evm types.EVMService, mockEVM *mocks2.EVMService)
	}{
		{
			name: "CallContract",
			run: func(t *testing.T, evm types.EVMService, mockEVM *mocks2.EVMService) {
				block := big.NewInt(100)
				conf := primitives.Finalized
				mockEVM.EXPECT().CallContract(mock.Anything, evmtypes.CallContractRequest{
					Msg:             &msg,
					BlockNumber:     block,
					ConfidenceLevel: conf,
				}).Return(&evmtypes.CallContractReply{
					Data: []byte("ok"),
				}, nil)
				reply, err := evm.CallContract(ctx, evmtypes.CallContractRequest{
					Msg:             &msg,
					BlockNumber:     block,
					ConfidenceLevel: conf,
				})
				require.NoError(t, err)
				require.Equal(t, []byte("ok"), reply.Data)
			},
		},
		{
			name: "FilterLogs",
			run: func(t *testing.T, evm types.EVMService, mockEVM *mocks2.EVMService) {
				filter := evmtypes.FilterQuery{
					Addresses: []evmtypes.Address{address, address1},
					FromBlock: big.NewInt(10),
					ToBlock:   big.NewInt(145),
					Topics:    [][][32]byte{{topic, topic2}, {topic3}},
				}
				conf := primitives.Finalized
				mockEVM.EXPECT().FilterLogs(mock.Anything, evmtypes.FilterLogsRequest{
					FilterQuery:     filter,
					ConfidenceLevel: conf,
				}).Return(&evmtypes.FilterLogsReply{Logs: []*evmtypes.Log{&evmLog}}, nil)

				reply, err := evm.FilterLogs(ctx, evmtypes.FilterLogsRequest{
					FilterQuery:     filter,
					ConfidenceLevel: conf,
				})
				require.NoError(t, err)
				require.Len(t, reply.Logs, 1)
				require.Equal(t, &evmLog, reply.Logs[0])
			},
		},
		{
			name: "BalanceAt",
			run: func(t *testing.T, evm types.EVMService, mockEVM *mocks2.EVMService) {
				addr := evmtypes.Address{0xbb}
				conf := primitives.Finalized
				mockEVM.EXPECT().BalanceAt(mock.Anything, evmtypes.BalanceAtRequest{
					Address:         addr,
					BlockNumber:     big.NewInt(200),
					ConfidenceLevel: conf,
				}).Return(&evmtypes.BalanceAtReply{Balance: big.NewInt(999)}, nil)
				reply, err := evm.BalanceAt(ctx, evmtypes.BalanceAtRequest{
					Address:         addr,
					BlockNumber:     big.NewInt(200),
					ConfidenceLevel: conf,
				})
				require.NoError(t, err)
				require.Equal(t, big.NewInt(999), reply.Balance)
			},
		},
		{
			name: "BalanceAt_ZeroAddress_NoError",
			run: func(t *testing.T, evm types.EVMService, mockEVM *mocks2.EVMService) {
				conf := primitives.Finalized
				mockEVM.EXPECT().BalanceAt(mock.Anything, evmtypes.BalanceAtRequest{
					Address:         zero,
					BlockNumber:     big.NewInt(0),
					ConfidenceLevel: conf,
				}).Return(&evmtypes.BalanceAtReply{Balance: big.NewInt(0)}, nil)

				reply, err := evm.BalanceAt(ctx, evmtypes.BalanceAtRequest{
					Address:         zero,
					BlockNumber:     big.NewInt(0),
					ConfidenceLevel: conf,
				})
				require.NoError(t, err)
				require.Equal(t, big.NewInt(0), reply.Balance)
			},
		},
		{
			name: "EstimateGas",
			run: func(t *testing.T, evm types.EVMService, mockEVM *mocks2.EVMService) {
				mockEVM.EXPECT().EstimateGas(mock.Anything, &msg).Return(uint64(42000), nil)
				out, err := evm.EstimateGas(ctx, &msg)
				require.NoError(t, err)
				require.Equal(t, uint64(42000), out)
			},
		},
		{
			name: "GetTransactionByHash",
			run: func(t *testing.T, evm types.EVMService, mockEVM *mocks2.EVMService) {
				tx := evmtypes.Transaction{
					To:       address,
					Data:     []byte("data"),
					Hash:     txHash,
					Nonce:    42,
					Gas:      24,
					GasPrice: big.NewInt(100),
					Value:    big.NewInt(300),
				}

				mockEVM.EXPECT().GetTransactionByHash(mock.Anything, evmtypes.GetTransactionByHashRequest{Hash: txHash}).Return(&tx, nil)
				out, err := evm.GetTransactionByHash(ctx, evmtypes.GetTransactionByHashRequest{Hash: txHash})
				require.NoError(t, err)
				require.Equal(t, tx, *out)
			},
		},
		{
			name: "GetTransactionReceipt",
			run: func(t *testing.T, evm types.EVMService, mockEVM *mocks2.EVMService) {
				receipt := evmtypes.Receipt{
					TxHash:            txHash,
					Logs:              []*evmtypes.Log{&evmLog},
					Status:            1,
					ContractAddress:   address1,
					GasUsed:           uint64(10),
					BlockHash:         evmtypes.Hash{22, 33, 44},
					BlockNumber:       big.NewInt(101),
					TransactionIndex:  uint64(10),
					EffectiveGasPrice: big.NewInt(12344),
				}
				mockEVM.EXPECT().GetTransactionReceipt(mock.Anything, evmtypes.GeTransactionReceiptRequest{Hash: txHash}).Return(&receipt, nil)
				out, err := evm.GetTransactionReceipt(ctx, evmtypes.GeTransactionReceiptRequest{Hash: txHash})
				require.NoError(t, err)
				require.Equal(t, receipt.TxHash, out.TxHash)
				require.Equal(t, receipt.Status, out.Status)
				require.Equal(t, receipt.ContractAddress, out.ContractAddress)
				require.Equal(t, receipt.GasUsed, out.GasUsed)
				require.Equal(t, receipt.BlockHash, out.BlockHash)
				require.Equal(t, receipt.BlockNumber, out.BlockNumber)
				require.Equal(t, receipt.TransactionIndex, out.TransactionIndex)
				require.Equal(t, receipt.EffectiveGasPrice, out.EffectiveGasPrice)
				require.Len(t, out.Logs, len(receipt.Logs))
			},
		},
		{
			name: "RegisterLogTracking",
			run: func(t *testing.T, evm types.EVMService, mockEVM *mocks2.EVMService) {
				lpFilterQuery := evmtypes.LPFilterQuery{
					Name:         "f name 1",
					Addresses:    [][20]byte{address, address1},
					EventSigs:    [][32]byte{{14, 16, 29}},
					Topic2:       [][32]byte{topic2},
					Topic3:       [][32]byte{topic3},
					Topic4:       [][32]byte{{20, 18, 14}},
					Retention:    time.Minute,
					MaxLogsKept:  uint64(10),
					LogsPerBlock: uint64(20),
				}

				mockEVM.EXPECT().RegisterLogTracking(mock.Anything, lpFilterQuery).Return(nil)
				require.NoError(t, evm.RegisterLogTracking(ctx, lpFilterQuery))
			},
		},
		{
			name: "UnregisterLogTracking",
			run: func(t *testing.T, evm types.EVMService, mockEVM *mocks2.EVMService) {
				mockEVM.EXPECT().UnregisterLogTracking(mock.Anything, "logs").Return(nil)
				require.NoError(t, evm.UnregisterLogTracking(ctx, "logs"))
			},
		},
		{
			name: "HeaderByNumber",
			run: func(t *testing.T, evm types.EVMService, mockEVM *mocks2.EVMService) {
				head1 := evmtypes.Header{Number: big.NewInt(123)}
				blockNumber := big.NewInt(123)
				conf := primitives.Finalized
				mockEVM.EXPECT().HeaderByNumber(mock.Anything, evmtypes.HeaderByNumberRequest{
					Number:          blockNumber,
					ConfidenceLevel: conf,
				}).Return(&evmtypes.HeaderByNumberReply{Header: &head1}, nil)
				reply, err := evm.HeaderByNumber(ctx, evmtypes.HeaderByNumberRequest{
					Number:          blockNumber,
					ConfidenceLevel: conf,
				})
				require.NoError(t, err)
				require.Equal(t, &head1, reply.Header)
			},
		},
		{
			name: "GetTransactionFee",
			run: func(t *testing.T, evm types.EVMService, mockEVM *mocks2.EVMService) {
				id := types.IdempotencyKey("fee-tx")
				fee := &evmtypes.TransactionFee{TransactionFee: big.NewInt(1000)}
				mockEVM.EXPECT().GetTransactionFee(mock.Anything, id).Return(fee, nil)
				out, err := evm.GetTransactionFee(ctx, id)
				require.NoError(t, err)
				require.Equal(t, fee, out)
			},
		},
		{
			name: "GetTransactionStatus",
			run: func(t *testing.T, evm types.EVMService, mockEVM *mocks2.EVMService) {
				id := types.IdempotencyKey("status-tx")
				mockEVM.EXPECT().GetTransactionStatus(mock.Anything, id).Return(types.Unconfirmed, nil)
				out, err := evm.GetTransactionStatus(ctx, id)
				require.NoError(t, err)
				require.Equal(t, types.Unconfirmed, out)
			},
		},
		{
			name: "QueryTrackedLogs",
			run: func(t *testing.T, evm types.EVMService, mockEVM *mocks2.EVMService) {
				fq := generateFixtureQuery()
				expLimitAndSort := query.NewLimitAndSort(query.CountLimit(10), query.SortByTimestamp{})
				expConfidence := primitives.Finalized
				mockEVM.EXPECT().QueryTrackedLogs(mock.Anything, fq, expLimitAndSort, expConfidence).Return([]*evmtypes.Log{&evmLog}, nil)
				out, err := evm.QueryTrackedLogs(ctx, fq, expLimitAndSort, expConfidence)
				require.NoError(t, err)
				require.Equal(t, &evmLog, out[0])
			},
		},
		{
			name: "SubmitTransaction",
			run: func(t *testing.T, evm types.EVMService, mockEVM *mocks2.EVMService) {
				txRequest := evmtypes.SubmitTransactionRequest{
					To:   address1,
					Data: []byte("data"),
				}
				expectedTxResult := evmtypes.TransactionResult{
					TxStatus: evmtypes.TxSuccess,
					TxHash:   evmtypes.Hash{1, 2, 3},
				}
				mockEVM.EXPECT().SubmitTransaction(mock.Anything, txRequest).Return(&expectedTxResult, nil)
				txResult, err := evm.SubmitTransaction(ctx, txRequest)
				require.NoError(t, err)
				require.Equal(t, &expectedTxResult, txResult)
			},
		},
		{
			name: "SubmitTransaction_ZeroTo_NoError",
			run: func(t *testing.T, evm types.EVMService, mockEVM *mocks2.EVMService) {
				txRequest := evmtypes.SubmitTransactionRequest{
					To:   zero,
					Data: []byte("data"),
				}
				expectedTxResult := evmtypes.TransactionResult{
					TxStatus: evmtypes.TxSuccess,
					TxHash:   evmtypes.Hash{9, 9, 9},
				}
				// Should not error even when To is zero address.
				mockEVM.EXPECT().SubmitTransaction(mock.Anything, txRequest).Return(&expectedTxResult, nil)
				txResult, err := evm.SubmitTransaction(ctx, txRequest)
				require.NoError(t, err)
				require.Equal(t, &expectedTxResult, txResult)
			},
		},
		{
			name: "CalculateTransactionFee",
			run: func(t *testing.T, evm types.EVMService, mockEVM *mocks2.EVMService) {
				gasInfo := evmtypes.ReceiptGasInfo{
					GasUsed:           1000,
					EffectiveGasPrice: big.NewInt(2000),
				}
				expectedFee := &evmtypes.TransactionFee{
					TransactionFee: big.NewInt(2000000),
				}
				mockEVM.EXPECT().CalculateTransactionFee(mock.Anything, gasInfo).Return(expectedFee, nil)
				fee, err := evm.CalculateTransactionFee(ctx, gasInfo)
				require.NoError(t, err)
				require.Equal(t, expectedFee, fee)
			},
		},
		{
			name: "GetForwarderForEOA_ErrorsOnZeroEoa",
			run: func(t *testing.T, evm types.EVMService, mockEVM *mocks2.EVMService) {
				mockEVM.EXPECT().GetForwarderForEOA(mock.Anything, zero, address, "str").Return(address1, nil)
				res, err := evm.GetForwarderForEOA(ctx, zero, address, "str")
				require.NoError(t, err)
				require.Equal(t, address1, res)
			},
		},
		{
			name: "GetForwarderForEOA_ErrorsOnZeroOCRID",
			run: func(t *testing.T, evm types.EVMService, mockEVM *mocks2.EVMService) {
				mockEVM.EXPECT().GetForwarderForEOA(mock.Anything, address, zero, "str").Return(address1, nil)
				res, err := evm.GetForwarderForEOA(ctx, address, zero, "str")
				require.NoError(t, err)
				require.Equal(t, address1, res)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockEVM := mocks2.NewEVMService(t)
			relayer1.On("EVM", mock.Anything, mock.Anything).Return(mockEVM, nil).Once()

			fetchedEVM, err := retrievedRelayer.EVM()
			require.NoError(t, err)

			tc.run(t, fetchedEVM, mockEVM)
		})
	}
}

func Test_RelayerSet_SolanaService(t *testing.T) {
	ctx := t.Context()
	stopCh := make(chan struct{})
	log := logger.Test(t)

	relayer1 := mocks.NewRelayer(t)
	relayers := map[types.RelayID]core.Relayer{
		{Network: "N1", ChainID: "C1"}: relayer1,
	}

	pluginName := "solana-relayerset-test"
	client, server := plugin.TestPluginGRPCConn(
		t,
		true,
		map[string]plugin.Plugin{
			pluginName: &testRelaySetPlugin{
				log:  log,
				impl: &TestRelayerSet{relayers: relayers},
				brokerExt: &net.BrokerExt{
					BrokerConfig: net.BrokerConfig{
						StopCh: stopCh,
						Logger: log,
					},
				},
			},
		},
	)
	defer client.Close()
	defer server.Stop()

	relayerSetClient, err := client.Dispense(pluginName)
	require.NoError(t, err)
	rc, ok := relayerSetClient.(*Client)
	require.True(t, ok)

	retrievedRelayer, err := rc.Get(ctx, types.RelayID{Network: "N1", ChainID: "C1"})
	require.NoError(t, err)

	tests := []struct {
		name string
		run  func(t *testing.T, sol types.SolanaService, mockSol *mocks2.SolanaService)
	}{
		{
			name: "GetBalance",
			run: func(t *testing.T, sol types.SolanaService, mockSol *mocks2.SolanaService) {
				address := soltypes.PublicKey{1, 2, 3}
				bal := uint64(32)
				mockSol.EXPECT().GetBalance(mock.Anything, soltypes.GetBalanceRequest{
					Addr:       address,
					Commitment: soltypes.CommitmentConfirmed,
				}).Return(&soltypes.GetBalanceReply{Value: bal}, nil)
				reply, err := sol.GetBalance(ctx, soltypes.GetBalanceRequest{
					Addr: address, Commitment: soltypes.CommitmentConfirmed,
				})
				require.NoError(t, err)
				require.Equal(t, bal, reply.Value)
			},
		},
		{
			name: "GetAccountInfoWithOpts",
			run: func(t *testing.T, sol types.SolanaService, mockSol *mocks2.SolanaService) {
				req := soltypes.GetAccountInfoRequest{
					Account: soltypes.PublicKey{9, 9, 9},
					Opts: &soltypes.GetAccountInfoOpts{
						Encoding:   soltypes.EncodingJSONParsed,
						Commitment: soltypes.CommitmentFinalized,
					},
				}
				slot := uint64(22)
				lamports := uint64(33)
				mockSol.EXPECT().
					GetAccountInfoWithOpts(mock.Anything, req).
					Return(&soltypes.GetAccountInfoReply{
						RPCContext: soltypes.RPCContext{Slot: slot},
						Value: &soltypes.Account{
							Lamports:   lamports,
							Executable: false,
						},
					}, nil)

				out, err := sol.GetAccountInfoWithOpts(ctx, req)
				require.NoError(t, err)
				require.Equal(t, lamports, out.Value.Lamports)
				require.Equal(t, slot, out.RPCContext.Slot)
			},
		},
		{
			name: "GetMultipleAccountsWithOpts",
			run: func(t *testing.T, sol types.SolanaService, mockSol *mocks2.SolanaService) {
				req := soltypes.GetMultipleAccountsRequest{
					Accounts: []soltypes.PublicKey{{1}, {2}},
					Opts: &soltypes.GetMultipleAccountsOpts{
						Encoding:   soltypes.EncodingBase64,
						Commitment: soltypes.CommitmentProcessed,
					},
				}
				slot := uint64(22)
				lamports := uint64(33)
				mockSol.EXPECT().
					GetMultipleAccountsWithOpts(mock.Anything, req).
					Return(&soltypes.GetMultipleAccountsReply{
						RPCContext: soltypes.RPCContext{Slot: slot},
						Value: []*soltypes.Account{
							{Lamports: lamports}, {Lamports: lamports},
						},
					}, nil)

				out, err := sol.GetMultipleAccountsWithOpts(ctx, req)
				require.NoError(t, err)
				require.Len(t, out.Value, 2)
				require.Equal(t, lamports, out.Value[0].Lamports)
			},
		},
		{
			name: "GetBlock",
			run: func(t *testing.T, sol types.SolanaService, mockSol *mocks2.SolanaService) {
				req := soltypes.GetBlockRequest{
					Slot: 42,
					Opts: &soltypes.GetBlockOpts{
						Commitment: soltypes.CommitmentConfirmed,
					},
				}
				expHash := soltypes.Hash{1, 2, 3}
				parSlot := uint64(33)
				expHeight := uint64(66)
				mockSol.EXPECT().
					GetBlock(mock.Anything, req).
					Return(&soltypes.GetBlockReply{
						Blockhash:   expHash,
						ParentSlot:  parSlot,
						BlockTime:   nil,
						BlockHeight: &expHeight,
					}, nil)

				out, err := sol.GetBlock(ctx, req)
				require.NoError(t, err)
				require.Equal(t, parSlot, out.ParentSlot)
				require.Equal(t, expHeight, *out.BlockHeight)
			},
		},
		{
			name: "GetSlotHeight",
			run: func(t *testing.T, sol types.SolanaService, mockSol *mocks2.SolanaService) {
				req := soltypes.GetSlotHeightRequest{Commitment: soltypes.CommitmentFinalized}
				mockSol.EXPECT().
					GetSlotHeight(mock.Anything, req).
					Return(&soltypes.GetSlotHeightReply{Height: 9090}, nil)

				out, err := sol.GetSlotHeight(ctx, req)
				require.NoError(t, err)
				require.Equal(t, uint64(9090), out.Height)
			},
		},
		{
			name: "GetTransaction",
			run: func(t *testing.T, sol types.SolanaService, mockSol *mocks2.SolanaService) {
				var sig soltypes.Signature
				copy(sig[:], []byte{1, 2, 3, 4})
				req := soltypes.GetTransactionRequest{Signature: sig}
				expTime := soltypes.UnixTimeSeconds(11)
				expFee := uint64(33)
				expSlot := uint64(17)
				mockSol.EXPECT().
					GetTransaction(mock.Anything, req).
					Return(&soltypes.GetTransactionReply{
						Slot:      expSlot,
						BlockTime: &expTime,
						Meta: &soltypes.TransactionMeta{
							Fee: expFee,
						},
					}, nil)

				out, err := sol.GetTransaction(ctx, req)
				require.NoError(t, err)
				require.Equal(t, expSlot, out.Slot)
				require.Equal(t, expFee, out.Meta.Fee)
			},
		},
		{
			name: "GetFeeForMessage",
			run: func(t *testing.T, sol types.SolanaService, mockSol *mocks2.SolanaService) {
				req := soltypes.GetFeeForMessageRequest{
					Message:    "AgAAAA==",
					Commitment: soltypes.CommitmentProcessed,
				}
				mockSol.EXPECT().
					GetFeeForMessage(mock.Anything, req).
					Return(&soltypes.GetFeeForMessageReply{Fee: 1234}, nil)

				out, err := sol.GetFeeForMessage(ctx, req)
				require.NoError(t, err)
				require.Equal(t, uint64(1234), out.Fee)
			},
		},
		{
			name: "GetSignatureStatuses",
			run: func(t *testing.T, sol types.SolanaService, mockSol *mocks2.SolanaService) {
				s1, s2 := soltypes.Signature{}, soltypes.Signature{}
				copy(s1[:], []byte{7})
				copy(s2[:], []byte{8})
				req := soltypes.GetSignatureStatusesRequest{Sigs: []soltypes.Signature{s1, s2}}
				expSlot1 := uint64(1)
				expConf1 := uint64(11)
				expSlot2 := uint64(33)
				expConf2 := uint64(22)
				mockSol.EXPECT().
					GetSignatureStatuses(mock.Anything, req).
					Return(&soltypes.GetSignatureStatusesReply{
						Results: []soltypes.GetSignatureStatusesResult{
							{Slot: uint64(expSlot1), Confirmations: &expConf1}, {Slot: expSlot2, Confirmations: &expConf2},
						},
					}, nil)

				out, err := sol.GetSignatureStatuses(ctx, req)
				require.NoError(t, err)
				require.Len(t, out.Results, 2)
				require.Equal(t, expSlot1, out.Results[0].Slot)
			},
		},
		{
			name: "SimulateTX",
			run: func(t *testing.T, sol types.SolanaService, mockSol *mocks2.SolanaService) {
				req := soltypes.SimulateTXRequest{
					Receiver:           soltypes.PublicKey{5, 5, 5},
					EncodedTransaction: "BASE64TX==",
					Opts: &soltypes.SimulateTXOpts{
						SigVerify:              true,
						Commitment:             soltypes.CommitmentProcessed,
						ReplaceRecentBlockhash: true,
					},
				}
				mockSol.EXPECT().
					SimulateTX(mock.Anything, req).
					Return(&soltypes.SimulateTXReply{
						Err:  "",
						Logs: []string{"log1", "log2"},
					}, nil)

				out, err := sol.SimulateTX(ctx, req)
				require.NoError(t, err)
				require.Len(t, out.Logs, 2)
			},
		},
		{
			name: "SubmitTransaction",
			run: func(t *testing.T, sol types.SolanaService, mockSol *mocks2.SolanaService) {
				units := uint32(1_400_999)
				price := uint64(11)
				req := soltypes.SubmitTransactionRequest{
					Cfg: &soltypes.ComputeConfig{
						ComputeLimit:    &units,
						ComputeMaxPrice: &price,
					},
					Receiver:           soltypes.PublicKey{3, 3, 3},
					EncodedTransaction: "BASE64TX==",
				}
				expected := &soltypes.SubmitTransactionReply{
					Signature:      soltypes.Signature{0xaa, 0xbb},
					IdempotencyKey: "idem-123",
					Status:         soltypes.TxSuccess,
				}
				mockSol.EXPECT().
					SubmitTransaction(mock.Anything, req).
					Return(expected, nil)

				out, err := sol.SubmitTransaction(ctx, req)
				require.NoError(t, err)
				require.Equal(t, expected.IdempotencyKey, out.IdempotencyKey)
				require.Equal(t, expected.Signature, out.Signature)
				require.Equal(t, expected.Status, out.Status)
			},
		},
		{
			name: "RegisterLogTracking",
			run: func(t *testing.T, sol types.SolanaService, mockSol *mocks2.SolanaService) {
				req := soltypes.LPFilterQuery{
					Name:          "my-filter",
					Address:       soltypes.PublicKey{0x11},
					EventName:     "MyEvent",
					StartingBlock: 1234,
					Retention:     3600,
				}
				mockSol.EXPECT().
					RegisterLogTracking(mock.Anything, req).
					Return(nil)

				err := sol.RegisterLogTracking(ctx, req)
				require.NoError(t, err)
			},
		},
		{
			name: "UnregisterLogTracking",
			run: func(t *testing.T, sol types.SolanaService, mockSol *mocks2.SolanaService) {
				filterName := "my-filter"
				mockSol.EXPECT().
					UnregisterLogTracking(mock.Anything, filterName).
					Return(nil)

				err := sol.UnregisterLogTracking(ctx, filterName)
				require.NoError(t, err)
			},
		},
		{
			name: "QueryTrackedLogs",
			run: func(t *testing.T, sol types.SolanaService, mockSol *mocks2.SolanaService) {
				// simple expression & limit
				filterExpr := []query.Expression{}
				primitiveExpressions := []query.Expression{query.TxHash("txHash")}
				primitiveExpressions = append(primitiveExpressions, solprimitives.NewAddressFilter(soltypes.PublicKey{1, 2, 3}))
				filterExpr = append(filterExpr, primitiveExpressions...)
				expected := []*soltypes.Log{
					{BlockNumber: 1, LogIndex: 0},
					{BlockNumber: 2, LogIndex: 5},
				}

				expLimitAndSort := query.NewLimitAndSort(query.CountLimit(10), query.SortByTimestamp{})
				mockSol.EXPECT().
					QueryTrackedLogs(mock.Anything, filterExpr, expLimitAndSort).
					Return(expected, nil)

				out, err := sol.QueryTrackedLogs(ctx, filterExpr, expLimitAndSort)
				require.NoError(t, err)
				require.Len(t, out, 2)
				require.Equal(t, int64(2), out[1].BlockNumber)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockSol := mocks2.NewSolanaService(t)
			relayer1.On("Solana", mock.Anything, mock.Anything).Return(mockSol, nil).Once()

			fetchedSol, err := retrievedRelayer.Solana()
			require.NoError(t, err)

			tc.run(t, fetchedSol, mockSol)
		})
	}
}

func Test_RelayerSet_AptosService(t *testing.T) {
	ctx := t.Context()
	stopCh := make(chan struct{})
	log := logger.Test(t)

	relayer1 := mocks.NewRelayer(t)
	relayers := map[types.RelayID]core.Relayer{
		{Network: "N1", ChainID: "C1"}: relayer1,
	}

	pluginName := "aptos-relayerset-test"
	client, server := plugin.TestPluginGRPCConn(
		t,
		true,
		map[string]plugin.Plugin{
			pluginName: &testRelaySetPlugin{
				log:  log,
				impl: &TestRelayerSet{relayers: relayers},
				brokerExt: &net.BrokerExt{
					BrokerConfig: net.BrokerConfig{
						StopCh: stopCh,
						Logger: log,
					},
				},
			},
		},
	)
	defer client.Close()
	defer server.Stop()

	relayerSetClient, err := client.Dispense(pluginName)
	require.NoError(t, err)
	rc, ok := relayerSetClient.(*Client)
	require.True(t, ok)

	retrievedRelayer, err := rc.Get(ctx, types.RelayID{Network: "N1", ChainID: "C1"})
	require.NoError(t, err)

	tests := []struct {
		name string
		run  func(t *testing.T, apt types.AptosService, mockApt *mocks2.AptosService)
	}{
		{
			name: "AccountAPTBalance",
			run: func(t *testing.T, apt types.AptosService, mockApt *mocks2.AptosService) {
				address := aptos.AccountAddress{
					0x01, 0x02, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				}
				balance := uint64(1000000000) // 1 APT
				req := aptos.AccountAPTBalanceRequest{Address: address}
				mockApt.EXPECT().AccountAPTBalance(mock.Anything, req).
					Return(&aptos.AccountAPTBalanceReply{Value: balance}, nil)

				reply, err := apt.AccountAPTBalance(ctx, req)
				require.NoError(t, err)
				require.Equal(t, balance, reply.Value)
			},
		},
		{
			name: "View",
			run: func(t *testing.T, apt types.AptosService, mockApt *mocks2.AptosService) {
				moduleAddr := aptos.AccountAddress{
					0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				}
				req := aptos.ViewRequest{
					Payload: &aptos.ViewPayload{
						Module: aptos.ModuleID{
							Address: moduleAddr,
							Name:    "coin",
						},
						Function: "balance",
						ArgTypes: []aptos.TypeTag{
							{Value: aptos.AddressTag{}},
						},
						Args: [][]byte{{0x01, 0x02}},
					},
				}
				expectedData := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x64} // 100
				mockApt.EXPECT().View(mock.Anything, req).
					Return(&aptos.ViewReply{Data: expectedData}, nil)

				reply, err := apt.View(ctx, req)
				require.NoError(t, err)
				require.Equal(t, expectedData, reply.Data)
			},
		},
		{
			name: "TransactionByHash",
			run: func(t *testing.T, apt types.AptosService, mockApt *mocks2.AptosService) {
				txHash := "0xabcdef1234567890"
				req := aptos.TransactionByHashRequest{Hash: txHash}
				version := uint64(12345)
				success := true
				expectedTx := &aptos.Transaction{
					Type:    aptos.TransactionVariantUser,
					Hash:    txHash,
					Version: &version,
					Success: &success,
					Data:    []byte(`{"sender":"0x1","sequence_number":42}`),
				}
				mockApt.EXPECT().TransactionByHash(mock.Anything, req).
					Return(&aptos.TransactionByHashReply{Transaction: expectedTx}, nil)

				reply, err := apt.TransactionByHash(ctx, req)
				require.NoError(t, err)
				require.NotNil(t, reply.Transaction)
				require.Equal(t, txHash, reply.Transaction.Hash)
				require.Equal(t, aptos.TransactionVariantUser, reply.Transaction.Type)
				require.NotNil(t, reply.Transaction.Version)
				require.Equal(t, version, *reply.Transaction.Version)
				require.NotNil(t, reply.Transaction.Success)
				require.True(t, *reply.Transaction.Success)
			},
		},
		{
			name: "TransactionByHash - Pending",
			run: func(t *testing.T, apt types.AptosService, mockApt *mocks2.AptosService) {
				txHash := "0xpending123"
				req := aptos.TransactionByHashRequest{Hash: txHash}
				expectedTx := &aptos.Transaction{
					Type:    aptos.TransactionVariantPending,
					Hash:    txHash,
					Version: nil,
					Success: nil,
					Data:    []byte(`{"pending":true}`),
				}
				mockApt.EXPECT().TransactionByHash(mock.Anything, req).
					Return(&aptos.TransactionByHashReply{Transaction: expectedTx}, nil)

				reply, err := apt.TransactionByHash(ctx, req)
				require.NoError(t, err)
				require.NotNil(t, reply.Transaction)
				require.Equal(t, aptos.TransactionVariantPending, reply.Transaction.Type)
				require.Nil(t, reply.Transaction.Version)
				require.Nil(t, reply.Transaction.Success)
			},
		},
		{
			name: "SubmitTransaction",
			run: func(t *testing.T, apt types.AptosService, mockApt *mocks2.AptosService) {
				receiverAddr := aptos.AccountAddress{
					0xBB, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
					0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				}
				req := aptos.SubmitTransactionRequest{
					ReceiverModuleID: aptos.ModuleID{
						Address: receiverAddr,
						Name:    "receiver",
					},
					EncodedPayload: []byte{0x01, 0x02, 0x03, 0x04},
					GasConfig: &aptos.GasConfig{
						MaxGasAmount: 10000,
						GasUnitPrice: 100,
					},
				}
				expectedReply := &aptos.SubmitTransactionReply{
					TxStatus:         aptos.TxSuccess,
					TxHash:           "0xtxhash123",
					TxIdempotencyKey: "key-456",
				}
				mockApt.EXPECT().SubmitTransaction(mock.Anything, req).
					Return(expectedReply, nil)

				reply, err := apt.SubmitTransaction(ctx, req)
				require.NoError(t, err)
				require.NotNil(t, reply.TxHash)
				require.Equal(t, expectedReply.TxHash, reply.TxHash)
				require.Equal(t, expectedReply.TxIdempotencyKey, reply.TxIdempotencyKey)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockApt := mocks2.NewAptosService(t)
			relayer1.On("Aptos", mock.Anything, mock.Anything).Return(mockApt, nil).Once()

			fetchedApt, err := retrievedRelayer.Aptos()
			require.NoError(t, err)

			tc.run(t, fetchedApt, mockApt)
		})
	}
}

func Test_RelayerSet_TONService(t *testing.T) {
	ctx := t.Context()
	stopCh := make(chan struct{})
	log := logger.Test(t)

	relayer1 := mocks.NewRelayer(t)
	relayers := map[types.RelayID]core.Relayer{
		{Network: "N1", ChainID: "C1"}: relayer1,
	}

	pluginName := "ton-relayerset-test"
	client, server := plugin.TestPluginGRPCConn(
		t,
		true,
		map[string]plugin.Plugin{
			pluginName: &testRelaySetPlugin{
				log:  log,
				impl: &TestRelayerSet{relayers: relayers},
				brokerExt: &net.BrokerExt{
					BrokerConfig: net.BrokerConfig{
						StopCh: stopCh,
						Logger: log,
					},
				},
			},
		},
	)
	defer client.Close()
	defer server.Stop()

	relayerSetClient, err := client.Dispense(pluginName)
	require.NoError(t, err)
	rc, ok := relayerSetClient.(*Client)
	require.True(t, ok)

	retrievedRelayer, err := rc.Get(ctx, types.RelayID{Network: "N1", ChainID: "C1"})
	require.NoError(t, err)

	tests := []struct {
		name string
		run  func(t *testing.T, ton types.TONService, mockTON *mocks2.TONService)
	}{
		{
			name: "GetMasterchainInfo",
			run: func(t *testing.T, ton types.TONService, mockTON *mocks2.TONService) {
				blockIDExt := &tontypes.BlockIDExt{
					Workchain: 0,
					Shard:     123,
					SeqNo:     1,
				}
				mockTON.EXPECT().GetMasterchainInfo(mock.Anything).Return(blockIDExt, nil)
				out, err := ton.GetMasterchainInfo(ctx)
				require.NoError(t, err)
				require.Equal(t, blockIDExt, out)
			},
		},
		{
			name: "GetBlockData",
			run: func(t *testing.T, ton types.TONService, mockTON *mocks2.TONService) {
				blockIDExt := &tontypes.BlockIDExt{Workchain: 0, Shard: 1, SeqNo: 100}
				block := &tontypes.Block{GlobalID: -217}
				mockTON.EXPECT().GetBlockData(mock.Anything, blockIDExt).Return(block, nil)
				out, err := ton.GetBlockData(ctx, blockIDExt)
				require.NoError(t, err)
				require.Equal(t, block, out)
			},
		},
		{
			name: "GetAccountBalance",
			run: func(t *testing.T, ton types.TONService, mockTON *mocks2.TONService) {
				addr := "0:abc123"
				blockID := &tontypes.BlockIDExt{Workchain: 0, Shard: 1, SeqNo: 100}
				balance := &tontypes.Balance{}
				mockTON.EXPECT().GetAccountBalance(mock.Anything, addr, blockID).Return(balance, nil)
				out, err := ton.GetAccountBalance(ctx, addr, blockID)
				require.NoError(t, err)
				require.Equal(t, balance, out)
			},
		},
		{
			name: "SendTx",
			run: func(t *testing.T, ton types.TONService, mockTON *mocks2.TONService) {
				addr := "0:abc123"
				body := []byte("body")
				stateInit := []byte("state-init")
				msg := tontypes.Message{
					Mode:      1,
					ToAddress: addr,
					Amount:    "1.0",
					Bounce:    false,
					Body:      body,
					StateInit: stateInit,
				}
				mockTON.EXPECT().SendTx(mock.Anything, msg).Return(nil)
				err := ton.SendTx(ctx, msg)
				require.NoError(t, err)
			},
		},
		{
			name: "GetTxStatus",
			run: func(t *testing.T, ton types.TONService, mockTON *mocks2.TONService) {
				lt := uint64(123456)
				status := types.Finalized
				exitCode := tontypes.ExitCode(0)
				mockTON.EXPECT().GetTxStatus(mock.Anything, lt).Return(status, exitCode, nil)
				s, c, err := ton.GetTxStatus(ctx, lt)
				require.NoError(t, err)
				require.Equal(t, status, s)
				require.Equal(t, exitCode, c)
			},
		},
		{
			name: "GetTxExecutionFees",
			run: func(t *testing.T, ton types.TONService, mockTON *mocks2.TONService) {
				lt := uint64(123456)
				fees := &tontypes.TransactionFee{TransactionFee: big.NewInt(100)}
				mockTON.EXPECT().GetTxExecutionFees(mock.Anything, lt).Return(fees, nil)
				out, err := ton.GetTxExecutionFees(ctx, lt)
				require.NoError(t, err)
				require.Equal(t, fees, out)
			},
		},
		{
			name: "HasFilter",
			run: func(t *testing.T, ton types.TONService, mockTON *mocks2.TONService) {
				filterName := "myFilter"
				mockTON.EXPECT().HasFilter(mock.Anything, filterName).Return(true)
				ok := ton.HasFilter(ctx, filterName)
				require.True(t, ok)
			},
		},
		{
			name: "RegisterFilter",
			run: func(t *testing.T, ton types.TONService, mockTON *mocks2.TONService) {
				filter := tontypes.LPFilterQuery{Name: "filter1"}
				mockTON.EXPECT().RegisterFilter(mock.Anything, filter).Return(nil)
				err := ton.RegisterFilter(ctx, filter)
				require.NoError(t, err)
			},
		},
		{
			name: "UnregisterFilter",
			run: func(t *testing.T, ton types.TONService, mockTON *mocks2.TONService) {
				filterName := "filter1"
				mockTON.EXPECT().UnregisterFilter(mock.Anything, filterName).Return(nil)
				err := ton.UnregisterFilter(ctx, filterName)
				require.NoError(t, err)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockTON := mocks2.NewTONService(t)
			relayer1.On("TON", mock.Anything, mock.Anything).Return(mockTON, nil).Once()

			fetchedTON, err := retrievedRelayer.TON()
			require.NoError(t, err)

			tc.run(t, fetchedTON, mockTON)
		})
	}
}

type TestContractReader struct {
	types.UnimplementedContractReader
	mockedContractReader *mocks2.ContractReader
}

func (t *TestContractReader) Start(ctx context.Context) error {
	return t.mockedContractReader.Start(ctx)
}

func (t *TestContractReader) Close() error {
	return t.mockedContractReader.Close()
}

func (t *TestContractReader) Ready() error {
	return t.mockedContractReader.Ready()
}
func (t *TestContractReader) HealthReport() map[string]error {
	return t.mockedContractReader.HealthReport()
}

func (t *TestContractReader) Name() string {
	return t.mockedContractReader.Name()
}

func (t *TestContractReader) GetLatestValue(ctx context.Context, readIdentifier string, confidenceLevel primitives.ConfidenceLevel, params, returnVal any) error {
	return t.mockedContractReader.GetLatestValue(ctx, readIdentifier, confidenceLevel, params, returnVal)
}

func (t *TestContractReader) GetLatestValueWithHeadData(ctx context.Context, readIdentifier string, confidenceLevel primitives.ConfidenceLevel, params, returnVal any) (*types.Head, error) {
	return t.mockedContractReader.GetLatestValueWithHeadData(ctx, readIdentifier, confidenceLevel, params, returnVal)
}

func (t *TestContractReader) BatchGetLatestValues(ctx context.Context, request types.BatchGetLatestValuesRequest) (types.BatchGetLatestValuesResult, error) {
	return t.mockedContractReader.BatchGetLatestValues(ctx, request)
}

func (t *TestContractReader) Bind(ctx context.Context, bindings []types.BoundContract) error {
	return t.mockedContractReader.Bind(ctx, bindings)
}

func (t *TestContractReader) Unbind(ctx context.Context, bindings []types.BoundContract) error {
	return t.mockedContractReader.Unbind(ctx, bindings)
}

func (t *TestContractReader) QueryKey(ctx context.Context, boundContract types.BoundContract, filter query.KeyFilter, limitAndSort query.LimitAndSort, sequenceDataType any) ([]types.Sequence, error) {
	return t.mockedContractReader.QueryKey(ctx, boundContract, filter, limitAndSort, sequenceDataType)
}

func (t *TestContractReader) QueryKeys(ctx context.Context, keyQueries []types.ContractKeyFilter, limitAndSort query.LimitAndSort) (iter.Seq2[string, types.Sequence], error) {
	return t.mockedContractReader.QueryKeys(ctx, keyQueries, limitAndSort)
}

type TestTON struct {
	mockedTONService *mocks2.TONService
}

func (t *TestTON) GetMasterchainInfo(ctx context.Context) (*tontypes.BlockIDExt, error) {
	return t.mockedTONService.GetMasterchainInfo(ctx)
}

func (t *TestTON) GetBlockData(ctx context.Context, block *tontypes.BlockIDExt) (*tontypes.Block, error) {
	return t.mockedTONService.GetBlockData(ctx, block)
}

func (t *TestTON) GetAccountBalance(ctx context.Context, address string, block *tontypes.BlockIDExt) (*tontypes.Balance, error) {
	return t.mockedTONService.GetAccountBalance(ctx, address, block)
}

func (t *TestTON) SendTx(ctx context.Context, msg ton.Message) error {
	return t.mockedTONService.SendTx(ctx, msg)
}

func (t *TestTON) GetTxStatus(ctx context.Context, lt uint64) (types.TransactionStatus, ton.ExitCode, error) {
	return t.mockedTONService.GetTxStatus(ctx, lt)
}

func (t *TestTON) GetTxExecutionFees(ctx context.Context, lt uint64) (*ton.TransactionFee, error) {
	return t.mockedTONService.GetTxExecutionFees(ctx, lt)
}

func (t *TestTON) HasFilter(ctx context.Context, name string) bool {
	return t.mockedTONService.HasFilter(ctx, name)
}

func (t *TestTON) RegisterFilter(ctx context.Context, filter ton.LPFilterQuery) error {
	return t.mockedTONService.RegisterFilter(ctx, filter)
}

func (t *TestTON) UnregisterFilter(ctx context.Context, name string) error {
	return t.mockedTONService.UnregisterFilter(ctx, name)
}

var _ types.TONService = (*TestTON)(nil)

type TestRelayerSet struct {
	relayers map[types.RelayID]core.Relayer
}

func (t *TestRelayerSet) Get(ctx context.Context, relayID types.RelayID) (core.Relayer, error) {
	if relayer, ok := t.relayers[relayID]; ok {
		return relayer, nil
	}

	return nil, fmt.Errorf("relayer with id %s not found", relayID)
}

func (t *TestRelayerSet) List(ctx context.Context, relayIDs ...types.RelayID) (map[types.RelayID]core.Relayer, error) {
	return t.relayers, nil
}

type testRelaySetPlugin struct {
	log logger.Logger
	plugin.NetRPCUnsupportedPlugin
	brokerExt *net.BrokerExt
	impl      core.RelayerSet
}

func (r *testRelaySetPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, client *grpc.ClientConn) (any, error) {
	r.brokerExt.Broker = broker

	return NewRelayerSetClient(logger.Nop(), r.brokerExt, client), nil
}

func (r *testRelaySetPlugin) GRPCServer(broker *plugin.GRPCBroker, server *grpc.Server) error {
	r.brokerExt.Broker = broker

	rs, _ := NewRelayerSetServer(r.log, r.impl, r.brokerExt)
	relayerset.RegisterRelayerSetServerWithDependants(server, rs)
	return nil
}

func generateFixtureQuery() []query.Expression {
	exprs := make([]query.Expression, 0)

	confirmationsValues := []primitives.ConfidenceLevel{primitives.Finalized, primitives.Unconfirmed, primitives.Safe}
	operatorValues := []primitives.ComparisonOperator{primitives.Eq, primitives.Neq, primitives.Gt, primitives.Lt, primitives.Gte, primitives.Lte}

	primitiveExpressions := []query.Expression{query.TxHash("txHash")}
	values := []evmtypes.Hash{topic3, topic2}
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
