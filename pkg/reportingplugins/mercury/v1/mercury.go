package mercury_v1

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	pkgerrors "github.com/pkg/errors"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-relay/pkg/reportingplugins/mercury"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/types"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
)

type Observation struct {
	BenchmarkPrice mercury.ObsResult[*big.Int]
	Bid            mercury.ObsResult[*big.Int]
	Ask            mercury.ObsResult[*big.Int]
}

// DataSource implementations must be thread-safe. Observe may be called by many
// different threads concurrently.
type DataSource interface {
	// Observe queries the data source. Returns a value or an error. Once the
	// context is expires, Observe may still do cheap computations and return a
	// result, but should return as quickly as possible.
	//
	// More details: In the current implementation, the context passed to
	// Observe will time out after MaxDurationObservation. However, Observe
	// should *not* make any assumptions about context timeout behavior. Once
	// the context times out, Observe should prioritize returning as quickly as
	// possible, but may still perform fast computations to return a result
	// rather than error. For example, if Observe medianizes a number of data
	// sources, some of which already returned a result to Observe prior to the
	// context's expiry, Observe might still compute their median, and return it
	// instead of an error.
	//
	// Important: Observe should not perform any potentially time-consuming
	// actions like database access, once the context passed has expired.
	Observe(ctx context.Context, repts ocrtypes.ReportTimestamp) (Observation, error)
}

var _ ocr3types.MercuryPluginFactory = Factory{}

const maxObservationLength = 32 + // feedID
	4 + // timestamp
	mercury.ByteWidthInt192 + // benchmarkPrice
	mercury.ByteWidthInt192 + // bid
	mercury.ByteWidthInt192 + // ask
	4 + // validFromTimestamp
	16 /* overapprox. of protobuf overhead */

type Factory struct {
	dataSource         DataSource
	logger             logger.Logger
	onchainConfigCodec mercury.OnchainConfigCodec
	reportCodec        ReportCodec
}

func NewFactory(ds DataSource, lggr logger.Logger, occ mercury.OnchainConfigCodec, rc ReportCodec) Factory {
	return Factory{ds, lggr, occ, rc}
}

func (fac Factory) NewMercuryPlugin(configuration ocr3types.MercuryPluginConfig) (ocr3types.MercuryPlugin, ocr3types.MercuryPluginInfo, error) {
	offchainConfig, err := mercury.DecodeOffchainConfig(configuration.OffchainConfig)
	if err != nil {
		return nil, ocr3types.MercuryPluginInfo{}, err
	}

	onchainConfig, err := fac.onchainConfigCodec.Decode(configuration.OnchainConfig)
	if err != nil {
		return nil, ocr3types.MercuryPluginInfo{}, err
	}

	maxReportLength, err := fac.reportCodec.MaxReportLength(configuration.N)
	if err != nil {
		return nil, ocr3types.MercuryPluginInfo{}, err
	}

	r := &reportingPlugin{
		offchainConfig,
		onchainConfig,
		fac.dataSource,
		fac.logger,
		fac.reportCodec,
		configuration.ConfigDigest,
		configuration.F,
		mercury.EpochRound{},
		new(big.Int),
		maxReportLength,
	}

	return r, ocr3types.MercuryPluginInfo{
		Name: "Mercury",
		Limits: ocr3types.MercuryPluginLimits{
			MaxObservationLength: maxObservationLength,
			MaxReportLength:      maxReportLength,
		},
	}, nil
}

var _ ocr3types.MercuryPlugin = (*reportingPlugin)(nil)

type reportingPlugin struct {
	offchainConfig mercury.OffchainConfig
	onchainConfig  mercury.OnchainConfig
	dataSource     DataSource
	logger         logger.Logger
	reportCodec    ReportCodec

	configDigest             ocrtypes.ConfigDigest
	f                        int
	latestAcceptedEpochRound mercury.EpochRound
	latestAcceptedMedian     *big.Int
	maxReportLength          int
}

func (rp *reportingPlugin) Observation(ctx context.Context, repts ocrtypes.ReportTimestamp, previousReport types.Report) (ocrtypes.Observation, error) {
	obs, err := rp.dataSource.Observe(ctx, repts)
	if err != nil {
		return nil, pkgerrors.Errorf("DataSource.Observe returned an error: %s", err)
	}

	p := MercuryObservationProto{Timestamp: uint32(time.Now().Unix())}
	var obsErrors []error

	if obs.BenchmarkPrice.Err != nil {
		obsErrors = append(obsErrors, pkgerrors.Wrap(obs.BenchmarkPrice.Err, "failed to observe BenchmarkPrice"))
	} else if benchmarkPrice, err := mercury.EncodeValueInt192(obs.BenchmarkPrice.Val); err != nil {
		obsErrors = append(obsErrors, pkgerrors.Wrap(err, "failed to observe BenchmarkPrice; encoding failed"))
	} else {
		p.BenchmarkPrice = benchmarkPrice
	}

	if obs.Bid.Err != nil {
		obsErrors = append(obsErrors, pkgerrors.Wrap(obs.Bid.Err, "failed to observe Bid"))
	} else if bid, err := mercury.EncodeValueInt192(obs.Bid.Val); err != nil {
		obsErrors = append(obsErrors, pkgerrors.Wrap(err, "failed to observe Bid; encoding failed"))
	} else {
		p.Bid = bid
	}

	if obs.Ask.Err != nil {
		obsErrors = append(obsErrors, pkgerrors.Wrap(obs.Ask.Err, "failed to observe Ask"))
	} else if bid, err := mercury.EncodeValueInt192(obs.Ask.Val); err != nil {
		obsErrors = append(obsErrors, pkgerrors.Wrap(err, "failed to observe Ask; encoding failed"))
	} else {
		p.Ask = bid
	}

	if obs.BenchmarkPrice.Err == nil && obs.Bid.Err == nil && obs.Ask.Err == nil {
		p.PricesValid = true
	}

	if len(obsErrors) > 0 {
		rp.logger.Warnw(fmt.Sprintf("Observe failed %d/4 observations", len(obsErrors)), "err", errors.Join(obsErrors...))
	}

	return proto.Marshal(&p)
}

func parseAttributedObservation(ao ocrtypes.AttributedObservation) (IParsedAttributedObservation, error) {
	var pao ParsedAttributedObservation
	var obs MercuryObservationProto
	if err := proto.Unmarshal(ao.Observation, &obs); err != nil {
		return ParsedAttributedObservation{}, pkgerrors.Errorf("attributed observation cannot be unmarshaled: %s", err)
	}

	pao.Timestamp = obs.Timestamp
	pao.Observer = ao.Observer

	if obs.PricesValid {
		var err error
		pao.BenchmarkPrice, err = mercury.DecodeValueInt192(obs.BenchmarkPrice)
		if err != nil {
			return ParsedAttributedObservation{}, pkgerrors.Errorf("benchmarkPrice cannot be converted to big.Int: %s", err)
		}
		pao.Bid, err = mercury.DecodeValueInt192(obs.Bid)
		if err != nil {
			return ParsedAttributedObservation{}, pkgerrors.Errorf("bid cannot be converted to big.Int: %s", err)
		}
		pao.Ask, err = mercury.DecodeValueInt192(obs.Ask)
		if err != nil {
			return ParsedAttributedObservation{}, pkgerrors.Errorf("ask cannot be converted to big.Int: %s", err)
		}
		pao.PricesValid = true
	}

	return pao, nil
}

func parseAttributedObservations(lggr logger.Logger, aos []ocrtypes.AttributedObservation) []IParsedAttributedObservation {
	paos := make([]IParsedAttributedObservation, 0, len(aos))
	for i, ao := range aos {
		pao, err := parseAttributedObservation(ao)
		if err != nil {
			lggr.Warnw("parseAttributedObservations: dropping invalid observation",
				"observer", ao.Observer,
				"error", err,
				"i", i,
			)
			continue
		}
		paos = append(paos, pao)
	}
	return paos
}

func (rp *reportingPlugin) Report(repts types.ReportTimestamp, previousReport types.Report, aos []types.AttributedObservation) (shouldReport bool, report types.Report, err error) {
	paos := parseAttributedObservations(rp.logger, aos)

	// By assumption, we have at most f malicious oracles, so there should be at least f+1 valid paos
	if !(rp.f+1 <= len(paos)) {
		return false, nil, pkgerrors.Errorf("only received %v valid attributed observations, but need at least f+1 (%v)", len(paos), rp.f+1)
	}

	var validFromTimestamp uint32
	if previousReport != nil {
		validFromTimestamp, err = rp.reportCodec.ObservationTimestampFromReport(previousReport)
		if err != nil {
			return false, nil, pkgerrors.Errorf("failed to extract observation timestamp from previous report: %s", err)
		}
	} else {
		// todo: get rid of this calculation here
		//validFromTimestamp = mercury.GetConsensusTimestamp(paos)
		validFromTimestamp = 0
	}

	should, err := rp.shouldReport(validFromTimestamp, repts, paos)
	if err != nil || !should {
		return false, nil, err
	}

	report, err = rp.reportCodec.BuildReport(paos, rp.f, validFromTimestamp)
	if err != nil {
		rp.logger.Debugw("failed to BuildReport", "paos", paos, "f", rp.f, "validFromTimestamp", validFromTimestamp, "repts", repts)
		return false, nil, err
	}

	if !(len(report) <= rp.maxReportLength) {
		return false, nil, pkgerrors.Errorf("report with len %d violates MaxReportLength limit set by ReportCodec (%d)", len(report), rp.maxReportLength)
	} else if len(report) == 0 {
		return false, nil, errors.New("report may not have zero length (invariant violation)")
	}

	return true, report, nil
}

func (rp *reportingPlugin) shouldReport(validFromTimestamp uint32, repts types.ReportTimestamp, paos []IParsedAttributedObservation) (bool, error) {
	if !(rp.f+1 <= len(paos)) {
		return false, pkgerrors.Errorf("only received %v valid attributed observations, but need at least f+1 (%v)", len(paos), rp.f+1)
	}

	// todo: add validFromTimestamp check

	if err := errors.Join(
		rp.checkBenchmarkPrice(paos),
		rp.checkBid(paos),
		rp.checkAsk(paos),
	); err != nil {
		rp.logger.Debugw("shouldReport: no", "err", err)
		return false, nil
	}

	rp.logger.Debugw("shouldReport: yes",
		"timestamp", repts,
	)
	return true, nil
}

func (rp *reportingPlugin) checkBenchmarkPrice(paos []IParsedAttributedObservation) error {
	mPaos := Convert(paos)
	return mercury.ValidateBenchmarkPrice(mPaos, rp.f, rp.onchainConfig.Min, rp.onchainConfig.Max)
}

func (rp *reportingPlugin) checkBid(paos []IParsedAttributedObservation) error {
	mPaos := Convert(paos)
	return mercury.ValidateBid(mPaos, rp.f, rp.onchainConfig.Min, rp.onchainConfig.Max)
}

func (rp *reportingPlugin) checkAsk(paos []IParsedAttributedObservation) error {
	mPaos := Convert(paos)
	return mercury.ValidateAsk(mPaos, rp.f, rp.onchainConfig.Min, rp.onchainConfig.Max)
}

func (rp *reportingPlugin) ShouldAcceptFinalizedReport(ctx context.Context, repts types.ReportTimestamp, report types.Report) (bool, error) {
	reportEpochRound := mercury.EpochRound{repts.Epoch, repts.Round}
	if !rp.latestAcceptedEpochRound.Less(reportEpochRound) {
		rp.logger.Debugw("ShouldAcceptFinalizedReport() = false, report is stale",
			"latestAcceptedEpochRound", rp.latestAcceptedEpochRound,
			"reportEpochRound", reportEpochRound,
		)
		return false, nil
	}

	if !(len(report) <= rp.maxReportLength) {
		rp.logger.Warnw("report violates MaxReportLength limit set by ReportCodec",
			"reportEpochRound", reportEpochRound,
			"reportLength", len(report),
			"maxReportLength", rp.maxReportLength,
		)
		return false, nil
	}

	rp.logger.Debugw("ShouldAcceptFinalizedReport() = true",
		"reportEpochRound", reportEpochRound,
		"latestAcceptedEpochRound", rp.latestAcceptedEpochRound,
	)

	rp.latestAcceptedEpochRound = reportEpochRound

	return true, nil
}

func (rp *reportingPlugin) ShouldTransmitAcceptedReport(ctx context.Context, repts types.ReportTimestamp, report types.Report) (bool, error) {
	return true, nil
}

func (rp *reportingPlugin) Close() error {
	return nil
}
