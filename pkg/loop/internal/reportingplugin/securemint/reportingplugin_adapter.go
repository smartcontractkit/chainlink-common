package securemint

import (
	"context"
	"encoding/binary"
	"fmt"

	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
	sm "github.com/smartcontractkit/chainlink-common/pkg/types/core/securemint"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

// ocr3ReportingPluginFactoryBytesToChainSelectorAdapter wraps a core.OCR3ReportingPluginFactory to implement ReportingPluginFactory[securemint.ChainSelector]
type ocr3ReportingPluginFactoryBytesToChainSelectorAdapter struct {
	core.OCR3ReportingPluginFactory
}

var _ ocr3types.ReportingPluginFactory[sm.ChainSelector] = (*ocr3ReportingPluginFactoryBytesToChainSelectorAdapter)(nil)

func (a *ocr3ReportingPluginFactoryBytesToChainSelectorAdapter) NewReportingPlugin(ctx context.Context, config ocr3types.ReportingPluginConfig) (ocr3types.ReportingPlugin[sm.ChainSelector], ocr3types.ReportingPluginInfo, error) {
	plugin, info, err := a.OCR3ReportingPluginFactory.NewReportingPlugin(ctx, config)
	if err != nil {
		return nil, ocr3types.ReportingPluginInfo{}, err
	}

	// Create a wrapper that converts between []byte and securemint.ChainSelector
	wrappedPlugin := &reportingPluginBytesToChainSelectorAdapter{plugin: plugin}
	return wrappedPlugin, info, nil
}

// reportingPluginBytesToChainSelectorAdapter wraps a ReportingPlugin[[]byte] to implement ReportingPlugin[securemint.ChainSelector]
type reportingPluginBytesToChainSelectorAdapter struct {
	plugin ocr3types.ReportingPlugin[[]byte]
}

var _ ocr3types.ReportingPlugin[sm.ChainSelector] = (*reportingPluginBytesToChainSelectorAdapter)(nil)

func (r *reportingPluginBytesToChainSelectorAdapter) Query(ctx context.Context, outctx ocr3types.OutcomeContext) (types.Query, error) {
	return r.plugin.Query(ctx, outctx)
}

func (r *reportingPluginBytesToChainSelectorAdapter) Observation(ctx context.Context, outctx ocr3types.OutcomeContext, query types.Query) (types.Observation, error) {
	return r.plugin.Observation(ctx, outctx, query)
}

func (r *reportingPluginBytesToChainSelectorAdapter) ValidateObservation(ctx context.Context, outctx ocr3types.OutcomeContext, query types.Query, ao types.AttributedObservation) error {
	return r.plugin.ValidateObservation(ctx, outctx, query, ao)
}

func (r *reportingPluginBytesToChainSelectorAdapter) ObservationQuorum(ctx context.Context, outctx ocr3types.OutcomeContext, query types.Query, aos []types.AttributedObservation) (bool, error) {
	return r.plugin.ObservationQuorum(ctx, outctx, query, aos)
}

func (r *reportingPluginBytesToChainSelectorAdapter) Outcome(ctx context.Context, outctx ocr3types.OutcomeContext, query types.Query, aos []types.AttributedObservation) (ocr3types.Outcome, error) {
	return r.plugin.Outcome(ctx, outctx, query, aos)
}

func (r *reportingPluginBytesToChainSelectorAdapter) Reports(ctx context.Context, seqNr uint64, outcome ocr3types.Outcome) ([]ocr3types.ReportPlus[sm.ChainSelector], error) {
	// Get reports from the underlying plugin (which returns []ocr3types.ReportPlus[[]byte])
	reports, err := r.plugin.Reports(ctx, seqNr, outcome)
	if err != nil {
		return nil, err
	}

	// Convert []ocr3types.ReportPlus[[]byte] to []ocr3types.ReportPlus[securemint.ChainSelector]
	reportsWithInfo := make([]ocr3types.ReportPlus[sm.ChainSelector], len(reports))
	for i, report := range reports {
		var chainSelector sm.ChainSelector
		if len(report.ReportWithInfo.Info) < 8 {
			return nil, fmt.Errorf("info is less than 8 bytes: %+v", report.ReportWithInfo.Info)
		}

		// info is a uint64 encoded as []byte (8 bytes, little endian)
		chainSelector = sm.ChainSelector(binary.LittleEndian.Uint64(report.ReportWithInfo.Info[:8]))

		reportsWithInfo[i] = ocr3types.ReportPlus[sm.ChainSelector]{
			ReportWithInfo: ocr3types.ReportWithInfo[sm.ChainSelector]{
				Report: report.ReportWithInfo.Report,
				Info:   chainSelector,
			},
			TransmissionScheduleOverride: report.TransmissionScheduleOverride,
		}
	}

	return reportsWithInfo, nil
}

func (r *reportingPluginBytesToChainSelectorAdapter) ShouldAcceptAttestedReport(ctx context.Context, seqNr uint64, report ocr3types.ReportWithInfo[sm.ChainSelector]) (bool, error) {
	chainSelectorBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(chainSelectorBytes, uint64(report.Info))

	reportBytes := ocr3types.ReportWithInfo[[]byte]{
		Report: report.Report,
		Info:   chainSelectorBytes,
	}
	return r.plugin.ShouldAcceptAttestedReport(ctx, seqNr, reportBytes)
}

func (r *reportingPluginBytesToChainSelectorAdapter) ShouldTransmitAcceptedReport(ctx context.Context, seqNr uint64, report ocr3types.ReportWithInfo[sm.ChainSelector]) (bool, error) {
	chainSelectorBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(chainSelectorBytes, uint64(report.Info))

	reportBytes := ocr3types.ReportWithInfo[[]byte]{
		Report: report.Report,
		Info:   chainSelectorBytes,
	}
	return r.plugin.ShouldTransmitAcceptedReport(ctx, seqNr, reportBytes)
}

func (r *reportingPluginBytesToChainSelectorAdapter) Close() error {
	return r.plugin.Close()
}

// reportingPluginFactoryChainSelectorToBytesAdapter wraps a ReportingPluginFactory[securemint.ChainSelector] to implement ocr3types.ReportingPluginFactory[[]byte]
type reportingPluginFactoryChainSelectorToBytesAdapter struct {
	ocr3types.ReportingPluginFactory[sm.ChainSelector]
}

var _ ocr3types.ReportingPluginFactory[[]byte] = (*reportingPluginFactoryChainSelectorToBytesAdapter)(nil)

func (r *reportingPluginFactoryChainSelectorToBytesAdapter) NewReportingPlugin(ctx context.Context, config ocr3types.ReportingPluginConfig) (ocr3types.ReportingPlugin[[]byte], ocr3types.ReportingPluginInfo, error) {
	plugin, info, err := r.ReportingPluginFactory.NewReportingPlugin(ctx, config)
	if err != nil {
		return nil, ocr3types.ReportingPluginInfo{}, err
	}

	wrappedPlugin := &reportingPluginChainSelectorToBytesAdapter{plugin: plugin}
	return wrappedPlugin, info, nil
}

// reportingPluginChainSelectorToBytesAdapter wraps a ReportingPlugin[securemint.ChainSelector] to implement ReportingPlugin[[]byte]
type reportingPluginChainSelectorToBytesAdapter struct {
	plugin ocr3types.ReportingPlugin[sm.ChainSelector]
}

var _ ocr3types.ReportingPlugin[[]byte] = (*reportingPluginChainSelectorToBytesAdapter)(nil)

func (r *reportingPluginChainSelectorToBytesAdapter) Query(ctx context.Context, outctx ocr3types.OutcomeContext) (types.Query, error) {
	return r.plugin.Query(ctx, outctx)
}

func (r *reportingPluginChainSelectorToBytesAdapter) Observation(ctx context.Context, outctx ocr3types.OutcomeContext, query types.Query) (types.Observation, error) {
	return r.plugin.Observation(ctx, outctx, query)
}

func (r *reportingPluginChainSelectorToBytesAdapter) ValidateObservation(ctx context.Context, outctx ocr3types.OutcomeContext, query types.Query, ao types.AttributedObservation) error {
	return r.plugin.ValidateObservation(ctx, outctx, query, ao)
}

func (r *reportingPluginChainSelectorToBytesAdapter) ObservationQuorum(ctx context.Context, outctx ocr3types.OutcomeContext, query types.Query, aos []types.AttributedObservation) (bool, error) {
	return r.plugin.ObservationQuorum(ctx, outctx, query, aos)
}

func (r *reportingPluginChainSelectorToBytesAdapter) Outcome(ctx context.Context, outctx ocr3types.OutcomeContext, query types.Query, aos []types.AttributedObservation) (ocr3types.Outcome, error) {
	return r.plugin.Outcome(ctx, outctx, query, aos)
}

func (r *reportingPluginChainSelectorToBytesAdapter) Reports(ctx context.Context, seqNr uint64, outcome ocr3types.Outcome) ([]ocr3types.ReportPlus[[]byte], error) {
	// Get reports from the underlying plugin (which returns []ocr3types.ReportPlus[securemint.ChainSelector])
	reports, err := r.plugin.Reports(ctx, seqNr, outcome)
	if err != nil {
		return nil, err
	}

	// Convert []ocr3types.ReportPlus[securemint.ChainSelector] to []ocr3types.ReportPlus[[]byte]
	reportsWithInfo := make([]ocr3types.ReportPlus[[]byte], len(reports))
	for i, report := range reports {
		info := make([]byte, 8)
		binary.LittleEndian.PutUint64(info, uint64(report.ReportWithInfo.Info))
		reportsWithInfo[i] = ocr3types.ReportPlus[[]byte]{
			ReportWithInfo: ocr3types.ReportWithInfo[[]byte]{
				Report: report.ReportWithInfo.Report,
				Info:   info,
			},
			TransmissionScheduleOverride: report.TransmissionScheduleOverride,
		}
	}
	return reportsWithInfo, nil
}

func (r *reportingPluginChainSelectorToBytesAdapter) ShouldAcceptAttestedReport(ctx context.Context, seqNr uint64, report ocr3types.ReportWithInfo[[]byte]) (bool, error) {
	chainSelector := sm.ChainSelector(binary.LittleEndian.Uint64(report.Info[:8]))

	reportBytes := ocr3types.ReportWithInfo[sm.ChainSelector]{
		Report: report.Report,
		Info:   chainSelector,
	}
	return r.plugin.ShouldAcceptAttestedReport(ctx, seqNr, reportBytes)
}

func (r *reportingPluginChainSelectorToBytesAdapter) ShouldTransmitAcceptedReport(ctx context.Context, seqNr uint64, report ocr3types.ReportWithInfo[[]byte]) (bool, error) {
	chainSelector := sm.ChainSelector(binary.LittleEndian.Uint64(report.Info[:8]))

	reportBytes := ocr3types.ReportWithInfo[sm.ChainSelector]{
		Report: report.Report,
		Info:   chainSelector,
	}
	return r.plugin.ShouldTransmitAcceptedReport(ctx, seqNr, reportBytes)
}

func (r *reportingPluginChainSelectorToBytesAdapter) Close() error {
	return r.plugin.Close()
}
