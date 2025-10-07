package test

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	testtypes "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/ccipocr3"
)

// ChainSpecificAddressCodec implementation

type ChainSpecificAddressCodecEvaluator interface {
	ccipocr3.ChainSpecificAddressCodec
	testtypes.Evaluator[ccipocr3.ChainSpecificAddressCodec]
}

type ChainSpecificAddressCodecTester interface {
	ccipocr3.ChainSpecificAddressCodec
	testtypes.Evaluator[ccipocr3.ChainSpecificAddressCodec]
	testtypes.AssertEqualer[ccipocr3.ChainSpecificAddressCodec]
}

func ChainSpecificAddressCodec(lggr logger.Logger) staticChainSpecificAddressCodec {
	return newStaticChainSpecificAddressCodec(lggr, staticChainSpecificAddressCodecConfig{
		addressBytesToString:     "test-address",
		addressStringToBytes:     []byte("test-address"),
		oracleIDAsAddressBytes:   []byte{42},
		transmitterBytesToString: "test-transmitter",
	})
}

var _ ChainSpecificAddressCodecTester = staticChainSpecificAddressCodec{}

type staticChainSpecificAddressCodecConfig struct {
	addressBytesToString     string
	addressStringToBytes     []byte
	oracleIDAsAddressBytes   []byte
	transmitterBytesToString string
}

type staticChainSpecificAddressCodec struct {
	staticChainSpecificAddressCodecConfig
}

func newStaticChainSpecificAddressCodec(lggr logger.Logger, cfg staticChainSpecificAddressCodecConfig) staticChainSpecificAddressCodec {
	lggr = logger.Named(lggr, "staticChainSpecificAddressCodec")
	return staticChainSpecificAddressCodec{
		staticChainSpecificAddressCodecConfig: cfg,
	}
}

func (s staticChainSpecificAddressCodec) AddressBytesToString(addr []byte) (string, error) {
	return s.addressBytesToString, nil
}

func (s staticChainSpecificAddressCodec) AddressStringToBytes(addr string) ([]byte, error) {
	return s.addressStringToBytes, nil
}

func (s staticChainSpecificAddressCodec) OracleIDAsAddressBytes(oracleID uint8) ([]byte, error) {
	return s.oracleIDAsAddressBytes, nil
}

func (s staticChainSpecificAddressCodec) TransmitterBytesToString(transmitter []byte) (string, error) {
	return s.transmitterBytesToString, nil
}

func (s staticChainSpecificAddressCodec) Evaluate(ctx context.Context, other ccipocr3.ChainSpecificAddressCodec) error {
	// Test AddressBytesToString
	otherResult, err := other.AddressBytesToString([]byte("test"))
	if err != nil {
		return fmt.Errorf("AddressBytesToString failed: %w", err)
	}
	myResult, err := s.AddressBytesToString([]byte("test"))
	if err != nil {
		return fmt.Errorf("AddressBytesToString failed: %w", err)
	}
	if otherResult != myResult {
		return fmt.Errorf("AddressBytesToString mismatch: got %s, expected %s", otherResult, myResult)
	}

	// Test AddressStringToBytes
	otherBytes, err := other.AddressStringToBytes("test")
	if err != nil {
		return fmt.Errorf("AddressStringToBytes failed: %w", err)
	}
	myBytes, err := s.AddressStringToBytes("test")
	if err != nil {
		return fmt.Errorf("AddressStringToBytes failed: %w", err)
	}
	if string(otherBytes) != string(myBytes) {
		return fmt.Errorf("AddressStringToBytes mismatch: got %s, expected %s", string(otherBytes), string(myBytes))
	}

	// Test OracleIDAsAddressBytes
	otherOracleAddr, err := other.OracleIDAsAddressBytes(42)
	if err != nil {
		return fmt.Errorf("OracleIDAsAddressBytes failed: %w", err)
	}
	myOracleAddr, err := s.OracleIDAsAddressBytes(42)
	if err != nil {
		return fmt.Errorf("OracleIDAsAddressBytes failed: %w", err)
	}
	if string(otherOracleAddr) != string(myOracleAddr) {
		return fmt.Errorf("OracleIDAsAddressBytes mismatch: got %s, expected %s", string(otherOracleAddr), string(myOracleAddr))
	}

	// Test TransmitterBytesToString
	otherTransmitter, err := other.TransmitterBytesToString([]byte("test"))
	if err != nil {
		return fmt.Errorf("TransmitterBytesToString failed: %w", err)
	}
	myTransmitter, err := s.TransmitterBytesToString([]byte("test"))
	if err != nil {
		return fmt.Errorf("TransmitterBytesToString failed: %w", err)
	}
	if otherTransmitter != myTransmitter {
		return fmt.Errorf("TransmitterBytesToString mismatch: got %s, expected %s", otherTransmitter, myTransmitter)
	}

	return nil
}

func (s staticChainSpecificAddressCodec) AssertEqual(ctx context.Context, t *testing.T, other ccipocr3.ChainSpecificAddressCodec) {
	t.Run("ChainSpecificAddressCodec", func(t *testing.T) {
		assert.NoError(t, s.Evaluate(ctx, other))
	})
}

// CommitPluginCodec implementation

type CommitPluginCodecEvaluator interface {
	ccipocr3.CommitPluginCodec
	testtypes.Evaluator[ccipocr3.CommitPluginCodec]
}

type CommitPluginCodecTester interface {
	ccipocr3.CommitPluginCodec
	testtypes.Evaluator[ccipocr3.CommitPluginCodec]
	testtypes.AssertEqualer[ccipocr3.CommitPluginCodec]
}

func CommitPluginCodec(lggr logger.Logger) staticCommitPluginCodec {
	return newStaticCommitPluginCodec(lggr, staticCommitPluginCodecConfig{
		encodedData: []byte("encoded-commit-report"),
		report:      ccipocr3.CommitPluginReport{},
	})
}

var _ CommitPluginCodecTester = staticCommitPluginCodec{}

type staticCommitPluginCodecConfig struct {
	encodedData []byte
	report      ccipocr3.CommitPluginReport
}

type staticCommitPluginCodec struct {
	staticCommitPluginCodecConfig
}

func newStaticCommitPluginCodec(lggr logger.Logger, cfg staticCommitPluginCodecConfig) staticCommitPluginCodec {
	lggr = logger.Named(lggr, "staticCommitPluginCodec")
	return staticCommitPluginCodec{
		staticCommitPluginCodecConfig: cfg,
	}
}

func (s staticCommitPluginCodec) Encode(ctx context.Context, report ccipocr3.CommitPluginReport) ([]byte, error) {
	return s.encodedData, nil
}

func (s staticCommitPluginCodec) Decode(ctx context.Context, data []byte) (ccipocr3.CommitPluginReport, error) {
	return s.report, nil
}

func (s staticCommitPluginCodec) Evaluate(ctx context.Context, other ccipocr3.CommitPluginCodec) error {
	// Test Encode
	otherEncoded, err := other.Encode(ctx, s.report)
	if err != nil {
		return fmt.Errorf("CommitPluginCodec Encode failed: %w", err)
	}
	myEncoded, err := s.Encode(ctx, s.report)
	if err != nil {
		return fmt.Errorf("CommitPluginCodec Encode failed: %w", err)
	}
	if string(otherEncoded) != string(myEncoded) {
		return fmt.Errorf("CommitPluginCodec Encode mismatch: got %s, expected %s", string(otherEncoded), string(myEncoded))
	}

	// Test Decode with round-trip verification
	otherDecoded, err := other.Decode(ctx, otherEncoded)
	if err != nil {
		return fmt.Errorf("CommitPluginCodec Decode failed: %w", err)
	}
	myDecoded, err := s.Decode(ctx, myEncoded)
	if err != nil {
		return fmt.Errorf("CommitPluginCodec Decode failed: %w", err)
	}

	// Deep compare all decoded report fields

	// Compare PriceUpdates.TokenPriceUpdates
	if len(otherDecoded.PriceUpdates.TokenPriceUpdates) != len(myDecoded.PriceUpdates.TokenPriceUpdates) {
		return fmt.Errorf("CommitPluginCodec Decode TokenPriceUpdates length mismatch: got %d, expected %d",
			len(otherDecoded.PriceUpdates.TokenPriceUpdates), len(myDecoded.PriceUpdates.TokenPriceUpdates))
	}
	for i, otherToken := range otherDecoded.PriceUpdates.TokenPriceUpdates {
		myToken := myDecoded.PriceUpdates.TokenPriceUpdates[i]
		if string(otherToken.TokenID) != string(myToken.TokenID) {
			return fmt.Errorf("CommitPluginCodec Decode TokenPriceUpdates[%d] TokenID mismatch: got %s, expected %s",
				i, string(otherToken.TokenID), string(myToken.TokenID))
		}
		if otherToken.Price.Cmp(myToken.Price.Int) != 0 {
			return fmt.Errorf("CommitPluginCodec Decode TokenPriceUpdates[%d] Price mismatch: got %s, expected %s",
				i, otherToken.Price.String(), myToken.Price.String())
		}
	}

	// Compare PriceUpdates.GasPriceUpdates
	if len(otherDecoded.PriceUpdates.GasPriceUpdates) != len(myDecoded.PriceUpdates.GasPriceUpdates) {
		return fmt.Errorf("CommitPluginCodec Decode GasPriceUpdates length mismatch: got %d, expected %d",
			len(otherDecoded.PriceUpdates.GasPriceUpdates), len(myDecoded.PriceUpdates.GasPriceUpdates))
	}
	for i, otherGas := range otherDecoded.PriceUpdates.GasPriceUpdates {
		myGas := myDecoded.PriceUpdates.GasPriceUpdates[i]
		if otherGas.ChainSel != myGas.ChainSel {
			return fmt.Errorf("CommitPluginCodec Decode GasPriceUpdates[%d] ChainSel mismatch: got %d, expected %d",
				i, otherGas.ChainSel, myGas.ChainSel)
		}
		if otherGas.GasPrice.Cmp(myGas.GasPrice.Int) != 0 {
			return fmt.Errorf("CommitPluginCodec Decode GasPriceUpdates[%d] GasPrice mismatch: got %s, expected %s",
				i, otherGas.GasPrice.String(), myGas.GasPrice.String())
		}
	}

	// Compare BlessedMerkleRoots
	if len(otherDecoded.BlessedMerkleRoots) != len(myDecoded.BlessedMerkleRoots) {
		return fmt.Errorf("CommitPluginCodec Decode BlessedMerkleRoots length mismatch: got %d, expected %d",
			len(otherDecoded.BlessedMerkleRoots), len(myDecoded.BlessedMerkleRoots))
	}
	for i, otherRoot := range otherDecoded.BlessedMerkleRoots {
		myRoot := myDecoded.BlessedMerkleRoots[i]
		if !otherRoot.Equals(myRoot) {
			return fmt.Errorf("CommitPluginCodec Decode BlessedMerkleRoots[%d] mismatch: got %v, expected %v",
				i, otherRoot, myRoot)
		}
	}

	// Compare UnblessedMerkleRoots
	if len(otherDecoded.UnblessedMerkleRoots) != len(myDecoded.UnblessedMerkleRoots) {
		return fmt.Errorf("CommitPluginCodec Decode UnblessedMerkleRoots length mismatch: got %d, expected %d",
			len(otherDecoded.UnblessedMerkleRoots), len(myDecoded.UnblessedMerkleRoots))
	}
	for i, otherRoot := range otherDecoded.UnblessedMerkleRoots {
		myRoot := myDecoded.UnblessedMerkleRoots[i]
		if !otherRoot.Equals(myRoot) {
			return fmt.Errorf("CommitPluginCodec Decode UnblessedMerkleRoots[%d] mismatch: got %v, expected %v",
				i, otherRoot, myRoot)
		}
	}

	// Compare RMNSignatures
	if len(otherDecoded.RMNSignatures) != len(myDecoded.RMNSignatures) {
		return fmt.Errorf("CommitPluginCodec Decode RMNSignatures length mismatch: got %d, expected %d",
			len(otherDecoded.RMNSignatures), len(myDecoded.RMNSignatures))
	}
	for i, otherSig := range otherDecoded.RMNSignatures {
		mySig := myDecoded.RMNSignatures[i]
		if otherSig.R != mySig.R {
			return fmt.Errorf("CommitPluginCodec Decode RMNSignatures[%d] R mismatch: got %x, expected %x",
				i, otherSig.R, mySig.R)
		}
		if otherSig.S != mySig.S {
			return fmt.Errorf("CommitPluginCodec Decode RMNSignatures[%d] S mismatch: got %x, expected %x",
				i, otherSig.S, mySig.S)
		}
	}

	return nil
}

func (s staticCommitPluginCodec) AssertEqual(ctx context.Context, t *testing.T, other ccipocr3.CommitPluginCodec) {
	t.Run("CommitPluginCodec", func(t *testing.T) {
		assert.NoError(t, s.Evaluate(ctx, other))
	})
}

// ExecutePluginCodec implementation

type ExecutePluginCodecEvaluator interface {
	ccipocr3.ExecutePluginCodec
	testtypes.Evaluator[ccipocr3.ExecutePluginCodec]
}

type ExecutePluginCodecTester interface {
	ccipocr3.ExecutePluginCodec
	testtypes.Evaluator[ccipocr3.ExecutePluginCodec]
	testtypes.AssertEqualer[ccipocr3.ExecutePluginCodec]
}

func ExecutePluginCodec(lggr logger.Logger) staticExecutePluginCodec {
	return newStaticExecutePluginCodec(lggr, staticExecutePluginCodecConfig{
		encodedData: []byte("encoded-execute-report"),
		report:      ccipocr3.ExecutePluginReport{},
	})
}

var _ ExecutePluginCodecTester = staticExecutePluginCodec{}

type staticExecutePluginCodecConfig struct {
	encodedData []byte
	report      ccipocr3.ExecutePluginReport
}

type staticExecutePluginCodec struct {
	staticExecutePluginCodecConfig
}

func newStaticExecutePluginCodec(lggr logger.Logger, cfg staticExecutePluginCodecConfig) staticExecutePluginCodec {
	lggr = logger.Named(lggr, "staticExecutePluginCodec")
	return staticExecutePluginCodec{
		staticExecutePluginCodecConfig: cfg,
	}
}

func (s staticExecutePluginCodec) Encode(ctx context.Context, report ccipocr3.ExecutePluginReport) ([]byte, error) {
	return s.encodedData, nil
}

func (s staticExecutePluginCodec) Decode(ctx context.Context, data []byte) (ccipocr3.ExecutePluginReport, error) {
	return s.report, nil
}

func (s staticExecutePluginCodec) Evaluate(ctx context.Context, other ccipocr3.ExecutePluginCodec) error {
	// Test Encode
	otherEncoded, err := other.Encode(ctx, s.report)
	if err != nil {
		return fmt.Errorf("ExecutePluginCodec Encode failed: %w", err)
	}
	myEncoded, err := s.Encode(ctx, s.report)
	if err != nil {
		return fmt.Errorf("ExecutePluginCodec Encode failed: %w", err)
	}
	if string(otherEncoded) != string(myEncoded) {
		return fmt.Errorf("ExecutePluginCodec Encode mismatch: got %s, expected %s", string(otherEncoded), string(myEncoded))
	}

	// Test Decode with round-trip verification
	otherDecoded, err := other.Decode(ctx, otherEncoded)
	if err != nil {
		return fmt.Errorf("ExecutePluginCodec Decode failed: %w", err)
	}
	myDecoded, err := s.Decode(ctx, myEncoded)
	if err != nil {
		return fmt.Errorf("ExecutePluginCodec Decode failed: %w", err)
	}

	// Deep compare all decoded report fields

	// Compare ChainReports
	if len(otherDecoded.ChainReports) != len(myDecoded.ChainReports) {
		return fmt.Errorf("ExecutePluginCodec Decode ChainReports length mismatch: got %d, expected %d",
			len(otherDecoded.ChainReports), len(myDecoded.ChainReports))
	}
	for i, otherChainReport := range otherDecoded.ChainReports {
		myChainReport := myDecoded.ChainReports[i]

		// Compare SourceChainSelector
		if otherChainReport.SourceChainSelector != myChainReport.SourceChainSelector {
			return fmt.Errorf("ExecutePluginCodec Decode ChainReports[%d] SourceChainSelector mismatch: got %d, expected %d",
				i, otherChainReport.SourceChainSelector, myChainReport.SourceChainSelector)
		}

		// Compare Messages
		if len(otherChainReport.Messages) != len(myChainReport.Messages) {
			return fmt.Errorf("ExecutePluginCodec Decode ChainReports[%d] Messages length mismatch: got %d, expected %d",
				i, len(otherChainReport.Messages), len(myChainReport.Messages))
		}
		for j, otherMsg := range otherChainReport.Messages {
			myMsg := myChainReport.Messages[j]

			// Compare Message Header
			if otherMsg.Header.MessageID != myMsg.Header.MessageID {
				return fmt.Errorf("ExecutePluginCodec Decode ChainReports[%d] Messages[%d] Header.MessageID mismatch: got %x, expected %x",
					i, j, otherMsg.Header.MessageID, myMsg.Header.MessageID)
			}
			if otherMsg.Header.SourceChainSelector != myMsg.Header.SourceChainSelector {
				return fmt.Errorf("ExecutePluginCodec Decode ChainReports[%d] Messages[%d] Header.SourceChainSelector mismatch: got %d, expected %d",
					i, j, otherMsg.Header.SourceChainSelector, myMsg.Header.SourceChainSelector)
			}
			if otherMsg.Header.DestChainSelector != myMsg.Header.DestChainSelector {
				return fmt.Errorf("ExecutePluginCodec Decode ChainReports[%d] Messages[%d] Header.DestChainSelector mismatch: got %d, expected %d",
					i, j, otherMsg.Header.DestChainSelector, myMsg.Header.DestChainSelector)
			}
			if otherMsg.Header.SequenceNumber != myMsg.Header.SequenceNumber {
				return fmt.Errorf("ExecutePluginCodec Decode ChainReports[%d] Messages[%d] Header.SequenceNumber mismatch: got %d, expected %d",
					i, j, otherMsg.Header.SequenceNumber, myMsg.Header.SequenceNumber)
			}
			if otherMsg.Header.Nonce != myMsg.Header.Nonce {
				return fmt.Errorf("ExecutePluginCodec Decode ChainReports[%d] Messages[%d] Header.Nonce mismatch: got %d, expected %d",
					i, j, otherMsg.Header.Nonce, myMsg.Header.Nonce)
			}
			if otherMsg.Header.TxHash != myMsg.Header.TxHash {
				return fmt.Errorf("ExecutePluginCodec Decode ChainReports[%d] Messages[%d] Header.TxHash mismatch: got %s, expected %s",
					i, j, otherMsg.Header.TxHash, myMsg.Header.TxHash)
			}
			if string(otherMsg.Header.OnRamp) != string(myMsg.Header.OnRamp) {
				return fmt.Errorf("ExecutePluginCodec Decode ChainReports[%d] Messages[%d] Header.OnRamp mismatch: got %s, expected %s",
					i, j, string(otherMsg.Header.OnRamp), string(myMsg.Header.OnRamp))
			}

			// Compare Message fields
			if string(otherMsg.Sender) != string(myMsg.Sender) {
				return fmt.Errorf("ExecutePluginCodec Decode ChainReports[%d] Messages[%d] Sender mismatch: got %s, expected %s",
					i, j, string(otherMsg.Sender), string(myMsg.Sender))
			}
			if string(otherMsg.Data) != string(myMsg.Data) {
				return fmt.Errorf("ExecutePluginCodec Decode ChainReports[%d] Messages[%d] Data mismatch: got %s, expected %s",
					i, j, string(otherMsg.Data), string(myMsg.Data))
			}
			if string(otherMsg.Receiver) != string(myMsg.Receiver) {
				return fmt.Errorf("ExecutePluginCodec Decode ChainReports[%d] Messages[%d] Receiver mismatch: got %s, expected %s",
					i, j, string(otherMsg.Receiver), string(myMsg.Receiver))
			}
			if string(otherMsg.ExtraArgs) != string(myMsg.ExtraArgs) {
				return fmt.Errorf("ExecutePluginCodec Decode ChainReports[%d] Messages[%d] ExtraArgs mismatch: got %s, expected %s",
					i, j, string(otherMsg.ExtraArgs), string(myMsg.ExtraArgs))
			}
			if string(otherMsg.FeeToken) != string(myMsg.FeeToken) {
				return fmt.Errorf("ExecutePluginCodec Decode ChainReports[%d] Messages[%d] FeeToken mismatch: got %s, expected %s",
					i, j, string(otherMsg.FeeToken), string(myMsg.FeeToken))
			}
			if otherMsg.FeeTokenAmount.Cmp(myMsg.FeeTokenAmount.Int) != 0 {
				return fmt.Errorf("ExecutePluginCodec Decode ChainReports[%d] Messages[%d] FeeTokenAmount mismatch: got %s, expected %s",
					i, j, otherMsg.FeeTokenAmount.String(), myMsg.FeeTokenAmount.String())
			}
			if otherMsg.FeeValueJuels.Cmp(myMsg.FeeValueJuels.Int) != 0 {
				return fmt.Errorf("ExecutePluginCodec Decode ChainReports[%d] Messages[%d] FeeValueJuels mismatch: got %s, expected %s",
					i, j, otherMsg.FeeValueJuels.String(), myMsg.FeeValueJuels.String())
			}

			// Compare TokenAmounts
			if len(otherMsg.TokenAmounts) != len(myMsg.TokenAmounts) {
				return fmt.Errorf("ExecutePluginCodec Decode ChainReports[%d] Messages[%d] TokenAmounts length mismatch: got %d, expected %d",
					i, j, len(otherMsg.TokenAmounts), len(myMsg.TokenAmounts))
			}
			for k, otherTokenAmount := range otherMsg.TokenAmounts {
				myTokenAmount := myMsg.TokenAmounts[k]
				if string(otherTokenAmount.SourcePoolAddress) != string(myTokenAmount.SourcePoolAddress) {
					return fmt.Errorf("ExecutePluginCodec Decode ChainReports[%d] Messages[%d] TokenAmounts[%d] SourcePoolAddress mismatch: got %s, expected %s",
						i, j, k, string(otherTokenAmount.SourcePoolAddress), string(myTokenAmount.SourcePoolAddress))
				}
				if string(otherTokenAmount.DestTokenAddress) != string(myTokenAmount.DestTokenAddress) {
					return fmt.Errorf("ExecutePluginCodec Decode ChainReports[%d] Messages[%d] TokenAmounts[%d] DestTokenAddress mismatch: got %s, expected %s",
						i, j, k, string(otherTokenAmount.DestTokenAddress), string(myTokenAmount.DestTokenAddress))
				}
				if string(otherTokenAmount.ExtraData) != string(myTokenAmount.ExtraData) {
					return fmt.Errorf("ExecutePluginCodec Decode ChainReports[%d] Messages[%d] TokenAmounts[%d] ExtraData mismatch: got %s, expected %s",
						i, j, k, string(otherTokenAmount.ExtraData), string(myTokenAmount.ExtraData))
				}
				if otherTokenAmount.Amount.Cmp(myTokenAmount.Amount.Int) != 0 {
					return fmt.Errorf("ExecutePluginCodec Decode ChainReports[%d] Messages[%d] TokenAmounts[%d] Amount mismatch: got %s, expected %s",
						i, j, k, otherTokenAmount.Amount.String(), myTokenAmount.Amount.String())
				}
				if string(otherTokenAmount.DestExecData) != string(myTokenAmount.DestExecData) {
					return fmt.Errorf("ExecutePluginCodec Decode ChainReports[%d] Messages[%d] TokenAmounts[%d] DestExecData mismatch: got %s, expected %s",
						i, j, k, string(otherTokenAmount.DestExecData), string(myTokenAmount.DestExecData))
				}
			}
		}

		// Compare OffchainTokenData
		if len(otherChainReport.OffchainTokenData) != len(myChainReport.OffchainTokenData) {
			return fmt.Errorf("ExecutePluginCodec Decode ChainReports[%d] OffchainTokenData length mismatch: got %d, expected %d",
				i, len(otherChainReport.OffchainTokenData), len(myChainReport.OffchainTokenData))
		}
		for j, otherOffchainData := range otherChainReport.OffchainTokenData {
			myOffchainData := myChainReport.OffchainTokenData[j]
			if len(otherOffchainData) != len(myOffchainData) {
				return fmt.Errorf("ExecutePluginCodec Decode ChainReports[%d] OffchainTokenData[%d] length mismatch: got %d, expected %d",
					i, j, len(otherOffchainData), len(myOffchainData))
			}
			for k, otherTokenData := range otherOffchainData {
				myTokenData := myOffchainData[k]
				if string(otherTokenData) != string(myTokenData) {
					return fmt.Errorf("ExecutePluginCodec Decode ChainReports[%d] OffchainTokenData[%d][%d] mismatch: got %s, expected %s",
						i, j, k, string(otherTokenData), string(myTokenData))
				}
			}
		}

		// Compare Proofs
		if len(otherChainReport.Proofs) != len(myChainReport.Proofs) {
			return fmt.Errorf("ExecutePluginCodec Decode ChainReports[%d] Proofs length mismatch: got %d, expected %d",
				i, len(otherChainReport.Proofs), len(myChainReport.Proofs))
		}
		for j, otherProof := range otherChainReport.Proofs {
			myProof := myChainReport.Proofs[j]
			if otherProof != myProof {
				return fmt.Errorf("ExecutePluginCodec Decode ChainReports[%d] Proofs[%d] mismatch: got %x, expected %x",
					i, j, otherProof, myProof)
			}
		}

		// Compare ProofFlagBits
		if otherChainReport.ProofFlagBits.Cmp(myChainReport.ProofFlagBits.Int) != 0 {
			return fmt.Errorf("ExecutePluginCodec Decode ChainReports[%d] ProofFlagBits mismatch: got %s, expected %s",
				i, otherChainReport.ProofFlagBits.String(), myChainReport.ProofFlagBits.String())
		}
	}

	return nil
}

func (s staticExecutePluginCodec) AssertEqual(ctx context.Context, t *testing.T, other ccipocr3.ExecutePluginCodec) {
	t.Run("ExecutePluginCodec", func(t *testing.T) {
		assert.NoError(t, s.Evaluate(ctx, other))
	})
}

// TokenDataEncoder implementation

type TokenDataEncoderEvaluator interface {
	ccipocr3.TokenDataEncoder
	testtypes.Evaluator[ccipocr3.TokenDataEncoder]
}

type TokenDataEncoderTester interface {
	ccipocr3.TokenDataEncoder
	testtypes.Evaluator[ccipocr3.TokenDataEncoder]
	testtypes.AssertEqualer[ccipocr3.TokenDataEncoder]
}

func TokenDataEncoder(lggr logger.Logger) staticTokenDataEncoder {
	return newStaticTokenDataEncoder(lggr, staticTokenDataEncoderConfig{
		encodedUSDC: ccipocr3.Bytes("encoded-usdc"),
	})
}

var _ TokenDataEncoderTester = staticTokenDataEncoder{}

type staticTokenDataEncoderConfig struct {
	encodedUSDC ccipocr3.Bytes
}

type staticTokenDataEncoder struct {
	staticTokenDataEncoderConfig
}

func newStaticTokenDataEncoder(lggr logger.Logger, cfg staticTokenDataEncoderConfig) staticTokenDataEncoder {
	lggr = logger.Named(lggr, "staticTokenDataEncoder")
	return staticTokenDataEncoder{
		staticTokenDataEncoderConfig: cfg,
	}
}

func (s staticTokenDataEncoder) EncodeUSDC(ctx context.Context, message ccipocr3.Bytes, attestation ccipocr3.Bytes) (ccipocr3.Bytes, error) {
	return s.encodedUSDC, nil
}

func (s staticTokenDataEncoder) Evaluate(ctx context.Context, other ccipocr3.TokenDataEncoder) error {
	// Test EncodeUSDC
	otherEncoded, err := other.EncodeUSDC(ctx, ccipocr3.Bytes("test-message"), ccipocr3.Bytes("test-attestation"))
	if err != nil {
		return fmt.Errorf("TokenDataEncoder EncodeUSDC failed: %w", err)
	}
	myEncoded, err := s.EncodeUSDC(ctx, ccipocr3.Bytes("test-message"), ccipocr3.Bytes("test-attestation"))
	if err != nil {
		return fmt.Errorf("TokenDataEncoder EncodeUSDC failed: %w", err)
	}
	if string(otherEncoded) != string(myEncoded) {
		return fmt.Errorf("TokenDataEncoder EncodeUSDC mismatch: got %s, expected %s", string(otherEncoded), string(myEncoded))
	}

	return nil
}

func (s staticTokenDataEncoder) AssertEqual(ctx context.Context, t *testing.T, other ccipocr3.TokenDataEncoder) {
	t.Run("TokenDataEncoder", func(t *testing.T) {
		assert.NoError(t, s.Evaluate(ctx, other))
	})
}

// SourceChainExtraDataCodec implementation

type SourceChainExtraDataCodecEvaluator interface {
	ccipocr3.SourceChainExtraDataCodec
	testtypes.Evaluator[ccipocr3.SourceChainExtraDataCodec]
}

type SourceChainExtraDataCodecTester interface {
	ccipocr3.SourceChainExtraDataCodec
	testtypes.Evaluator[ccipocr3.SourceChainExtraDataCodec]
	testtypes.AssertEqualer[ccipocr3.SourceChainExtraDataCodec]
}

func SourceChainExtraDataCodec(lggr logger.Logger) staticSourceChainExtraDataCodec {
	return newStaticSourceChainExtraDataCodec(lggr, staticSourceChainExtraDataCodecConfig{
		extraArgsMap:    map[string]any{"gasLimit": uint64(100000)},
		destExecDataMap: map[string]any{"data": "test-data"},
	})
}

var _ SourceChainExtraDataCodecTester = staticSourceChainExtraDataCodec{}

type staticSourceChainExtraDataCodecConfig struct {
	extraArgsMap    map[string]any
	destExecDataMap map[string]any
}

type staticSourceChainExtraDataCodec struct {
	staticSourceChainExtraDataCodecConfig
}

func newStaticSourceChainExtraDataCodec(lggr logger.Logger, cfg staticSourceChainExtraDataCodecConfig) staticSourceChainExtraDataCodec {
	lggr = logger.Named(lggr, "staticSourceChainExtraDataCodec")
	return staticSourceChainExtraDataCodec{
		staticSourceChainExtraDataCodecConfig: cfg,
	}
}

func (s staticSourceChainExtraDataCodec) DecodeExtraArgsToMap(extraArgs ccipocr3.Bytes) (map[string]any, error) {
	return s.extraArgsMap, nil
}

func (s staticSourceChainExtraDataCodec) DecodeDestExecDataToMap(destExecData ccipocr3.Bytes) (map[string]any, error) {
	return s.destExecDataMap, nil
}

func (s staticSourceChainExtraDataCodec) Evaluate(ctx context.Context, other ccipocr3.SourceChainExtraDataCodec) error {
	// Test DecodeExtraArgsToMap
	otherExtraArgs, err := other.DecodeExtraArgsToMap(ccipocr3.Bytes("test-extra-args"))
	if err != nil {
		return fmt.Errorf("SourceChainExtraDataCodec DecodeExtraArgsToMap failed: %w", err)
	}
	myExtraArgs, err := s.DecodeExtraArgsToMap(ccipocr3.Bytes("test-extra-args"))
	if err != nil {
		return fmt.Errorf("SourceChainExtraDataCodec DecodeExtraArgsToMap failed: %w", err)
	}
	if len(otherExtraArgs) != len(myExtraArgs) {
		return fmt.Errorf("SourceChainExtraDataCodec DecodeExtraArgsToMap length mismatch: got %d, expected %d",
			len(otherExtraArgs), len(myExtraArgs))
	}

	// Test DecodeDestExecDataToMap
	otherDestExecData, err := other.DecodeDestExecDataToMap(ccipocr3.Bytes("test-dest-exec-data"))
	if err != nil {
		return fmt.Errorf("SourceChainExtraDataCodec DecodeDestExecDataToMap failed: %w", err)
	}
	myDestExecData, err := s.DecodeDestExecDataToMap(ccipocr3.Bytes("test-dest-exec-data"))
	if err != nil {
		return fmt.Errorf("SourceChainExtraDataCodec DecodeDestExecDataToMap failed: %w", err)
	}
	if len(otherDestExecData) != len(myDestExecData) {
		return fmt.Errorf("SourceChainExtraDataCodec DecodeDestExecDataToMap length mismatch: got %d, expected %d",
			len(otherDestExecData), len(myDestExecData))
	}

	return nil
}

func (s staticSourceChainExtraDataCodec) AssertEqual(ctx context.Context, t *testing.T, other ccipocr3.SourceChainExtraDataCodec) {
	t.Run("SourceChainExtraDataCodec", func(t *testing.T) {
		assert.NoError(t, s.Evaluate(ctx, other))
	})
}

// MessageHasher implementation

type MessageHasherEvaluator interface {
	ccipocr3.MessageHasher
	testtypes.Evaluator[ccipocr3.MessageHasher]
}

type MessageHasherTester interface {
	ccipocr3.MessageHasher
	testtypes.Evaluator[ccipocr3.MessageHasher]
	testtypes.AssertEqualer[ccipocr3.MessageHasher]
}

func MessageHasher(lggr logger.Logger) staticMessageHasher {
	return newStaticMessageHasher(lggr, staticMessageHasherConfig{
		hash: ccipocr3.Bytes32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
	})
}

var _ MessageHasherTester = staticMessageHasher{}

type staticMessageHasherConfig struct {
	hash ccipocr3.Bytes32
}

type staticMessageHasher struct {
	staticMessageHasherConfig
}

func newStaticMessageHasher(lggr logger.Logger, cfg staticMessageHasherConfig) staticMessageHasher {
	lggr = logger.Named(lggr, "staticMessageHasher")
	return staticMessageHasher{
		staticMessageHasherConfig: cfg,
	}
}

func (s staticMessageHasher) Hash(ctx context.Context, message ccipocr3.Message) (ccipocr3.Bytes32, error) {
	return s.hash, nil
}

func (s staticMessageHasher) Evaluate(ctx context.Context, other ccipocr3.MessageHasher) error {
	testMessage := ccipocr3.Message{
		Header: ccipocr3.RampMessageHeader{
			MessageID:           ccipocr3.Bytes32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32},
			SourceChainSelector: ccipocr3.ChainSelector(1),
			DestChainSelector:   ccipocr3.ChainSelector(2),
			SequenceNumber:      ccipocr3.SeqNum(100),
			Nonce:               42,
			TxHash:              "0x1234567890abcdef",
			OnRamp:              ccipocr3.UnknownAddress("0xabcdef1234567890"),
		},
		Sender:         ccipocr3.UnknownAddress("0xsender"),
		Data:           ccipocr3.Bytes("test-data"),
		Receiver:       ccipocr3.UnknownAddress("0xreceiver"),
		ExtraArgs:      ccipocr3.Bytes("extra-args"),
		FeeToken:       ccipocr3.UnknownAddress("0xfeetoken"),
		FeeTokenAmount: ccipocr3.NewBigInt(big.NewInt(1000)),
		FeeValueJuels:  ccipocr3.NewBigInt(big.NewInt(2000)),
		TokenAmounts: []ccipocr3.RampTokenAmount{
			{
				SourcePoolAddress: ccipocr3.UnknownAddress("0x1111111111111111111111111111111111111111"),
				DestTokenAddress:  ccipocr3.UnknownAddress("0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
				ExtraData:         ccipocr3.Bytes("extra-token-data-1"),
				Amount:            ccipocr3.NewBigInt(big.NewInt(1)),
				DestExecData:      ccipocr3.Bytes("dest-exec-data-1"),
			},
		},
	}

	otherHash, err := other.Hash(ctx, testMessage)
	if err != nil {
		return fmt.Errorf("MessageHasher other Hash failed: %w", err)
	}
	myHash, err := s.Hash(ctx, testMessage)
	if err != nil {
		return fmt.Errorf("MessageHasher Hash failed: %w", err)
	}
	if otherHash != myHash {
		return fmt.Errorf("MessageHasher Hash mismatch: got %x, expected %x", otherHash, myHash)
	}

	return nil
}

func (s staticMessageHasher) AssertEqual(ctx context.Context, t *testing.T, other ccipocr3.MessageHasher) {
	t.Run("MessageHasher", func(t *testing.T) {
		assert.NoError(t, s.Evaluate(ctx, other))
	})
}
