package datafeeds

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	chainselectors "github.com/smartcontractkit/chain-selectors"
	ocrcommon "github.com/smartcontractkit/libocr/commontypes"
	ocr2types "github.com/smartcontractkit/libocr/offchainreporting2/types"
	ocr3types "github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/types"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/types/chains/solana"
	"github.com/smartcontractkit/chainlink-protos/cre/go/values"
)

type SolanaEncoderKey = string

const (
	/*
		OutputFormat for solana:
		"account_context_hash": <"hash">,
		"payload": []reports{timestamp uint32, answer *big.Int, dataId [16]byte }
		Solana encoder compatible idl config:
		encoderConfig := map[string]any{
			report_schema": `{
			"kind": "struct",
			"fields": [
			{ "name": "payload", "type": { "vec": { "defined": "DecimalReport" } } }
			]
			}`,
			"defined_types": `[
			      {
				"name":"DecimalReport",
				 "type":{
				  "kind":"struct",
				  "fields":[
				    { "name":"timestamp", "type":"u32" },
				    { "name":"answer",    "type":"u128" },
				    { "name": "dataId",   "type": {"array": ["u8",16]}}
				  ]
				}
			      }
			]`,
				}

	*/
	TopLevelPayloadListFieldName    = SolanaEncoderKey("payload")
	TopLevelAccountCtxHashFieldName = SolanaEncoderKey("account_context_hash")
	SolTimestampOutputFieldName     = SolanaEncoderKey("timestamp")
	SolAnswerOutputFieldName        = SolanaEncoderKey("answer")
	SolDataIDOutputFieldName        = SolanaEncoderKey("dataId")
)

// secureMintReport represents the inner report structure, mimics the Report type in the SM plugin repo
type secureMintReport struct {
	ConfigDigest   ocr2types.ConfigDigest  `json:"configDigest"`
	SeqNr          uint64                  `json:"seqNr"`
	Block          uint64                  `json:"block"`
	Mintable       *big.Int                `json:"mintable"`
	AccountContext solana.AccountMetaSlice `json:"-"`
}

type wrappedMintReport struct {
	report               secureMintReport        `json:"report"`
	solanaAccountContext solana.AccountMetaSlice `json:"solanaAccountContext"`
}

// chainSelector represents the chain selector type, mimics the ChainSelector type in the SM plugin repo
type chainSelector uint64

// SecureMintAggregatorConfig is the config for the SecureMint aggregator.
// This aggregator is designed to pick out reports for a specific chain selector.
type SecureMintAggregatorConfig struct {
	// TargetChainSelector is the chain selector to look for
	TargetChainSelector chainSelector `mapstructure:"targetChainSelector"`
	DataID              [16]byte      `mapstructure:"dataID"`
}

// ToMap converts the SecureMintAggregatorConfig to a values.Map, which is suitable for the
// [NewAggregator] function in the OCR3 Aggregator interface.
func (c SecureMintAggregatorConfig) ToMap() (*values.Map, error) {
	v, err := values.WrapMap(c)
	if err != nil {
		// this should never happen since we are wrapping a struct
		return &values.Map{}, fmt.Errorf("failed to wrap SecureMintAggregatorConfig: %w", err)
	}
	return v, nil
}

var _ types.Aggregator = (*SecureMintAggregator)(nil)

type SecureMintAggregator struct {
	config     SecureMintAggregatorConfig
	formatters *formatterFactory
}

type chainReportFormatter interface {
	packReport(lggr logger.Logger, report *wrappedMintReport) (*values.Map, error)
}

type evmReportFormatter struct {
	targetChainSelector chainSelector
	dataID              [16]byte
}

func (f *evmReportFormatter) packReport(lggr logger.Logger, wreport *wrappedMintReport) (*values.Map, error) {
	report := wreport.report
	smReportAsAnswer, err := packSecureMintReportIntoUint224ForEVM(report.Mintable, report.Block)
	if err != nil {
		return nil, fmt.Errorf("failed to pack secure mint report for evm into uint224: %w", err)
	}

	lggr.Debugw("packed report into answer", "smReportAsAnswer", smReportAsAnswer)

	// This is what the DF Cache contract expects:
	// abi: "(bytes16 dataId, uint32 timestamp, uint224 answer)[] Reports"
	toWrap := []any{
		map[EVMEncoderKey]any{
			DataIDOutputFieldName:    f.dataID,
			AnswerOutputFieldName:    smReportAsAnswer,
			TimestampOutputFieldName: uint32(report.SeqNr),
		},
	}

	wrappedReport, err := values.NewMap(map[string]any{
		TopLevelListOutputFieldName: toWrap,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to wrap report: %w", err)
	}

	return wrappedReport, nil
}

func newEVMReportFormatter(chainSelector chainSelector, config SecureMintAggregatorConfig) chainReportFormatter {
	return &evmReportFormatter{targetChainSelector: chainSelector, dataID: config.DataID}
}

type solanaReportFormatter struct {
	targetChainSelector chainSelector
	dataID              [16]byte
}

func (f *solanaReportFormatter) packReport(lggr logger.Logger, wreport *wrappedMintReport) (*values.Map, error) {
	report := wreport.report
	// pack answer
	smReportAsAnswer, err := packSecureMintReportIntoU128ForSolana(report.Mintable, report.Block)
	if err != nil {
		return nil, fmt.Errorf("failed to pack secure mint report for solana into u128: %w", err)
	}
	lggr.Debugw("packed report into answer", "smReportAsAnswer", smReportAsAnswer)

	// hash account contexts
	var accounts = make([]byte, 0)
	for _, acc := range wreport.solanaAccountContext {
		accounts = append(accounts, acc.PublicKey[:]...)
	}
	lggr.Debugf("accounts length: %d", len(wreport.solanaAccountContext))
	accountContextHash := sha256.Sum256(accounts)
	lggr.Debugw("calculated account context hash", "accountContextHash", accountContextHash)

	if report.SeqNr > (1<<32 - 1) { // timestamp must fit in u32 in solana
		return nil, fmt.Errorf("timestamp exceeds u32 bounds: %v", report.SeqNr)
	}

	toWrap := []any{
		map[SolanaEncoderKey]any{
			SolTimestampOutputFieldName: uint32(report.SeqNr),
			SolAnswerOutputFieldName:    smReportAsAnswer,
			SolDataIDOutputFieldName:    f.dataID,
		},
	}
	lggr.Debugf("pass dataID %x", f.dataID)

	wrappedReport, err := values.NewMap(map[string]any{
		TopLevelAccountCtxHashFieldName: accountContextHash,
		TopLevelPayloadListFieldName:    toWrap,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to wrap report: %w", err)
	}

	return wrappedReport, nil
}

func newSolanaReportFormatter(chainSelector chainSelector, config SecureMintAggregatorConfig) chainReportFormatter {
	return &solanaReportFormatter{targetChainSelector: chainSelector, dataID: config.DataID}
}

// chainReportFormatterBuilder is a function that returns a chainReportFormatter for a given chain selector and config
type chainReportFormatterBuilder func(chainSelector chainSelector, config SecureMintAggregatorConfig) chainReportFormatter

type formatterFactory struct {
	builders map[chainSelector]chainReportFormatterBuilder
}

// register registers a new chain report formatter builder for a given chain selector
func (r *formatterFactory) register(chSel chainSelector, builder chainReportFormatterBuilder) {
	r.builders[chSel] = builder
}

// get uses a chain report formatter builder to create a chain report formatter
func (r *formatterFactory) get(chSel chainSelector, config SecureMintAggregatorConfig) (chainReportFormatter, error) {
	b, ok := r.builders[chSel]
	if !ok {
		return nil, fmt.Errorf("no formatter registered for chain selector: %d", chSel)
	}

	return b(chSel, config), nil
}

// newFormatterFactory collects all chain report formatters per chain family so that they can be used to pack reports for different chains
func newFormatterFactory() *formatterFactory {
	r := formatterFactory{
		builders: map[chainSelector]chainReportFormatterBuilder{},
	}

	// EVM
	for _, selector := range chainselectors.EvmChainIdToChainSelector() {
		r.register(chainSelector(selector), newEVMReportFormatter)
	}

	// Solana
	for _, selector := range chainselectors.SolanaChainIdToChainSelector() {
		r.register(chainSelector(selector), newSolanaReportFormatter)
	}

	return &r
}

// NewSecureMintAggregator creates a new SecureMintAggregator instance based on the provided configuration.
// The config should be a [values.Map] that represents the [SecureMintAggregatorConfig]. See [SecureMintAggregatorConfig.ToMap]
func NewSecureMintAggregator(config values.Map) (types.Aggregator, error) {
	parsedConfig, err := parseSecureMintConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config (%+v): %w", config, err)
	}
	registry := newFormatterFactory()

	return &SecureMintAggregator{
		config:     parsedConfig,
		formatters: registry,
	}, nil
}

// Aggregate implements the Aggregator interface
// This implementation:
// 1. Extracts OCRTriggerEvent from observations
// 2. Deserializes the inner ReportWithInfo to get chain selector and report
// 3. Validates chain selector matches target and sequence number is higher than previous
// 4. Returns the report in the format expected by the DF Cache, packing the mintable and block number into the decimal 'answer' field
func (a *SecureMintAggregator) Aggregate(lggr logger.Logger, previousOutcome *types.AggregationOutcome, observations map[ocrcommon.OracleID][]values.Value, f int) (*types.AggregationOutcome, error) {
	lggr = logger.Named(lggr, "SecureMintAggregator")

	lggr.Debugw("Aggregate called", "config", a.config, "observations", len(observations), "f", f, "previousOutcome", previousOutcome)

	if len(observations) == 0 {
		return nil, errors.New("no observations")
	}

	// Extract and validate reports from all observations
	validReports, err := a.extractAndValidateReports(lggr, observations, previousOutcome)
	if err != nil {
		return nil, fmt.Errorf("failed to extract and validate reports: %w", err)
	}

	if len(validReports) == 0 {
		lggr.Infow("no reports selected", "targetChainSelector", a.config.TargetChainSelector)
		return &types.AggregationOutcome{
			ShouldReport: false,
		}, nil
	}

	// Take the first valid report
	targetReport := validReports[0]

	// Create the aggregation outcome
	outcome, err := a.createOutcome(lggr, targetReport)
	if err != nil {
		return nil, fmt.Errorf("failed to create outcome: %w", err)
	}

	lggr.Debugw("SecureMint Aggregate complete", "targetChainSelector", a.config.TargetChainSelector)
	return outcome, nil
}

// extractAndValidateReports extracts OCRTriggerEvent from observations and validates them
func (a *SecureMintAggregator) extractAndValidateReports(lggr logger.Logger, observations map[ocrcommon.OracleID][]values.Value, previousOutcome *types.AggregationOutcome) ([]*wrappedMintReport, error) {
	var validReports []*wrappedMintReport
	var foundMatchingChainSelector bool

	for nodeID, nodeObservations := range observations {
		lggr = logger.With(lggr, "nodeID", nodeID)

		for _, observation := range nodeObservations {
			lggr.Debugw("processing observation", "observation", observation)

			// Extract OCRTriggerEvent from the observation
			type ObsWithCtx struct {
				Event  capabilities.OCRTriggerEvent `mapstructure:"event"`
				Solana solana.AccountMetaSlice      `mapstructure:"solana"`
			}

			obsWithContext := &ObsWithCtx{}

			if err := observation.UnwrapTo(obsWithContext); err != nil {
				lggr.Warnw("could not unwrap OCRTriggerEvent", "err", err, "observation", observation)
				continue
			}
			triggerEvent := obsWithContext.Event

			lggr.Debugw("Obs with context", "obs with ctx", obsWithContext)

			lggr.Debugw("triggerEvent", "triggerEvent", triggerEvent)

			// Deserialize the ReportWithInfo
			var reportWithInfo ocr3types.ReportWithInfo[chainSelector]
			if err := json.Unmarshal(triggerEvent.Report, &reportWithInfo); err != nil {
				lggr.Errorw("failed to unmarshal ReportWithInfo", "err", err)
				continue
			}

			// Check if chain selector matches target
			if reportWithInfo.Info != a.config.TargetChainSelector {
				lggr.Debugw("chain selector mismatch", "got", reportWithInfo.Info, "expected", a.config.TargetChainSelector)
				continue
			}

			// We found a matching chain selector
			foundMatchingChainSelector = true

			// Deserialize the inner secureMintReport
			var innerReport secureMintReport
			if err := json.Unmarshal(reportWithInfo.Report, &innerReport); err != nil {
				lggr.Errorw("failed to unmarshal secureMintReport", "err", err)
				continue
			}
			report := &wrappedMintReport{
				report:               innerReport,
				solanaAccountContext: obsWithContext.Solana,
			}

			validReports = append(validReports, report)
		}
	}

	// Return appropriate error based on what we found
	if !foundMatchingChainSelector {
		lggr.Infow("no reports found for target chain selector, ignoring", "targetChainSelector", a.config.TargetChainSelector)
		return nil, nil
	}

	return validReports, nil
}

// createOutcome creates the final aggregation outcome which can be sent to the KeystoneForwarder
func (a *SecureMintAggregator) createOutcome(lggr logger.Logger, report *wrappedMintReport) (*types.AggregationOutcome, error) {
	lggr = logger.Named(lggr, "SecureMintAggregator")
	lggr.Debugw("createOutcome called", "report", report)

	reportFormatter, err := a.formatters.get(
		a.config.TargetChainSelector,
		a.config,
	)
	if err != nil {
		return nil, fmt.Errorf("encountered issue fetching report formatter in createOutcome %w", err)
	}

	wrappedReport, err := reportFormatter.packReport(lggr, report)

	if err != nil {
		return nil, fmt.Errorf("encountered issue generating report in createOutcome %w", err)
	}

	reportsProto := values.Proto(wrappedReport)

	// Store the sequence number in metadata for next round
	metadata := []byte{byte(report.report.SeqNr)} // Simple metadata for now

	aggOutcome := &types.AggregationOutcome{
		EncodableOutcome: reportsProto.GetMapValue(),
		Metadata:         metadata,
		LastSeenAt:       report.report.SeqNr,
		ShouldReport:     true, // Always report since we found and verified the target report
	}

	lggr.Debugw("SecureMint AggregationOutcome created", "aggOutcome", aggOutcome)
	return aggOutcome, nil
}

// parseSecureMintConfig parses the user-facing, type-less, SecureMint aggregator config into the internal typed config.
func parseSecureMintConfig(config values.Map) (SecureMintAggregatorConfig, error) {
	type rawConfig struct {
		TargetChainSelector string `mapstructure:"targetChainSelector"`
		DataID              string `mapstructure:"dataID"`
	}

	var rawCfg rawConfig
	if err := config.UnwrapTo(&rawCfg); err != nil {
		return SecureMintAggregatorConfig{}, fmt.Errorf("failed to unwrap values.Map %+v: %w", config, err)
	}

	if rawCfg.TargetChainSelector == "" {
		return SecureMintAggregatorConfig{}, errors.New("targetChainSelector is required")
	}

	sel, err := strconv.ParseUint(rawCfg.TargetChainSelector, 10, 64)
	if err != nil {
		return SecureMintAggregatorConfig{}, fmt.Errorf("invalid chain selector: %w", err)
	}

	if rawCfg.DataID == "" {
		return SecureMintAggregatorConfig{}, errors.New("dataID is required")
	}

	// strip 0x prefix if present
	dataID := strings.TrimPrefix(rawCfg.DataID, "0x")

	decodedDataID, err := hex.DecodeString(dataID)
	if err != nil {
		return SecureMintAggregatorConfig{}, fmt.Errorf("invalid dataID: %v %w", dataID, err)
	}

	if len(decodedDataID) != 16 {
		return SecureMintAggregatorConfig{}, fmt.Errorf("dataID must be 16 bytes, got %d", len(decodedDataID))
	}

	parsedConfig := SecureMintAggregatorConfig{
		TargetChainSelector: chainSelector(sel),
		DataID:              [16]byte(decodedDataID),
	}

	return parsedConfig, nil
}

var maxMintableEVM = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(1)) // 2^128 - 1

// packSecureMintReportIntoUint224ForEVM packs the mintable and block number into a single uint224 so that it can be used as a price in the DF Cache contract
// (top 32 - not used / middle 64 - block number / lower 128 - mintable amount)
func packSecureMintReportIntoUint224ForEVM(mintable *big.Int, blockNumber uint64) (*big.Int, error) {
	// Handle nil mintable
	if mintable == nil {
		return nil, errors.New("mintable cannot be nil")
	}

	// Validate that mintable fits in 128 bits
	if mintable.Cmp(maxMintableEVM) > 0 {
		return nil, fmt.Errorf("mintable amount %v exceeds maximum 128-bit value %v", mintable, maxMintableEVM)
	}

	packed := big.NewInt(0)
	// Put mintable in lower 128 bits
	packed.Or(packed, mintable)

	// Put block number in middle 64 bits (bits 128-191)
	blockNumberAsBigInt := new(big.Int).SetUint64(blockNumber)
	packed.Or(packed, new(big.Int).Lsh(blockNumberAsBigInt, 128))

	return packed, nil
}

var maxMintableSolana = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 91), big.NewInt(1)) // 2^91 - 1
var maxBlockNumberSolana uint64 = 1<<36 - 1                                                  // 2^36 - 1

// TODO: will ripcord be added for top bit?
// (top 1 - not used / middle 36 - block number / lower 91 - mintable amount)
func packSecureMintReportIntoU128ForSolana(mintable *big.Int, blockNumber uint64) (*big.Int, error) {
	// Handle nil mintable
	if mintable == nil {
		return nil, errors.New("mintable cannot be nil")
	}

	// Validate that mintable fits in 91 bits
	if mintable.Cmp(maxMintableSolana) > 0 {
		return nil, fmt.Errorf("mintable amount %v exceeds maximum 91-bit value %v", mintable, maxMintableSolana)
	}

	packed := big.NewInt(0)
	// Put mintable in lower 91 bits
	packed.Or(packed, mintable)

	if blockNumber > maxBlockNumberSolana {
		return nil, fmt.Errorf("block number %d exceeds maximum 36-bit value %d", blockNumber, maxBlockNumberSolana)
	}

	// Put block number in middle 36 bits (bits 91-126)
	blockNumberAsBigInt := new(big.Int).SetUint64(blockNumber)
	packed.Or(packed, new(big.Int).Lsh(blockNumberAsBigInt, 91))

	return packed, nil
}
