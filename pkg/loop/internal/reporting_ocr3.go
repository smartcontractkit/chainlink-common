package internal

import (
	"context"
	"math"
	"time"

	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/libocr/commontypes"
	libocr "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	ocr3 "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/ocr3"
)

type ocr3reportingPluginFactoryClient struct {
	*brokerExt
	*serviceClient
	grpc ocr3.ReportingPluginFactoryClient
}

func newOCR3ReportingPluginFactoryClient(b *brokerExt, cc grpc.ClientConnInterface) *ocr3reportingPluginFactoryClient {
	return &ocr3reportingPluginFactoryClient{b.withName("OCR3ReportingPluginProviderClient"), newServiceClient(b, cc), ocr3.NewReportingPluginFactoryClient(cc)}
}

func (r *ocr3reportingPluginFactoryClient) NewReportingPlugin(config ocr3types.ReportingPluginConfig) (ocr3types.ReportingPlugin[any], ocr3types.ReportingPluginInfo, error) {
	ctx, cancel := r.stopCtx()
	defer cancel()

	reply, err := r.grpc.NewReportingPlugin(ctx, &ocr3.NewReportingPluginRequest{ReportingPluginConfig: &ocr3.ReportingPluginConfig{
		ConfigDigest:                            config.ConfigDigest[:],
		OracleID:                                uint32(config.OracleID),
		N:                                       uint32(config.N),
		F:                                       uint32(config.F),
		OnchainConfig:                           config.OnchainConfig,
		OffchainConfig:                          config.OffchainConfig,
		EstimatedRoundInterval:                  int64(config.EstimatedRoundInterval),
		MaxDurationQuery:                        int64(config.MaxDurationQuery),
		MaxDurationObservation:                  int64(config.MaxDurationObservation),
		MaxDurationShouldTransmitAcceptedReport: int64(config.MaxDurationShouldTransmitAcceptedReport),
		MaxDurationShouldAcceptAttestedReport:   int64(config.MaxDurationShouldAcceptAttestedReport),
	}})
	if err != nil {
		return nil, ocr3types.ReportingPluginInfo{}, err
	}
	rpi := ocr3types.ReportingPluginInfo{
		Name: reply.ReportingPluginInfo.Name,
		Limits: ocr3types.ReportingPluginLimits{
			MaxQueryLength:       int(reply.ReportingPluginInfo.ReportingPluginLimits.MaxQueryLength),
			MaxObservationLength: int(reply.ReportingPluginInfo.ReportingPluginLimits.MaxObservationLength),
			MaxReportLength:      int(reply.ReportingPluginInfo.ReportingPluginLimits.MaxReportLength),
			MaxOutcomeLength:     int(reply.ReportingPluginInfo.ReportingPluginLimits.MaxOutcomeLength),
			MaxReportCount:       int(reply.ReportingPluginInfo.ReportingPluginLimits.MaxReportCount),
		},
	}
	cc, err := r.brokerExt.dial(reply.ReportingPluginID)
	if err != nil {
		return nil, ocr3types.ReportingPluginInfo{}, err
	}
	return newOCR3ReportingPluginClient(r.brokerExt, cc), rpi, nil
}

var _ ocr3.ReportingPluginFactoryServer = (*ocr3reportingPluginFactoryServer)(nil)

type ocr3reportingPluginFactoryServer struct {
	ocr3.UnimplementedReportingPluginFactoryServer

	*brokerExt

	impl ocr3types.ReportingPluginFactory[any]
}

func newOCR3ReportingPluginFactoryServer(impl ocr3types.ReportingPluginFactory[any], b *brokerExt) *ocr3reportingPluginFactoryServer {
	return &ocr3reportingPluginFactoryServer{impl: impl, brokerExt: b.withName("OCR3ReportingPluginFactoryServer")}
}

func (r *ocr3reportingPluginFactoryServer) NewReportingPlugin(ctx context.Context, request *ocr3.NewReportingPluginRequest) (*ocr3.NewReportingPluginReply, error) {
	cfg := ocr3types.ReportingPluginConfig{
		ConfigDigest:                            libocr.ConfigDigest{},
		OracleID:                                commontypes.OracleID(request.ReportingPluginConfig.OracleID),
		N:                                       int(request.ReportingPluginConfig.N),
		F:                                       int(request.ReportingPluginConfig.F),
		OnchainConfig:                           request.ReportingPluginConfig.OnchainConfig,
		OffchainConfig:                          request.ReportingPluginConfig.OffchainConfig,
		EstimatedRoundInterval:                  time.Duration(request.ReportingPluginConfig.EstimatedRoundInterval),
		MaxDurationQuery:                        time.Duration(request.ReportingPluginConfig.MaxDurationQuery),
		MaxDurationObservation:                  time.Duration(request.ReportingPluginConfig.MaxDurationObservation),
		MaxDurationShouldTransmitAcceptedReport: time.Duration(request.ReportingPluginConfig.MaxDurationShouldTransmitAcceptedReport),
		MaxDurationShouldAcceptAttestedReport:   time.Duration(request.ReportingPluginConfig.MaxDurationShouldAcceptAttestedReport),
	}
	if l := len(request.ReportingPluginConfig.ConfigDigest); l != 32 {
		return nil, pb.ErrConfigDigestLen(l)
	}
	copy(cfg.ConfigDigest[:], request.ReportingPluginConfig.ConfigDigest)

	rp, rpi, err := r.impl.NewReportingPlugin(cfg)
	if err != nil {
		return nil, err
	}

	const name = "OCR3ReportingPlugin"
	id, _, err := r.serveNew(name, func(s *grpc.Server) {
		ocr3.RegisterReportingPluginServer(s, &ocr3reportingPluginServer{impl: rp})
	}, resource{rp, name})
	if err != nil {
		return nil, err
	}

	return &ocr3.NewReportingPluginReply{ReportingPluginID: id, ReportingPluginInfo: &ocr3.ReportingPluginInfo{
		Name: rpi.Name,
		ReportingPluginLimits: &ocr3.ReportingPluginLimits{
			MaxQueryLength:       uint64(rpi.Limits.MaxQueryLength),
			MaxObservationLength: uint64(rpi.Limits.MaxObservationLength),
			MaxOutcomeLength:     uint64(rpi.Limits.MaxOutcomeLength),
			MaxReportLength:      uint64(rpi.Limits.MaxReportLength),
			MaxReportCount:       uint64(rpi.Limits.MaxReportCount),
		},
	},
	}, nil
}

var _ ocr3types.ReportingPlugin[any] = (*ocr3reportingPluginClient)(nil)

type ocr3reportingPluginClient struct {
	*brokerExt
	grpc ocr3.ReportingPluginClient
}

func (o *ocr3reportingPluginClient) Query(ctx context.Context, outctx ocr3types.OutcomeContext) (libocr.Query, error) {
	reply, err := o.grpc.Query(ctx, &ocr3.QueryRequest{
		OutcomeContext: pbOutcomeContext(outctx),
	})
	if err != nil {
		return nil, err
	}
	return reply.Query, nil
}

func (o *ocr3reportingPluginClient) Observation(ctx context.Context, outctx ocr3types.OutcomeContext, query libocr.Query) (libocr.Observation, error) {
	reply, err := o.grpc.Observation(ctx, &ocr3.ObservationRequest{
		OutcomeContext: pbOutcomeContext(outctx),
		Query:          query,
	})
	if err != nil {
		return nil, err
	}
	return reply.Observation, nil
}

func (o *ocr3reportingPluginClient) ValidateObservation(outctx ocr3types.OutcomeContext, query libocr.Query, ao libocr.AttributedObservation) error {
	_, err := o.grpc.ValidateObservation(context.Background(), &ocr3.ValidateObservationRequest{
		OutcomeContext: pbOutcomeContext(outctx),
		Query:          query,
		Ao:             pbAttributedObservation(ao),
	})
	return err
}

func (o *ocr3reportingPluginClient) ObservationQuorum(outctx ocr3types.OutcomeContext, query libocr.Query) (ocr3types.Quorum, error) {
	reply, err := o.grpc.ObservationQuorum(context.Background(), &ocr3.ObservationQuorumRequest{
		OutcomeContext: pbOutcomeContext(outctx),
		Query:          query,
	})
	if err != nil {
		return 0, err
	}
	return ocr3types.Quorum(reply.Quorum), nil
}

func (o *ocr3reportingPluginClient) Outcome(outctx ocr3types.OutcomeContext, query libocr.Query, aos []libocr.AttributedObservation) (ocr3types.Outcome, error) {
	reply, err := o.grpc.Outcome(context.Background(), &ocr3.OutcomeRequest{
		OutcomeContext: pbOutcomeContext(outctx),
		Query:          query,
		Ao:             pbOcr3AttributedObservations(aos),
	})
	if err != nil {
		return nil, err
	}
	return reply.Outcome, nil
}

func (o *ocr3reportingPluginClient) Reports(seqNr uint64, outcome ocr3types.Outcome) ([]ocr3types.ReportWithInfo[any], error) {
	reply, err := o.grpc.Reports(context.Background(), &ocr3.ReportsRequest{
		SeqNr:   seqNr,
		Outcome: outcome,
	})
	if err != nil {
		return nil, err
	}
	return reportsWithInfo(reply.ReportWithInfo), nil
}

func (o *ocr3reportingPluginClient) ShouldAcceptAttestedReport(ctx context.Context, u uint64, ri ocr3types.ReportWithInfo[any]) (bool, error) {
	reply, err := o.grpc.ShouldAcceptAttestedReport(ctx, &ocr3.ShouldAcceptAttestedReportRequest{
		SegNr: u,
		Ri:    &ocr3.ReportWithInfo{Report: ri.Report},
	})
	if err != nil {
		return false, err
	}
	return reply.ShouldAccept, nil
}

func (o *ocr3reportingPluginClient) ShouldTransmitAcceptedReport(ctx context.Context, u uint64, ri ocr3types.ReportWithInfo[any]) (bool, error) {
	reply, err := o.grpc.ShouldTransmitAcceptedReport(ctx, &ocr3.ShouldTransmitAcceptedReportRequest{
		SegNr: u,
		Ri:    &ocr3.ReportWithInfo{Report: ri.Report},
	})
	if err != nil {
		return false, err
	}
	return reply.ShouldTransmit, nil
}

func (o *ocr3reportingPluginClient) Close() error {
	ctx, cancel := o.stopCtx()
	defer cancel()

	_, err := o.grpc.Close(ctx, &emptypb.Empty{})
	return err
}

func newOCR3ReportingPluginClient(b *brokerExt, cc grpc.ClientConnInterface) *ocr3reportingPluginClient {
	return &ocr3reportingPluginClient{b.withName("OCR3ReportingPluginClient"), ocr3.NewReportingPluginClient(cc)}
}

var _ ocr3.ReportingPluginServer = (*ocr3reportingPluginServer)(nil)

type ocr3reportingPluginServer struct {
	ocr3.UnimplementedReportingPluginServer

	impl ocr3types.ReportingPlugin[any]
}

func (o *ocr3reportingPluginServer) Query(ctx context.Context, request *ocr3.QueryRequest) (*ocr3.QueryReply, error) {
	oc := outcomeContext(request.OutcomeContext)
	q, err := o.impl.Query(ctx, oc)
	if err != nil {
		return nil, err
	}
	return &ocr3.QueryReply{Query: q}, nil
}

func (o *ocr3reportingPluginServer) Observation(ctx context.Context, request *ocr3.ObservationRequest) (*ocr3.ObservationReply, error) {
	obs, err := o.impl.Observation(ctx, outcomeContext(request.OutcomeContext), request.Query)
	if err != nil {
		return nil, err
	}
	return &ocr3.ObservationReply{Observation: obs}, nil
}

func (o *ocr3reportingPluginServer) ValidateObservation(ctx context.Context, request *ocr3.ValidateObservationRequest) (*emptypb.Empty, error) {
	ao, err := ocr3AttributedObservation(request.Ao)
	if err != nil {
		return nil, err
	}
	err = o.impl.ValidateObservation(outcomeContext(request.OutcomeContext), request.Query, ao)
	return new(emptypb.Empty), err
}

func (o *ocr3reportingPluginServer) ObservationQuorum(ctx context.Context, request *ocr3.ObservationQuorumRequest) (*ocr3.ObservationQuorumReply, error) {
	oq, err := o.impl.ObservationQuorum(outcomeContext(request.OutcomeContext), request.Query)
	if err != nil {
		return nil, err
	}
	return &ocr3.ObservationQuorumReply{Quorum: int32(oq)}, nil
}

func (o *ocr3reportingPluginServer) Outcome(ctx context.Context, request *ocr3.OutcomeRequest) (*ocr3.OutcomeReply, error) {
	aos, err := ocr3AttributedObservations(request.Ao)
	if err != nil {
		return nil, err
	}
	out, err := o.impl.Outcome(outcomeContext(request.OutcomeContext), request.Query, aos)
	if err != nil {
		return nil, err
	}
	return &ocr3.OutcomeReply{
		Outcome: out,
	}, nil
}

func (o *ocr3reportingPluginServer) Reports(ctx context.Context, request *ocr3.ReportsRequest) (*ocr3.ReportsReply, error) {
	ri, err := o.impl.Reports(request.SeqNr, request.Outcome)
	if err != nil {
		return nil, err
	}
	return &ocr3.ReportsReply{
		ReportWithInfo: pbReportsWithInfo(ri),
	}, nil
}

func (o *ocr3reportingPluginServer) ShouldAcceptAttestedReport(ctx context.Context, request *ocr3.ShouldAcceptAttestedReportRequest) (*ocr3.ShouldAcceptAttestedReportReply, error) {
	sa, err := o.impl.ShouldAcceptAttestedReport(ctx, request.SegNr, ocr3types.ReportWithInfo[any]{
		Report: request.Ri.Report,
	})
	if err != nil {
		return nil, err
	}
	return &ocr3.ShouldAcceptAttestedReportReply{
		ShouldAccept: sa,
	}, nil
}

func (o *ocr3reportingPluginServer) ShouldTransmitAcceptedReport(ctx context.Context, request *ocr3.ShouldTransmitAcceptedReportRequest) (*ocr3.ShouldTransmitAcceptedReportReply, error) {
	st, err := o.impl.ShouldTransmitAcceptedReport(ctx, request.SegNr, ocr3types.ReportWithInfo[any]{
		Report: request.Ri.Report,
	})
	if err != nil {
		return nil, err
	}
	return &ocr3.ShouldTransmitAcceptedReportReply{
		ShouldTransmit: st,
	}, nil
}

func (o *ocr3reportingPluginServer) Close(ctx context.Context, empty *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, o.impl.Close()
}

func pbOutcomeContext(oc ocr3types.OutcomeContext) *ocr3.OutcomeContext {
	return &ocr3.OutcomeContext{
		SeqNr:           oc.SeqNr,
		PreviousOutcome: oc.PreviousOutcome,
		Epoch:           oc.Epoch, //nolint:all
		Round:           oc.Round, //nolint:all
	}
}

func pbAttributedObservation(ao libocr.AttributedObservation) *ocr3.AttributedObservation {
	return &ocr3.AttributedObservation{
		Observation: ao.Observation,
		Observer:    uint32(ao.Observer),
	}
}

func pbReportsWithInfo(rwi []ocr3types.ReportWithInfo[any]) (ri []*ocr3.ReportWithInfo) {
	for _, r := range rwi {
		ri = append(ri, &ocr3.ReportWithInfo{
			Report: r.Report,
		})
	}
	return
}

func pbOcr3AttributedObservations(aos []libocr.AttributedObservation) (pbaos []*ocr3.AttributedObservation) {
	for _, ao := range aos {
		pbaos = append(pbaos, pbAttributedObservation(ao))
	}

	return pbaos
}

func outcomeContext(oc *ocr3.OutcomeContext) ocr3types.OutcomeContext {
	return ocr3types.OutcomeContext{
		SeqNr:           oc.SeqNr,
		PreviousOutcome: oc.PreviousOutcome,
		Epoch:           oc.Epoch, //nolint:all
		Round:           oc.Round, //nolint:all
	}
}

func ocr3AttributedObservation(pbo *ocr3.AttributedObservation) (o libocr.AttributedObservation, err error) {
	o.Observation = pbo.Observation
	if pbo.Observer > math.MaxUint8 {
		err = pb.ErrUint8Bounds{Name: "Observer", U: pbo.Observer}
		return
	}
	o.Observer = commontypes.OracleID(pbo.Observer)
	return
}

func ocr3AttributedObservations(pbo []*ocr3.AttributedObservation) (o []libocr.AttributedObservation, err error) {
	for _, ao := range pbo {
		a, err := ocr3AttributedObservation(ao)
		if err != nil {
			return nil, err
		}
		o = append(o, a)
	}
	return
}

func reportsWithInfo(ri []*ocr3.ReportWithInfo) (rwi []ocr3types.ReportWithInfo[any]) {
	for _, r := range ri {
		rwi = append(rwi, ocr3types.ReportWithInfo[any]{
			Report: r.Report,
		})
	}
	return
}
