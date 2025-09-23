package test

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	testtypes "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/ccipocr3"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
)

type ChainAccessorEvaluator interface {
	ccipocr3.ChainAccessor
	testtypes.Evaluator[ccipocr3.ChainAccessor]
}

type ChainAccessorTester interface {
	ccipocr3.ChainAccessor
	testtypes.Evaluator[ccipocr3.ChainAccessor]
	testtypes.AssertEqualer[ccipocr3.ChainAccessor]
}

// ChainAccessor is a static implementation of the ChainAccessorTester interface.
// It is to be used in tests to verify grpc implementations of the ChainAccessor interface.
func ChainAccessor(lggr logger.Logger) staticChainAccessor {
	testTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	return newStaticChainAccessor(lggr, staticChainAccessorConfig{
		contractAddress: []byte("test-contract-address"),
		chainConfigSnapshot: ccipocr3.ChainConfigSnapshot{
			Offramp: ccipocr3.OfframpConfig{
				CommitLatestOCRConfig: ccipocr3.OCRConfigResponse{
					OCRConfig: ccipocr3.OCRConfig{
						ConfigInfo: ccipocr3.ConfigInfo{
							ConfigDigest:                   ccipocr3.Bytes32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
							F:                              3,
							N:                              10,
							IsSignatureVerificationEnabled: true,
						},
						Signers:      [][]byte{[]byte("signer1"), []byte("signer2"), []byte("signer3")},
						Transmitters: [][]byte{[]byte("transmitter1"), []byte("transmitter2"), []byte("transmitter3")},
					},
				},
				ExecLatestOCRConfig: ccipocr3.OCRConfigResponse{
					OCRConfig: ccipocr3.OCRConfig{
						ConfigInfo: ccipocr3.ConfigInfo{
							ConfigDigest:                   ccipocr3.Bytes32{32, 31, 30, 29, 28, 27, 26, 25, 24, 23, 22, 21, 20, 19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1},
							F:                              2,
							N:                              7,
							IsSignatureVerificationEnabled: true,
						},
						Signers:      [][]byte{[]byte("exec-signer1"), []byte("exec-signer2")},
						Transmitters: [][]byte{[]byte("exec-transmitter1"), []byte("exec-transmitter2")},
					},
				},
				StaticConfig: ccipocr3.OffRampStaticChainConfig{
					ChainSelector:        ccipocr3.ChainSelector(1),
					GasForCallExactCheck: 5000,
					RmnRemote:            []byte("rmn-remote-address"),
					TokenAdminRegistry:   []byte("token-admin-registry"),
					NonceManager:         []byte("nonce-manager-address"),
				},
				DynamicConfig: ccipocr3.OffRampDynamicChainConfig{
					FeeQuoter:                               []byte("fee-quoter-address"),
					PermissionLessExecutionThresholdSeconds: 3600,
					IsRMNVerificationDisabled:               false,
					MessageInterceptor:                      []byte("message-interceptor"),
				},
			},
			RMNProxy: ccipocr3.RMNProxyConfig{
				RemoteAddress: []byte("rmn-proxy-remote"),
			},
			RMNRemote: ccipocr3.RMNRemoteConfig{
				DigestHeader: ccipocr3.RMNDigestHeader{
					DigestHeader: ccipocr3.Bytes32{11, 22, 33, 44, 55, 66, 77, 88, 99, 110, 121, 132, 143, 154, 165, 176, 187, 198, 209, 220, 231, 242, 253, 8, 19, 30, 41, 52, 63, 74, 85, 96},
				},
				VersionedConfig: ccipocr3.VersionedConfig{
					Version: 1,
					Config: ccipocr3.Config{
						RMNHomeContractConfigDigest: ccipocr3.Bytes32{255, 254, 253, 252, 251, 250, 249, 248, 247, 246, 245, 244, 243, 242, 241, 240, 239, 238, 237, 236, 235, 234, 233, 232, 231, 230, 229, 228, 227, 226, 225, 224},
						Signers: []ccipocr3.Signer{
							{
								OnchainPublicKey: []byte("rmn-signer-1-pubkey"),
								NodeIndex:        0,
							},
							{
								OnchainPublicKey: []byte("rmn-signer-2-pubkey"),
								NodeIndex:        1,
							},
						},
						FSign: 1,
					},
				},
			},
			FeeQuoter: ccipocr3.FeeQuoterConfig{
				StaticConfig: ccipocr3.FeeQuoterStaticConfig{
					MaxFeeJuelsPerMsg:  ccipocr3.NewBigInt(big.NewInt(1000000)),
					LinkToken:          []byte("link-token-address"),
					StalenessThreshold: 3600,
				},
			},
			OnRamp: ccipocr3.OnRampConfig{
				DynamicConfig: ccipocr3.GetOnRampDynamicConfigResponse{
					DynamicConfig: ccipocr3.OnRampDynamicConfig{
						FeeQuoter:              []byte("onramp-fee-quoter"),
						ReentrancyGuardEntered: false,
						MessageInterceptor:     []byte("onramp-interceptor"),
						FeeAggregator:          []byte("fee-aggregator"),
						AllowListAdmin:         []byte("allowlist-admin"),
					},
				},
				DestChainConfig: ccipocr3.OnRampDestChainConfig{
					SequenceNumber:   42,
					AllowListEnabled: true,
					Router:           []byte("onramp-router"),
				},
			},
			Router: ccipocr3.RouterConfig{
				WrappedNativeAddress: ccipocr3.Bytes([]byte("wrapped-native-token")),
			},
			CurseInfo: ccipocr3.CurseInfo{
				CursedSourceChains: map[ccipocr3.ChainSelector]bool{
					ccipocr3.ChainSelector(2): true,
					ccipocr3.ChainSelector(3): false,
				},
				CursedDestination: false,
				GlobalCurse:       false,
			},
		},
		sourceChainConfigs: map[ccipocr3.ChainSelector]ccipocr3.SourceChainConfig{
			ccipocr3.ChainSelector(1): {
				Router:                    []byte("router1"),
				IsEnabled:                 true,
				IsRMNVerificationDisabled: false,
				MinSeqNr:                  uint64(3),
				OnRamp:                    ccipocr3.UnknownAddress("onRamp1"),
			},
			ccipocr3.ChainSelector(2): {
				Router:                    []byte("router2"),
				IsEnabled:                 true,
				IsRMNVerificationDisabled: true,
				MinSeqNr:                  uint64(10),
				OnRamp:                    ccipocr3.UnknownAddress("onRamp2"),
			},
		},
		chainFeeComponents: ccipocr3.ChainFeeComponents{
			ExecutionFee:        big.NewInt(50000),
			DataAvailabilityFee: big.NewInt(25000),
		},
		commitReports: []ccipocr3.CommitPluginReportWithMeta{
			{
				Report: ccipocr3.CommitPluginReport{
					PriceUpdates: ccipocr3.PriceUpdates{
						TokenPriceUpdates: []ccipocr3.TokenPrice{
							{
								TokenID: ccipocr3.UnknownEncodedAddress("token1"),
								Price:   ccipocr3.NewBigInt(big.NewInt(1500000)),
							},
						},
						GasPriceUpdates: []ccipocr3.GasPriceChain{
							{
								ChainSel: ccipocr3.ChainSelector(1),
								GasPrice: ccipocr3.NewBigInt(big.NewInt(20000000000)),
							},
						},
					},
					BlessedMerkleRoots: []ccipocr3.MerkleRootChain{
						{
							ChainSel:      ccipocr3.ChainSelector(1),
							OnRampAddress: ccipocr3.UnknownAddress("onramp1"),
							SeqNumsRange: ccipocr3.SeqNumRange{
								ccipocr3.SeqNum(1),
								ccipocr3.SeqNum(10),
							},
							MerkleRoot: ccipocr3.Bytes32{100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115, 116, 117, 118, 119, 120, 121, 122, 123, 124, 125, 126, 127, 128, 129, 130, 131},
						},
					},
					UnblessedMerkleRoots: []ccipocr3.MerkleRootChain{},
					RMNSignatures:        []ccipocr3.RMNECDSASignature{},
				},
				Timestamp: testTime,
				BlockNum:  123456,
			},
		},
		executedMessages: map[ccipocr3.ChainSelector][]ccipocr3.SeqNum{
			ccipocr3.ChainSelector(1): {ccipocr3.SeqNum(1), ccipocr3.SeqNum(2), ccipocr3.SeqNum(3)},
			ccipocr3.ChainSelector(2): {ccipocr3.SeqNum(5), ccipocr3.SeqNum(6)},
		},
		nextSeqNums: map[ccipocr3.ChainSelector]ccipocr3.SeqNum{
			ccipocr3.ChainSelector(1): ccipocr3.SeqNum(4),
			ccipocr3.ChainSelector(2): ccipocr3.SeqNum(7),
		},
		nonces: map[ccipocr3.ChainSelector]map[string]uint64{
			ccipocr3.ChainSelector(1): {
				"sender1": 10,
				"sender2": 25,
			},
			ccipocr3.ChainSelector(2): {
				"sender3": 5,
			},
		},
		chainFeePriceUpdates: map[ccipocr3.ChainSelector]ccipocr3.TimestampedUnixBig{
			ccipocr3.ChainSelector(1): {
				Timestamp: uint32(testTime.Unix()),
				Value:     big.NewInt(2000000000),
			},
			ccipocr3.ChainSelector(2): {
				Timestamp: uint32(testTime.Add(time.Hour).Unix()),
				Value:     big.NewInt(1800000000),
			},
		},
		latestPriceSeqNr: 42,
		messages: []ccipocr3.Message{
			{
				Header: ccipocr3.RampMessageHeader{
					MessageID:           ccipocr3.Bytes32{200, 201, 202, 203, 204, 205, 206, 207, 208, 209, 210, 211, 212, 213, 214, 215, 216, 217, 218, 219, 220, 221, 222, 223, 224, 225, 226, 227, 228, 229, 230, 231},
					SourceChainSelector: ccipocr3.ChainSelector(1),
					DestChainSelector:   ccipocr3.ChainSelector(2),
					SequenceNumber:      ccipocr3.SeqNum(1),
					Nonce:               1,
					MsgHash:             ccipocr3.Bytes32{},
					OnRamp:              ccipocr3.UnknownAddress("test-onramp"),
					TxHash:              "0x123abc456def789",
				},
				Sender:         ccipocr3.UnknownAddress("sender1"),
				Data:           ccipocr3.Bytes("test message data"),
				Receiver:       ccipocr3.UnknownAddress("receiver1"),
				ExtraArgs:      ccipocr3.Bytes("extra args"),
				FeeToken:       ccipocr3.UnknownAddress("fee-token"),
				FeeTokenAmount: ccipocr3.NewBigInt(big.NewInt(1000)),
				FeeValueJuels:  ccipocr3.NewBigInt(big.NewInt(5000)),
				TokenAmounts: []ccipocr3.RampTokenAmount{
					{
						SourcePoolAddress: ccipocr3.UnknownAddress("source-pool"),
						DestTokenAddress:  ccipocr3.UnknownAddress("dest-token"),
						ExtraData:         ccipocr3.Bytes("token-extra-data"),
						Amount:            ccipocr3.NewBigInt(big.NewInt(10000)),
						DestExecData:      ccipocr3.Bytes("dest-exec-data"),
					},
				},
			},
		},
		latestMessageSeqNr: 43,
		expectedNextSeqNr:  44,
		tokenPriceUSD: ccipocr3.TimestampedUnixBig{
			Value:     big.NewInt(1500000000),
			Timestamp: uint32(testTime.Unix()),
		},
		feeQuoterDestChainConfig: ccipocr3.FeeQuoterDestChainConfig{
			IsEnabled:                         true,
			MaxNumberOfTokensPerMsg:           5,
			MaxDataBytes:                      10000,
			MaxPerMsgGasLimit:                 3000000,
			DestGasOverhead:                   50000,
			DestGasPerPayloadByteBase:         16,
			DestGasPerPayloadByteHigh:         24,
			DestGasPerPayloadByteThreshold:    1000,
			DestDataAvailabilityOverheadGas:   10000,
			DestGasPerDataAvailabilityByte:    4,
			DestDataAvailabilityMultiplierBps: 6840,
			DefaultTokenFeeUSDCents:           50,
			DefaultTokenDestGasOverhead:       125000,
			DefaultTxGasLimit:                 200000,
			GasMultiplierWeiPerEth:            1100000000000000000,
			NetworkFeeUSDCents:                500,
			GasPriceStalenessThreshold:        3600,
			EnforceOutOfOrder:                 false,
			ChainFamilySelector:               [4]byte{0x45, 0x56, 0x4d, 0x00},
		},
		rmnCurseInfo: ccipocr3.CurseInfo{
			CursedSourceChains: map[ccipocr3.ChainSelector]bool{
				ccipocr3.ChainSelector(99): true,
			},
			CursedDestination: false,
			GlobalCurse:       false,
		},
	})
}

var _ ChainAccessorTester = staticChainAccessor{}

type staticChainAccessorConfig struct {
	contractAddress          []byte
	chainConfigSnapshot      ccipocr3.ChainConfigSnapshot
	sourceChainConfigs       map[ccipocr3.ChainSelector]ccipocr3.SourceChainConfig
	chainFeeComponents       ccipocr3.ChainFeeComponents
	commitReports            []ccipocr3.CommitPluginReportWithMeta
	executedMessages         map[ccipocr3.ChainSelector][]ccipocr3.SeqNum
	nextSeqNums              map[ccipocr3.ChainSelector]ccipocr3.SeqNum
	nonces                   map[ccipocr3.ChainSelector]map[string]uint64
	chainFeePriceUpdates     map[ccipocr3.ChainSelector]ccipocr3.TimestampedUnixBig
	latestPriceSeqNr         uint64
	messages                 []ccipocr3.Message
	latestMessageSeqNr       ccipocr3.SeqNum
	expectedNextSeqNr        ccipocr3.SeqNum
	tokenPriceUSD            ccipocr3.TimestampedUnixBig
	feeQuoterDestChainConfig ccipocr3.FeeQuoterDestChainConfig
	rmnCurseInfo             ccipocr3.CurseInfo
}

type staticChainAccessor struct {
	staticChainAccessorConfig
}

func newStaticChainAccessor(lggr logger.Logger, cfg staticChainAccessorConfig) staticChainAccessor {
	lggr = logger.Named(lggr, "staticChainAccessor")
	return staticChainAccessor{
		staticChainAccessorConfig: cfg,
	}
}

// AllAccessors implementation

func (s staticChainAccessor) GetContractAddress(contractName string) ([]byte, error) {
	return s.contractAddress, nil
}

func (s staticChainAccessor) GetAllConfigsLegacy(ctx context.Context, destChainSelector ccipocr3.ChainSelector, sourceChainSelectors []ccipocr3.ChainSelector) (ccipocr3.ChainConfigSnapshot, map[ccipocr3.ChainSelector]ccipocr3.SourceChainConfig, error) {
	return s.chainConfigSnapshot, s.sourceChainConfigs, nil
}

func (s staticChainAccessor) GetChainFeeComponents(ctx context.Context) (ccipocr3.ChainFeeComponents, error) {
	return s.chainFeeComponents, nil
}

func (s staticChainAccessor) Sync(ctx context.Context, contractName string, contractAddress ccipocr3.UnknownAddress) error {
	return nil
}

// DestinationAccessor implementation

func (s staticChainAccessor) CommitReportsGTETimestamp(ctx context.Context, ts time.Time, confidence primitives.ConfidenceLevel, limit int) ([]ccipocr3.CommitPluginReportWithMeta, error) {
	return s.commitReports, nil
}

func (s staticChainAccessor) ExecutedMessages(ctx context.Context, ranges map[ccipocr3.ChainSelector][]ccipocr3.SeqNumRange, confidence primitives.ConfidenceLevel) (map[ccipocr3.ChainSelector][]ccipocr3.SeqNum, error) {
	return s.executedMessages, nil
}

func (s staticChainAccessor) NextSeqNum(ctx context.Context, sources []ccipocr3.ChainSelector) (map[ccipocr3.ChainSelector]ccipocr3.SeqNum, error) {
	return s.nextSeqNums, nil
}

func (s staticChainAccessor) Nonces(ctx context.Context, addresses map[ccipocr3.ChainSelector][]ccipocr3.UnknownEncodedAddress) (map[ccipocr3.ChainSelector]map[string]uint64, error) {
	return s.nonces, nil
}

func (s staticChainAccessor) GetChainFeePriceUpdate(ctx context.Context, chains []ccipocr3.ChainSelector) (map[ccipocr3.ChainSelector]ccipocr3.TimestampedUnixBig, error) {
	return s.chainFeePriceUpdates, nil
}

func (s staticChainAccessor) GetLatestPriceSeqNr(ctx context.Context) (ccipocr3.SeqNum, error) {
	return ccipocr3.SeqNum(s.latestPriceSeqNr), nil
}

// SourceAccessor implementation

func (s staticChainAccessor) MsgsBetweenSeqNums(ctx context.Context, dest ccipocr3.ChainSelector, seqNumRange ccipocr3.SeqNumRange) ([]ccipocr3.Message, error) {
	return s.messages, nil
}

func (s staticChainAccessor) LatestMessageTo(ctx context.Context, dest ccipocr3.ChainSelector) (ccipocr3.SeqNum, error) {
	return s.latestMessageSeqNr, nil
}

func (s staticChainAccessor) GetExpectedNextSequenceNumber(ctx context.Context, dest ccipocr3.ChainSelector) (ccipocr3.SeqNum, error) {
	return s.expectedNextSeqNr, nil
}

func (s staticChainAccessor) GetTokenPriceUSD(ctx context.Context, address ccipocr3.UnknownAddress) (ccipocr3.TimestampedUnixBig, error) {
	return s.tokenPriceUSD, nil
}

func (s staticChainAccessor) GetFeeQuoterDestChainConfig(ctx context.Context, dest ccipocr3.ChainSelector) (ccipocr3.FeeQuoterDestChainConfig, error) {
	return s.feeQuoterDestChainConfig, nil
}

// USDCMessageReader implementation
func (s staticChainAccessor) MessagesByTokenID(ctx context.Context, source, dest ccipocr3.ChainSelector, tokens map[ccipocr3.MessageTokenID]ccipocr3.RampTokenAmount) (map[ccipocr3.MessageTokenID]ccipocr3.Bytes, error) {
	// Return static test data for USDC messages
	result := make(map[ccipocr3.MessageTokenID]ccipocr3.Bytes)
	for tokenID := range tokens {
		// Return some test message bytes
		result[tokenID] = ccipocr3.Bytes(fmt.Sprintf("usdc-message-data-for-%s", tokenID.String()))
	}
	return result, nil
}

// PriceReader implementation
func (s staticChainAccessor) GetFeedPricesUSD(ctx context.Context, tokens []ccipocr3.UnknownEncodedAddress, tokenInfo map[ccipocr3.UnknownEncodedAddress]ccipocr3.TokenInfo) (ccipocr3.TokenPriceMap, error) {
	// Return static test prices
	result := make(ccipocr3.TokenPriceMap)
	for i, token := range tokens {
		// Generate different prices for different tokens
		price := big.NewInt(1000000 + int64(i)*100000) // $1.00, $1.10, $1.20, etc. in wei units
		result[token] = ccipocr3.NewBigInt(price)
	}
	return result, nil
}

func (s staticChainAccessor) GetFeeQuoterTokenUpdates(ctx context.Context, tokens []ccipocr3.UnknownEncodedAddress, chain ccipocr3.ChainSelector) (map[ccipocr3.UnknownEncodedAddress]ccipocr3.TimestampedUnixBig, error) {
	// Return static test token updates
	result := make(map[ccipocr3.UnknownEncodedAddress]ccipocr3.TimestampedUnixBig)
	testTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	for i, token := range tokens {
		// Generate different prices for different tokens
		price := big.NewInt(2000000 + int64(i)*50000) // Different prices from GetFeedPricesUSD
		result[token] = ccipocr3.TimestampedUnixBig{
			Timestamp: uint32(testTime.Add(time.Duration(i) * time.Minute).Unix()), // Different timestamps
			Value:     price,
		}
	}
	return result, nil
}

// Evaluate implements ChainAccessorEvaluator.
func (s staticChainAccessor) Evaluate(ctx context.Context, other ccipocr3.ChainAccessor) error {
	// Delegate to individual evaluation functions for better readability
	evaluators := []func(context.Context, ccipocr3.ChainAccessor) error{
		s.evaluateGetContractAddress,
		s.evaluateGetAllConfigsLegacy,
		s.evaluateGetChainFeeComponents,
		s.evaluateSync,
		s.evaluateCommitReportsGTETimestamp,
		s.evaluateExecutedMessages,
		s.evaluateNextSeqNum,
		s.evaluateNonces,
		s.evaluateGetChainFeePriceUpdate,
		s.evaluateGetLatestPriceSeqNr,
		s.evaluateMsgsBetweenSeqNums,
		s.evaluateLatestMessageTo,
		s.evaluateGetExpectedNextSequenceNumber,
		s.evaluateGetTokenPriceUSD,
		s.evaluateGetFeeQuoterDestChainConfig,
		s.evaluateMessagesByTokenID,
		s.evaluateGetFeedPricesUSD,
		s.evaluateGetFeeQuoterTokenUpdates,
	}

	for _, evaluate := range evaluators {
		if err := evaluate(ctx, other); err != nil {
			return err
		}
	}

	return nil
}

// Individual evaluation functions for better readability

func (s staticChainAccessor) evaluateGetContractAddress(ctx context.Context, other ccipocr3.ChainAccessor) error {
	otherAddr, err := other.GetContractAddress("test-contract")
	if err != nil {
		return fmt.Errorf("GetContractAddress failed: %w", err)
	}
	myAddr, err := s.GetContractAddress("test-contract")
	if err != nil {
		return fmt.Errorf("GetContractAddress failed: %w", err)
	}
	if string(otherAddr) != string(myAddr) {
		return fmt.Errorf("GetContractAddress mismatch: got %s, expected %s", string(otherAddr), string(myAddr))
	}
	return nil
}

func (s staticChainAccessor) evaluateGetAllConfigsLegacy(ctx context.Context, other ccipocr3.ChainAccessor) error {
	otherSnapshot, otherConfigs, err := other.GetAllConfigsLegacy(ctx, 1, []ccipocr3.ChainSelector{2, 3})
	if err != nil {
		return fmt.Errorf("GetAllConfigsLegacy failed: %w", err)
	}
	mySnapshot, myConfigs, err := s.GetAllConfigsLegacy(ctx, 1, []ccipocr3.ChainSelector{2, 3})
	if err != nil {
		return fmt.Errorf("GetAllConfigsLegacy failed: %w", err)
	}

	// Compare all ChainConfigSnapshot fields properly

	// 1. Compare Offramp
	if otherSnapshot.Offramp.StaticConfig.ChainSelector != mySnapshot.Offramp.StaticConfig.ChainSelector {
		return fmt.Errorf("GetAllConfigsLegacy Offramp.StaticConfig.ChainSelector mismatch: got %d, expected %d",
			otherSnapshot.Offramp.StaticConfig.ChainSelector, mySnapshot.Offramp.StaticConfig.ChainSelector)
	}
	if otherSnapshot.Offramp.StaticConfig.GasForCallExactCheck != mySnapshot.Offramp.StaticConfig.GasForCallExactCheck {
		return fmt.Errorf("GetAllConfigsLegacy Offramp.StaticConfig.GasForCallExactCheck mismatch: got %d, expected %d",
			otherSnapshot.Offramp.StaticConfig.GasForCallExactCheck, mySnapshot.Offramp.StaticConfig.GasForCallExactCheck)
	}
	if string(otherSnapshot.Offramp.StaticConfig.RmnRemote) != string(mySnapshot.Offramp.StaticConfig.RmnRemote) {
		return fmt.Errorf("GetAllConfigsLegacy Offramp.StaticConfig.RmnRemote mismatch: got %s, expected %s",
			string(otherSnapshot.Offramp.StaticConfig.RmnRemote), string(mySnapshot.Offramp.StaticConfig.RmnRemote))
	}
	if string(otherSnapshot.Offramp.StaticConfig.TokenAdminRegistry) != string(mySnapshot.Offramp.StaticConfig.TokenAdminRegistry) {
		return fmt.Errorf("GetAllConfigsLegacy Offramp.StaticConfig.TokenAdminRegistry mismatch: got %s, expected %s",
			string(otherSnapshot.Offramp.StaticConfig.TokenAdminRegistry), string(mySnapshot.Offramp.StaticConfig.TokenAdminRegistry))
	}
	if string(otherSnapshot.Offramp.StaticConfig.NonceManager) != string(mySnapshot.Offramp.StaticConfig.NonceManager) {
		return fmt.Errorf("GetAllConfigsLegacy Offramp.StaticConfig.NonceManager mismatch: got %s, expected %s",
			string(otherSnapshot.Offramp.StaticConfig.NonceManager), string(mySnapshot.Offramp.StaticConfig.NonceManager))
	}

	// Compare Offramp DynamicConfig
	if string(otherSnapshot.Offramp.DynamicConfig.FeeQuoter) != string(mySnapshot.Offramp.DynamicConfig.FeeQuoter) {
		return fmt.Errorf("GetAllConfigsLegacy Offramp.DynamicConfig.FeeQuoter mismatch: got %s, expected %s",
			string(otherSnapshot.Offramp.DynamicConfig.FeeQuoter), string(mySnapshot.Offramp.DynamicConfig.FeeQuoter))
	}
	if otherSnapshot.Offramp.DynamicConfig.PermissionLessExecutionThresholdSeconds != mySnapshot.Offramp.DynamicConfig.PermissionLessExecutionThresholdSeconds {
		return fmt.Errorf("GetAllConfigsLegacy Offramp.DynamicConfig.PermissionLessExecutionThresholdSeconds mismatch: got %d, expected %d",
			otherSnapshot.Offramp.DynamicConfig.PermissionLessExecutionThresholdSeconds, mySnapshot.Offramp.DynamicConfig.PermissionLessExecutionThresholdSeconds)
	}
	if otherSnapshot.Offramp.DynamicConfig.IsRMNVerificationDisabled != mySnapshot.Offramp.DynamicConfig.IsRMNVerificationDisabled {
		return fmt.Errorf("GetAllConfigsLegacy Offramp.DynamicConfig.IsRMNVerificationDisabled mismatch: got %t, expected %t",
			otherSnapshot.Offramp.DynamicConfig.IsRMNVerificationDisabled, mySnapshot.Offramp.DynamicConfig.IsRMNVerificationDisabled)
	}
	if string(otherSnapshot.Offramp.DynamicConfig.MessageInterceptor) != string(mySnapshot.Offramp.DynamicConfig.MessageInterceptor) {
		return fmt.Errorf("GetAllConfigsLegacy Offramp.DynamicConfig.MessageInterceptor mismatch: got %s, expected %s",
			string(otherSnapshot.Offramp.DynamicConfig.MessageInterceptor), string(mySnapshot.Offramp.DynamicConfig.MessageInterceptor))
	}

	// 2. Compare RMNProxy
	if string(otherSnapshot.RMNProxy.RemoteAddress) != string(mySnapshot.RMNProxy.RemoteAddress) {
		return fmt.Errorf("GetAllConfigsLegacy RMNProxy.RemoteAddress mismatch: got %s, expected %s",
			string(otherSnapshot.RMNProxy.RemoteAddress), string(mySnapshot.RMNProxy.RemoteAddress))
	}

	// 3. Compare RMNRemote
	if otherSnapshot.RMNRemote.DigestHeader.DigestHeader != mySnapshot.RMNRemote.DigestHeader.DigestHeader {
		return fmt.Errorf("GetAllConfigsLegacy RMNRemote.DigestHeader.DigestHeader mismatch: got %x, expected %x",
			otherSnapshot.RMNRemote.DigestHeader.DigestHeader, mySnapshot.RMNRemote.DigestHeader.DigestHeader)
	}

	// 4. Compare FeeQuoter
	if otherSnapshot.FeeQuoter.StaticConfig.MaxFeeJuelsPerMsg.Cmp(mySnapshot.FeeQuoter.StaticConfig.MaxFeeJuelsPerMsg.Int) != 0 {
		return fmt.Errorf("GetAllConfigsLegacy FeeQuoter.StaticConfig.MaxFeeJuelsPerMsg mismatch: got %s, expected %s",
			otherSnapshot.FeeQuoter.StaticConfig.MaxFeeJuelsPerMsg.String(), mySnapshot.FeeQuoter.StaticConfig.MaxFeeJuelsPerMsg.String())
	}
	if string(otherSnapshot.FeeQuoter.StaticConfig.LinkToken) != string(mySnapshot.FeeQuoter.StaticConfig.LinkToken) {
		return fmt.Errorf("GetAllConfigsLegacy FeeQuoter.StaticConfig.LinkToken mismatch: got %s, expected %s",
			string(otherSnapshot.FeeQuoter.StaticConfig.LinkToken), string(mySnapshot.FeeQuoter.StaticConfig.LinkToken))
	}
	if otherSnapshot.FeeQuoter.StaticConfig.StalenessThreshold != mySnapshot.FeeQuoter.StaticConfig.StalenessThreshold {
		return fmt.Errorf("GetAllConfigsLegacy FeeQuoter.StaticConfig.StalenessThreshold mismatch: got %d, expected %d",
			otherSnapshot.FeeQuoter.StaticConfig.StalenessThreshold, mySnapshot.FeeQuoter.StaticConfig.StalenessThreshold)
	}

	// 5. Compare OnRamp
	if string(otherSnapshot.OnRamp.DynamicConfig.DynamicConfig.FeeQuoter) != string(mySnapshot.OnRamp.DynamicConfig.DynamicConfig.FeeQuoter) {
		return fmt.Errorf("GetAllConfigsLegacy OnRamp.DynamicConfig.DynamicConfig.FeeQuoter mismatch: got %s, expected %s",
			string(otherSnapshot.OnRamp.DynamicConfig.DynamicConfig.FeeQuoter), string(mySnapshot.OnRamp.DynamicConfig.DynamicConfig.FeeQuoter))
	}
	if otherSnapshot.OnRamp.DynamicConfig.DynamicConfig.ReentrancyGuardEntered != mySnapshot.OnRamp.DynamicConfig.DynamicConfig.ReentrancyGuardEntered {
		return fmt.Errorf("GetAllConfigsLegacy OnRamp.DynamicConfig.DynamicConfig.ReentrancyGuardEntered mismatch: got %t, expected %t",
			otherSnapshot.OnRamp.DynamicConfig.DynamicConfig.ReentrancyGuardEntered, mySnapshot.OnRamp.DynamicConfig.DynamicConfig.ReentrancyGuardEntered)
	}
	if string(otherSnapshot.OnRamp.DynamicConfig.DynamicConfig.MessageInterceptor) != string(mySnapshot.OnRamp.DynamicConfig.DynamicConfig.MessageInterceptor) {
		return fmt.Errorf("GetAllConfigsLegacy OnRamp.DynamicConfig.DynamicConfig.MessageInterceptor mismatch: got %s, expected %s",
			string(otherSnapshot.OnRamp.DynamicConfig.DynamicConfig.MessageInterceptor), string(mySnapshot.OnRamp.DynamicConfig.DynamicConfig.MessageInterceptor))
	}
	if string(otherSnapshot.OnRamp.DynamicConfig.DynamicConfig.FeeAggregator) != string(mySnapshot.OnRamp.DynamicConfig.DynamicConfig.FeeAggregator) {
		return fmt.Errorf("GetAllConfigsLegacy OnRamp.DynamicConfig.DynamicConfig.FeeAggregator mismatch: got %s, expected %s",
			string(otherSnapshot.OnRamp.DynamicConfig.DynamicConfig.FeeAggregator), string(mySnapshot.OnRamp.DynamicConfig.DynamicConfig.FeeAggregator))
	}
	if string(otherSnapshot.OnRamp.DynamicConfig.DynamicConfig.AllowListAdmin) != string(mySnapshot.OnRamp.DynamicConfig.DynamicConfig.AllowListAdmin) {
		return fmt.Errorf("GetAllConfigsLegacy OnRamp.DynamicConfig.DynamicConfig.AllowListAdmin mismatch: got %s, expected %s",
			string(otherSnapshot.OnRamp.DynamicConfig.DynamicConfig.AllowListAdmin), string(mySnapshot.OnRamp.DynamicConfig.DynamicConfig.AllowListAdmin))
	}

	// Compare OnRamp DestChainConfig
	if otherSnapshot.OnRamp.DestChainConfig.SequenceNumber != mySnapshot.OnRamp.DestChainConfig.SequenceNumber {
		return fmt.Errorf("GetAllConfigsLegacy OnRamp.DestChainConfig.SequenceNumber mismatch: got %d, expected %d",
			otherSnapshot.OnRamp.DestChainConfig.SequenceNumber, mySnapshot.OnRamp.DestChainConfig.SequenceNumber)
	}
	if otherSnapshot.OnRamp.DestChainConfig.AllowListEnabled != mySnapshot.OnRamp.DestChainConfig.AllowListEnabled {
		return fmt.Errorf("GetAllConfigsLegacy OnRamp.DestChainConfig.AllowListEnabled mismatch: got %t, expected %t",
			otherSnapshot.OnRamp.DestChainConfig.AllowListEnabled, mySnapshot.OnRamp.DestChainConfig.AllowListEnabled)
	}
	if string(otherSnapshot.OnRamp.DestChainConfig.Router) != string(mySnapshot.OnRamp.DestChainConfig.Router) {
		return fmt.Errorf("GetAllConfigsLegacy OnRamp.DestChainConfig.Router mismatch: got %s, expected %s",
			string(otherSnapshot.OnRamp.DestChainConfig.Router), string(mySnapshot.OnRamp.DestChainConfig.Router))
	}

	// 6. Compare Router
	if string(otherSnapshot.Router.WrappedNativeAddress) != string(mySnapshot.Router.WrappedNativeAddress) {
		return fmt.Errorf("GetAllConfigsLegacy Router.WrappedNativeAddress mismatch: got %s, expected %s",
			string(otherSnapshot.Router.WrappedNativeAddress), string(mySnapshot.Router.WrappedNativeAddress))
	}

	// 7. Compare CurseInfo
	if otherSnapshot.CurseInfo.CursedDestination != mySnapshot.CurseInfo.CursedDestination {
		return fmt.Errorf("GetAllConfigsLegacy CurseInfo.CursedDestination mismatch: got %t, expected %t",
			otherSnapshot.CurseInfo.CursedDestination, mySnapshot.CurseInfo.CursedDestination)
	}
	if otherSnapshot.CurseInfo.GlobalCurse != mySnapshot.CurseInfo.GlobalCurse {
		return fmt.Errorf("GetAllConfigsLegacy CurseInfo.GlobalCurse mismatch: got %t, expected %t",
			otherSnapshot.CurseInfo.GlobalCurse, mySnapshot.CurseInfo.GlobalCurse)
	}
	if len(otherSnapshot.CurseInfo.CursedSourceChains) != len(mySnapshot.CurseInfo.CursedSourceChains) {
		return fmt.Errorf("GetAllConfigsLegacy CurseInfo.CursedSourceChains length mismatch: got %d, expected %d",
			len(otherSnapshot.CurseInfo.CursedSourceChains), len(mySnapshot.CurseInfo.CursedSourceChains))
	}
	for chainSel, myCursed := range mySnapshot.CurseInfo.CursedSourceChains {
		otherCursed, exists := otherSnapshot.CurseInfo.CursedSourceChains[chainSel]
		if !exists {
			return fmt.Errorf("GetAllConfigsLegacy CurseInfo.CursedSourceChains missing chain %d in other snapshot", chainSel)
		}
		if otherCursed != myCursed {
			return fmt.Errorf("GetAllConfigsLegacy CurseInfo.CursedSourceChains chain %d mismatch: got %t, expected %t",
				chainSel, otherCursed, myCursed)
		}
	}

	// Compare source chain configs with all fields
	if len(otherConfigs) != len(myConfigs) {
		return fmt.Errorf("GetAllConfigsLegacy configs length mismatch: got %d, expected %d", len(otherConfigs), len(myConfigs))
	}
	for chainSel, myConfig := range myConfigs {
		otherConfig, exists := otherConfigs[chainSel]
		if !exists {
			return fmt.Errorf("GetAllConfigsLegacy missing chain %d in other configs", chainSel)
		}
		if string(otherConfig.Router) != string(myConfig.Router) {
			return fmt.Errorf("GetAllConfigsLegacy config chain %d Router mismatch: got %s, expected %s",
				chainSel, string(otherConfig.Router), string(myConfig.Router))
		}
		if otherConfig.IsEnabled != myConfig.IsEnabled {
			return fmt.Errorf("GetAllConfigsLegacy config chain %d IsEnabled mismatch: got %t, expected %t",
				chainSel, otherConfig.IsEnabled, myConfig.IsEnabled)
		}
		if otherConfig.IsRMNVerificationDisabled != myConfig.IsRMNVerificationDisabled {
			return fmt.Errorf("GetAllConfigsLegacy config chain %d IsRMNVerificationDisabled mismatch: got %t, expected %t",
				chainSel, otherConfig.IsRMNVerificationDisabled, myConfig.IsRMNVerificationDisabled)
		}
		if otherConfig.MinSeqNr != myConfig.MinSeqNr {
			return fmt.Errorf("GetAllConfigsLegacy config chain %d MinSeqNr mismatch: got %d, expected %d",
				chainSel, otherConfig.MinSeqNr, myConfig.MinSeqNr)
		}
		if string(otherConfig.OnRamp) != string(myConfig.OnRamp) {
			return fmt.Errorf("GetAllConfigsLegacy config chain %d OnRamp mismatch: got %s, expected %s",
				chainSel, string(otherConfig.OnRamp), string(myConfig.OnRamp))
		}
	}
	return nil
}

func (s staticChainAccessor) evaluateGetChainFeeComponents(ctx context.Context, other ccipocr3.ChainAccessor) error {
	otherFeeComponents, err := other.GetChainFeeComponents(ctx)
	if err != nil {
		return fmt.Errorf("GetChainFeeComponents failed: %w", err)
	}
	myFeeComponents, err := s.GetChainFeeComponents(ctx)
	if err != nil {
		return fmt.Errorf("GetChainFeeComponents failed: %w", err)
	}
	if otherFeeComponents.ExecutionFee.Cmp(myFeeComponents.ExecutionFee) != 0 {
		return fmt.Errorf("GetChainFeeComponents ExecutionFee mismatch: got %s, expected %s",
			otherFeeComponents.ExecutionFee.String(), myFeeComponents.ExecutionFee.String())
	}
	if otherFeeComponents.DataAvailabilityFee.Cmp(myFeeComponents.DataAvailabilityFee) != 0 {
		return fmt.Errorf("GetChainFeeComponents DataAvailabilityFee mismatch: got %s, expected %s",
			otherFeeComponents.DataAvailabilityFee.String(), myFeeComponents.DataAvailabilityFee.String())
	}
	return nil
}

func (s staticChainAccessor) evaluateSync(ctx context.Context, other ccipocr3.ChainAccessor) error {
	err := other.Sync(ctx, "test-contract", ccipocr3.UnknownAddress("test-address"))
	if err != nil {
		return fmt.Errorf("Sync failed: %w", err)
	}
	return nil
}

func (s staticChainAccessor) evaluateCommitReportsGTETimestamp(ctx context.Context, other ccipocr3.ChainAccessor) error {
	testTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	otherReports, err := other.CommitReportsGTETimestamp(ctx, testTime, primitives.Finalized, 10)
	if err != nil {
		return fmt.Errorf("CommitReportsGTETimestamp failed: %w", err)
	}
	myReports, err := s.CommitReportsGTETimestamp(ctx, testTime, primitives.Finalized, 10)
	if err != nil {
		return fmt.Errorf("CommitReportsGTETimestamp failed: %w", err)
	}
	if len(otherReports) != len(myReports) {
		return fmt.Errorf("CommitReportsGTETimestamp length mismatch: got %d, expected %d", len(otherReports), len(myReports))
	}
	return nil
}

func (s staticChainAccessor) evaluateExecutedMessages(ctx context.Context, other ccipocr3.ChainAccessor) error {
	ranges := map[ccipocr3.ChainSelector][]ccipocr3.SeqNumRange{
		ccipocr3.ChainSelector(1): {ccipocr3.NewSeqNumRange(1, 10)},
		ccipocr3.ChainSelector(2): {ccipocr3.NewSeqNumRange(5, 15)},
	}
	otherMessages, err := other.ExecutedMessages(ctx, ranges, primitives.Finalized)
	if err != nil {
		return fmt.Errorf("ExecutedMessages failed: %w", err)
	}
	myMessages, err := s.ExecutedMessages(ctx, ranges, primitives.Finalized)
	if err != nil {
		return fmt.Errorf("ExecutedMessages failed: %w", err)
	}
	if len(otherMessages) != len(myMessages) {
		return fmt.Errorf("ExecutedMessages length mismatch: got %d, expected %d", len(otherMessages), len(myMessages))
	}
	return nil
}

func (s staticChainAccessor) evaluateNextSeqNum(ctx context.Context, other ccipocr3.ChainAccessor) error {
	sources := []ccipocr3.ChainSelector{1, 2}
	otherSeqNums, err := other.NextSeqNum(ctx, sources)
	if err != nil {
		return fmt.Errorf("NextSeqNum failed: %w", err)
	}
	mySeqNums, err := s.NextSeqNum(ctx, sources)
	if err != nil {
		return fmt.Errorf("NextSeqNum failed: %w", err)
	}
	if len(otherSeqNums) != len(mySeqNums) {
		return fmt.Errorf("NextSeqNum length mismatch: got %d, expected %d", len(otherSeqNums), len(mySeqNums))
	}
	for chainSel, mySeq := range mySeqNums {
		otherSeq, exists := otherSeqNums[chainSel]
		if !exists {
			return fmt.Errorf("NextSeqNum missing chain %d in other seq nums", chainSel)
		}
		if otherSeq != mySeq {
			return fmt.Errorf("NextSeqNum mismatch for chain %d: got %d, expected %d", chainSel, otherSeq, mySeq)
		}
	}
	return nil
}

func (s staticChainAccessor) evaluateNonces(ctx context.Context, other ccipocr3.ChainAccessor) error {
	addresses := map[ccipocr3.ChainSelector][]ccipocr3.UnknownEncodedAddress{
		ccipocr3.ChainSelector(1): {"sender1", "sender2"},
		ccipocr3.ChainSelector(2): {"sender3"},
	}
	otherNonces, err := other.Nonces(ctx, addresses)
	if err != nil {
		return fmt.Errorf("Nonces failed: %w", err)
	}
	myNonces, err := s.Nonces(ctx, addresses)
	if err != nil {
		return fmt.Errorf("Nonces failed: %w", err)
	}
	if len(otherNonces) != len(myNonces) {
		return fmt.Errorf("Nonces length mismatch: got %d, expected %d", len(otherNonces), len(myNonces))
	}
	for chainSel, myNoncesForChain := range myNonces {
		otherNoncesForChain, exists := otherNonces[chainSel]
		if !exists {
			return fmt.Errorf("Nonces missing chain %d in other nonces", chainSel)
		}
		if len(otherNoncesForChain) != len(myNoncesForChain) {
			return fmt.Errorf("Nonces chain %d length mismatch: got %d, expected %d", chainSel, len(otherNoncesForChain), len(myNoncesForChain))
		}
		for addr, myNonce := range myNoncesForChain {
			otherNonce, exists := otherNoncesForChain[addr]
			if !exists {
				return fmt.Errorf("Nonces chain %d missing address %s in other nonces", chainSel, string(addr))
			}
			if otherNonce != myNonce {
				return fmt.Errorf("Nonces chain %d address %s mismatch: got %d, expected %d", chainSel, string(addr), otherNonce, myNonce)
			}
		}
	}
	return nil
}

func (s staticChainAccessor) evaluateGetChainFeePriceUpdate(ctx context.Context, other ccipocr3.ChainAccessor) error {
	chains := []ccipocr3.ChainSelector{1, 2}
	otherUpdates, err := other.GetChainFeePriceUpdate(ctx, chains)
	if err != nil {
		return fmt.Errorf("GetChainFeePriceUpdate failed: %w", err)
	}
	myUpdates, err := s.GetChainFeePriceUpdate(ctx, chains)
	if err != nil {
		return fmt.Errorf("GetChainFeePriceUpdate failed: %w", err)
	}
	if len(otherUpdates) != len(myUpdates) {
		return fmt.Errorf("GetChainFeePriceUpdate length mismatch: got %d, expected %d", len(otherUpdates), len(myUpdates))
	}
	for chainSel, myUpdate := range myUpdates {
		otherUpdate, exists := otherUpdates[chainSel]
		if !exists {
			return fmt.Errorf("GetChainFeePriceUpdate missing chain %d in other updates", chainSel)
		}
		if otherUpdate.Value.Cmp(myUpdate.Value) != 0 {
			return fmt.Errorf("GetChainFeePriceUpdate chain %d Value mismatch: got %s, expected %s",
				chainSel, otherUpdate.Value.String(), myUpdate.Value.String())
		}
		if otherUpdate.Timestamp != myUpdate.Timestamp {
			return fmt.Errorf("GetChainFeePriceUpdate chain %d Timestamp mismatch: got %d, expected %d",
				chainSel, otherUpdate.Timestamp, myUpdate.Timestamp)
		}
	}
	return nil
}

func (s staticChainAccessor) evaluateGetLatestPriceSeqNr(ctx context.Context, other ccipocr3.ChainAccessor) error {
	otherSeqNr, err := other.GetLatestPriceSeqNr(ctx)
	if err != nil {
		return fmt.Errorf("GetLatestPriceSeqNr failed: %w", err)
	}
	mySeqNr, err := s.GetLatestPriceSeqNr(ctx)
	if err != nil {
		return fmt.Errorf("GetLatestPriceSeqNr failed: %w", err)
	}
	if otherSeqNr != mySeqNr {
		return fmt.Errorf("GetLatestPriceSeqNr mismatch: got %d, expected %d", otherSeqNr, mySeqNr)
	}
	return nil
}

func (s staticChainAccessor) evaluateMsgsBetweenSeqNums(ctx context.Context, other ccipocr3.ChainAccessor) error {
	seqRange := ccipocr3.NewSeqNumRange(1, 10)
	otherMsgs, err := other.MsgsBetweenSeqNums(ctx, ccipocr3.ChainSelector(2), seqRange)
	if err != nil {
		return fmt.Errorf("MsgsBetweenSeqNums failed: %w", err)
	}
	myMsgs, err := s.MsgsBetweenSeqNums(ctx, ccipocr3.ChainSelector(2), seqRange)
	if err != nil {
		return fmt.Errorf("MsgsBetweenSeqNums failed: %w", err)
	}
	if len(otherMsgs) != len(myMsgs) {
		return fmt.Errorf("MsgsBetweenSeqNums length mismatch: got %d, expected %d", len(otherMsgs), len(myMsgs))
	}
	return nil
}

func (s staticChainAccessor) evaluateLatestMessageTo(ctx context.Context, other ccipocr3.ChainAccessor) error {
	otherLatestSeq, err := other.LatestMessageTo(ctx, ccipocr3.ChainSelector(2))
	if err != nil {
		return fmt.Errorf("LatestMessageTo failed: %w", err)
	}
	myLatestSeq, err := s.LatestMessageTo(ctx, ccipocr3.ChainSelector(2))
	if err != nil {
		return fmt.Errorf("LatestMessageTo failed: %w", err)
	}
	if otherLatestSeq != myLatestSeq {
		return fmt.Errorf("LatestMessageTo mismatch: got %d, expected %d", otherLatestSeq, myLatestSeq)
	}
	return nil
}

func (s staticChainAccessor) evaluateGetExpectedNextSequenceNumber(ctx context.Context, other ccipocr3.ChainAccessor) error {
	otherExpectedSeq, err := other.GetExpectedNextSequenceNumber(ctx, ccipocr3.ChainSelector(2))
	if err != nil {
		return fmt.Errorf("GetExpectedNextSequenceNumber failed: %w", err)
	}
	myExpectedSeq, err := s.GetExpectedNextSequenceNumber(ctx, ccipocr3.ChainSelector(2))
	if err != nil {
		return fmt.Errorf("GetExpectedNextSequenceNumber failed: %w", err)
	}
	if otherExpectedSeq != myExpectedSeq {
		return fmt.Errorf("GetExpectedNextSequenceNumber mismatch: got %d, expected %d", otherExpectedSeq, myExpectedSeq)
	}
	return nil
}

func (s staticChainAccessor) evaluateGetTokenPriceUSD(ctx context.Context, other ccipocr3.ChainAccessor) error {
	otherPrice, err := other.GetTokenPriceUSD(ctx, ccipocr3.UnknownAddress("test-token"))
	if err != nil {
		return fmt.Errorf("GetTokenPriceUSD failed: %w", err)
	}
	myPrice, err := s.GetTokenPriceUSD(ctx, ccipocr3.UnknownAddress("test-token"))
	if err != nil {
		return fmt.Errorf("GetTokenPriceUSD failed: %w", err)
	}
	if otherPrice.Value.Cmp(myPrice.Value) != 0 {
		return fmt.Errorf("GetTokenPriceUSD Value mismatch: got %s, expected %s", otherPrice.Value.String(), myPrice.Value.String())
	}
	if otherPrice.Timestamp != myPrice.Timestamp {
		return fmt.Errorf("GetTokenPriceUSD Timestamp mismatch: got %d, expected %d", otherPrice.Timestamp, myPrice.Timestamp)
	}
	return nil
}

func (s staticChainAccessor) evaluateGetFeeQuoterDestChainConfig(ctx context.Context, other ccipocr3.ChainAccessor) error {
	otherConfig, err := other.GetFeeQuoterDestChainConfig(ctx, ccipocr3.ChainSelector(2))
	if err != nil {
		return fmt.Errorf("GetFeeQuoterDestChainConfig failed: %w", err)
	}
	myConfig, err := s.GetFeeQuoterDestChainConfig(ctx, ccipocr3.ChainSelector(2))
	if err != nil {
		return fmt.Errorf("GetFeeQuoterDestChainConfig failed: %w", err)
	}

	// Compare all FeeQuoterDestChainConfig fields
	if otherConfig.IsEnabled != myConfig.IsEnabled {
		return fmt.Errorf("GetFeeQuoterDestChainConfig IsEnabled mismatch: got %t, expected %t", otherConfig.IsEnabled, myConfig.IsEnabled)
	}
	if otherConfig.MaxNumberOfTokensPerMsg != myConfig.MaxNumberOfTokensPerMsg {
		return fmt.Errorf("GetFeeQuoterDestChainConfig MaxNumberOfTokensPerMsg mismatch: got %d, expected %d", otherConfig.MaxNumberOfTokensPerMsg, myConfig.MaxNumberOfTokensPerMsg)
	}
	if otherConfig.MaxDataBytes != myConfig.MaxDataBytes {
		return fmt.Errorf("GetFeeQuoterDestChainConfig MaxDataBytes mismatch: got %d, expected %d", otherConfig.MaxDataBytes, myConfig.MaxDataBytes)
	}
	if otherConfig.MaxPerMsgGasLimit != myConfig.MaxPerMsgGasLimit {
		return fmt.Errorf("GetFeeQuoterDestChainConfig MaxPerMsgGasLimit mismatch: got %d, expected %d", otherConfig.MaxPerMsgGasLimit, myConfig.MaxPerMsgGasLimit)
	}
	if otherConfig.DestGasOverhead != myConfig.DestGasOverhead {
		return fmt.Errorf("GetFeeQuoterDestChainConfig DestGasOverhead mismatch: got %d, expected %d", otherConfig.DestGasOverhead, myConfig.DestGasOverhead)
	}
	if otherConfig.DestGasPerPayloadByteBase != myConfig.DestGasPerPayloadByteBase {
		return fmt.Errorf("GetFeeQuoterDestChainConfig DestGasPerPayloadByteBase mismatch: got %d, expected %d", otherConfig.DestGasPerPayloadByteBase, myConfig.DestGasPerPayloadByteBase)
	}
	if otherConfig.DestGasPerPayloadByteHigh != myConfig.DestGasPerPayloadByteHigh {
		return fmt.Errorf("GetFeeQuoterDestChainConfig DestGasPerPayloadByteHigh mismatch: got %d, expected %d", otherConfig.DestGasPerPayloadByteHigh, myConfig.DestGasPerPayloadByteHigh)
	}
	if otherConfig.DestGasPerPayloadByteThreshold != myConfig.DestGasPerPayloadByteThreshold {
		return fmt.Errorf("GetFeeQuoterDestChainConfig DestGasPerPayloadByteThreshold mismatch: got %d, expected %d", otherConfig.DestGasPerPayloadByteThreshold, myConfig.DestGasPerPayloadByteThreshold)
	}
	if otherConfig.DestDataAvailabilityOverheadGas != myConfig.DestDataAvailabilityOverheadGas {
		return fmt.Errorf("GetFeeQuoterDestChainConfig DestDataAvailabilityOverheadGas mismatch: got %d, expected %d", otherConfig.DestDataAvailabilityOverheadGas, myConfig.DestDataAvailabilityOverheadGas)
	}
	if otherConfig.DestGasPerDataAvailabilityByte != myConfig.DestGasPerDataAvailabilityByte {
		return fmt.Errorf("GetFeeQuoterDestChainConfig DestGasPerDataAvailabilityByte mismatch: got %d, expected %d", otherConfig.DestGasPerDataAvailabilityByte, myConfig.DestGasPerDataAvailabilityByte)
	}
	if otherConfig.DestDataAvailabilityMultiplierBps != myConfig.DestDataAvailabilityMultiplierBps {
		return fmt.Errorf("GetFeeQuoterDestChainConfig DestDataAvailabilityMultiplierBps mismatch: got %d, expected %d", otherConfig.DestDataAvailabilityMultiplierBps, myConfig.DestDataAvailabilityMultiplierBps)
	}
	if otherConfig.DefaultTokenFeeUSDCents != myConfig.DefaultTokenFeeUSDCents {
		return fmt.Errorf("GetFeeQuoterDestChainConfig DefaultTokenFeeUSDCents mismatch: got %d, expected %d", otherConfig.DefaultTokenFeeUSDCents, myConfig.DefaultTokenFeeUSDCents)
	}
	if otherConfig.DefaultTokenDestGasOverhead != myConfig.DefaultTokenDestGasOverhead {
		return fmt.Errorf("GetFeeQuoterDestChainConfig DefaultTokenDestGasOverhead mismatch: got %d, expected %d", otherConfig.DefaultTokenDestGasOverhead, myConfig.DefaultTokenDestGasOverhead)
	}
	if otherConfig.DefaultTxGasLimit != myConfig.DefaultTxGasLimit {
		return fmt.Errorf("GetFeeQuoterDestChainConfig DefaultTxGasLimit mismatch: got %d, expected %d", otherConfig.DefaultTxGasLimit, myConfig.DefaultTxGasLimit)
	}
	if otherConfig.GasMultiplierWeiPerEth != myConfig.GasMultiplierWeiPerEth {
		return fmt.Errorf("GetFeeQuoterDestChainConfig GasMultiplierWeiPerEth mismatch: got %d, expected %d", otherConfig.GasMultiplierWeiPerEth, myConfig.GasMultiplierWeiPerEth)
	}
	if otherConfig.NetworkFeeUSDCents != myConfig.NetworkFeeUSDCents {
		return fmt.Errorf("GetFeeQuoterDestChainConfig NetworkFeeUSDCents mismatch: got %d, expected %d", otherConfig.NetworkFeeUSDCents, myConfig.NetworkFeeUSDCents)
	}
	if otherConfig.GasPriceStalenessThreshold != myConfig.GasPriceStalenessThreshold {
		return fmt.Errorf("GetFeeQuoterDestChainConfig GasPriceStalenessThreshold mismatch: got %d, expected %d", otherConfig.GasPriceStalenessThreshold, myConfig.GasPriceStalenessThreshold)
	}
	if otherConfig.EnforceOutOfOrder != myConfig.EnforceOutOfOrder {
		return fmt.Errorf("GetFeeQuoterDestChainConfig EnforceOutOfOrder mismatch: got %t, expected %t", otherConfig.EnforceOutOfOrder, myConfig.EnforceOutOfOrder)
	}
	if otherConfig.ChainFamilySelector != myConfig.ChainFamilySelector {
		return fmt.Errorf("GetFeeQuoterDestChainConfig ChainFamilySelector mismatch: got %x, expected %x", otherConfig.ChainFamilySelector, myConfig.ChainFamilySelector)
	}
	return nil
}

func (s staticChainAccessor) evaluateMessagesByTokenID(ctx context.Context, other ccipocr3.ChainAccessor) error {
	tokens := map[ccipocr3.MessageTokenID]ccipocr3.RampTokenAmount{
		ccipocr3.NewMessageTokenID(1, 0): {
			SourcePoolAddress: ccipocr3.UnknownAddress("test-source-pool"),
			DestTokenAddress:  ccipocr3.UnknownAddress("test-dest-token"),
			ExtraData:         ccipocr3.Bytes("test-extra-data"),
			Amount:            ccipocr3.NewBigInt(big.NewInt(12345)),
		},
	}

	otherMessages, err := other.MessagesByTokenID(ctx, ccipocr3.ChainSelector(1), ccipocr3.ChainSelector(2), tokens)
	if err != nil {
		return fmt.Errorf("MessagesByTokenID failed: %w", err)
	}
	myMessages, err := s.MessagesByTokenID(ctx, ccipocr3.ChainSelector(1), ccipocr3.ChainSelector(2), tokens)
	if err != nil {
		return fmt.Errorf("MessagesByTokenID failed: %w", err)
	}

	if len(otherMessages) != len(myMessages) {
		return fmt.Errorf("MessagesByTokenID length mismatch: got %d, expected %d", len(otherMessages), len(myMessages))
	}

	for tokenID, myMessage := range myMessages {
		otherMessage, exists := otherMessages[tokenID]
		if !exists {
			return fmt.Errorf("MessagesByTokenID missing tokenID %s in other messages", tokenID.String())
		}
		if string(otherMessage) != string(myMessage) {
			return fmt.Errorf("MessagesByTokenID tokenID %s mismatch: got %s, expected %s", tokenID.String(), string(otherMessage), string(myMessage))
		}
	}
	return nil
}

func (s staticChainAccessor) evaluateGetFeedPricesUSD(ctx context.Context, other ccipocr3.ChainAccessor) error {
	tokens := []ccipocr3.UnknownEncodedAddress{"token1", "token2", "token3"}
	tokenInfo := map[ccipocr3.UnknownEncodedAddress]ccipocr3.TokenInfo{
		"token1": {
			AggregatorAddress: ccipocr3.UnknownEncodedAddress("0x1234567890123456789012345678901234567890"),
			DeviationPPB:      ccipocr3.NewBigInt(big.NewInt(1000000000)), // 1%
			Decimals:          18,
		},
	}

	otherPrices, err := other.GetFeedPricesUSD(ctx, tokens, tokenInfo)
	if err != nil {
		return fmt.Errorf("GetFeedPricesUSD failed: %w", err)
	}
	myPrices, err := s.GetFeedPricesUSD(ctx, tokens, tokenInfo)
	if err != nil {
		return fmt.Errorf("GetFeedPricesUSD failed: %w", err)
	}

	if len(otherPrices) != len(myPrices) {
		return fmt.Errorf("GetFeedPricesUSD length mismatch: got %d, expected %d", len(otherPrices), len(myPrices))
	}

	for token, myPrice := range myPrices {
		otherPrice, exists := otherPrices[token]
		if !exists {
			return fmt.Errorf("GetFeedPricesUSD missing token %s in other prices", string(token))
		}
		if otherPrice.Cmp(myPrice.Int) != 0 {
			return fmt.Errorf("GetFeedPricesUSD token %s mismatch: got %s, expected %s", string(token), otherPrice.String(), myPrice.String())
		}
	}
	return nil
}

func (s staticChainAccessor) evaluateGetFeeQuoterTokenUpdates(ctx context.Context, other ccipocr3.ChainAccessor) error {
	tokens := []ccipocr3.UnknownEncodedAddress{"token1", "token2"}
	chain := ccipocr3.ChainSelector(1)

	otherUpdates, err := other.GetFeeQuoterTokenUpdates(ctx, tokens, chain)
	if err != nil {
		return fmt.Errorf("GetFeeQuoterTokenUpdates failed: %w", err)
	}
	myUpdates, err := s.GetFeeQuoterTokenUpdates(ctx, tokens, chain)
	if err != nil {
		return fmt.Errorf("GetFeeQuoterTokenUpdates failed: %w", err)
	}

	if len(otherUpdates) != len(myUpdates) {
		return fmt.Errorf("GetFeeQuoterTokenUpdates length mismatch: got %d, expected %d", len(otherUpdates), len(myUpdates))
	}

	for token, myUpdate := range myUpdates {
		otherUpdate, exists := otherUpdates[token]
		if !exists {
			return fmt.Errorf("GetFeeQuoterTokenUpdates missing token %s in other updates", string(token))
		}
		if otherUpdate.Value.Cmp(myUpdate.Value) != 0 {
			return fmt.Errorf("GetFeeQuoterTokenUpdates token %s value mismatch: got %s, expected %s", string(token), otherUpdate.Value.String(), myUpdate.Value.String())
		}
		if otherUpdate.Timestamp != myUpdate.Timestamp {
			return fmt.Errorf("GetFeeQuoterTokenUpdates token %s timestamp mismatch: got %d, expected %d", string(token), otherUpdate.Timestamp, myUpdate.Timestamp)
		}
	}
	return nil
}

// AssertEqual implements ChainAccessorTester.
func (s staticChainAccessor) AssertEqual(ctx context.Context, t *testing.T, other ccipocr3.ChainAccessor) {
	t.Run("ChainAccessor", func(t *testing.T) {
		assert.NoError(t, s.Evaluate(ctx, other))
	})
}
