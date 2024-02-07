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
)

type RI struct{}

type ocr3reportingPluginFactoryClient struct {
	*brokerExt
	*serviceClient
	grpc pb.OCR3ReportingPluginFactoryClient
}

func newOCR3ReportingPluginFactoryClient(b *brokerExt, cc grpc.ClientConnInterface) *ocr3reportingPluginFactoryClient {
	return &ocr3reportingPluginFactoryClient{b.withName("OCR3ReportingPluginProviderClient"), newServiceClient(b, cc), pb.NewOCR3ReportingPluginFactoryClient(cc)}
}

func (r *ocr3reportingPluginFactoryClient) NewReportingPlugin(config ocr3types.ReportingPluginConfig) (ocr3types.ReportingPlugin[any], ocr3types.ReportingPluginInfo, error) {
	ctx, cancel := r.stopCtx()
	defer cancel()

	reply, err := r.grpc.NewReportingPlugin(ctx, &pb.OCR3NewReportingPluginRequest{ReportingPluginConfig: &pb.OCR3ReportingPluginConfig{
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

var _ pb.OCR3ReportingPluginFactoryServer = (*ocr3reportingPluginFactoryServer)(nil)

type ocr3reportingPluginFactoryServer struct {
	pb.UnimplementedOCR3ReportingPluginFactoryServer

	*brokerExt

	impl ocr3types.ReportingPluginFactory[any]
}

func newOCR3ReportingPluginFactoryServer(impl ocr3types.ReportingPluginFactory[any], b *brokerExt) *ocr3reportingPluginFactoryServer {
	return &ocr3reportingPluginFactoryServer{impl: impl, brokerExt: b.withName("OCR3ReportingPluginFactoryServer")}
}

func (r *ocr3reportingPluginFactoryServer) NewReportingPlugin(ctx context.Context, request *pb.OCR3NewReportingPluginRequest) (*pb.OCR3NewReportingPluginReply, error) {
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
		MaxDurationShouldAcceptAttestedReport:   time.Duration(request.ReportingPluginConfig.MaxDurationShouldTransmitAcceptedReport),
		MaxDurationShouldTransmitAcceptedReport: time.Duration(request.ReportingPluginConfig.MaxDurationShouldTransmitAcceptedReport),
	}
	if l := len(request.ReportingPluginConfig.ConfigDigest); l != 32 {
		return nil, ErrConfigDigestLen(l)
	}
	copy(cfg.ConfigDigest[:], request.ReportingPluginConfig.ConfigDigest)

	rp, rpi, err := r.impl.NewReportingPlugin(cfg)
	if err != nil {
		return nil, err
	}

	const name = "ReportingPlugin"
	id, _, err := r.serveNew(name, func(s *grpc.Server) {
		pb.RegisterOCR3ReportingPluginServer(s, &ocr3reportingPluginServer{impl: rp})
	}, resource{rp, name})
	if err != nil {
		return nil, err
	}

	return &pb.OCR3NewReportingPluginReply{ReportingPluginID: id, ReportingPluginInfo: &pb.OCR3ReportingPluginInfo{
		Name: rpi.Name,
		ReportingPluginLimits: &pb.OCR3ReportingPluginLimits{
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
	grpc pb.OCR3ReportingPluginClient
}

func (r *ocr3reportingPluginClient) Query(ctx context.Context, outctx ocr3types.OutcomeContext) (libocr.Query, error) {
	reply, err := r.grpc.Query(ctx, &pb.OCR3QueryRequest{
		OutcomeContext: pbOutccomeContext(outctx),
	})
	if err != nil {
		return nil, err
	}
	return reply.Query, nil
}

func (r *ocr3reportingPluginClient) Observation(ctx context.Context, outctx ocr3types.OutcomeContext, query libocr.Query) (libocr.Observation, error) {
	reply, err := r.grpc.Observation(ctx, &pb.OCR3ObservationRequest{
		OutcomeContext: pbOutccomeContext(outctx),
		Query:          query,
	})
	if err != nil {
		return nil, err
	}
	return reply.Observation, nil
}

func (r *ocr3reportingPluginClient) ValidateObservation(outctx ocr3types.OutcomeContext, query libocr.Query, ao libocr.AttributedObservation) error {
	_, err := r.grpc.ValidateObservation(context.Background(), &pb.OCR3ValidateObservationRequest{
		OutcomeContext: pbOutccomeContext(outctx),
		Query:          query,
		Ao:             pbAtributesObservation(ao),
	})
	return err
}

func (r *ocr3reportingPluginClient) ObservationQuorum(outctx ocr3types.OutcomeContext, query libocr.Query) (ocr3types.Quorum, error) {
	reply, err := r.grpc.ObservationQuorum(context.Background(), &pb.OCR3ObservationQuorumRequest{
		OutcomeContext: pbOutccomeContext(outctx),
		Query:          query,
	})
	if err != nil {
		return 0, err
	}
	return ocr3types.Quorum(reply.Quorum), nil
}

func (r *ocr3reportingPluginClient) Outcome(outctx ocr3types.OutcomeContext, query libocr.Query, aos []libocr.AttributedObservation) (ocr3types.Outcome, error) {
	reply, err := r.grpc.Outcome(context.Background(), &pb.OCR3OutcomeRequest{
		OutcomeContext: pbOutccomeContext(outctx),
		Query:          query,
		Ao:             pbAtributesObservations(aos),
	})
	if err != nil {
		return nil, err
	}
	return reply.Outcome, nil
}

func (r *ocr3reportingPluginClient) Reports(seqNr uint64, outcome ocr3types.Outcome) ([]ocr3types.ReportWithInfo[any], error) {
	reply, err := r.grpc.Reports(context.Background(), &pb.OCR3ReportsRequest{
		SeqNr:   seqNr,
		Outcome: outcome,
	})
	if err != nil {
		return nil, err
	}
	return reportsWithInfo(reply.ReportWithInfo), nil
}

func (r *ocr3reportingPluginClient) ShouldAcceptAttestedReport(ctx context.Context, u uint64, ri ocr3types.ReportWithInfo[any]) (bool, error) {
	reply, err := r.grpc.ShouldAcceptAttestedReport(ctx, &pb.OCR3ShouldAcceptAttestedReportRequest{
		SegNr: u,
		Ri:    &pb.OCR3ReportWithInfo{Report: ri.Report},
	})
	if err != nil {
		return false, err
	}
	return reply.ShouldAccept, nil
}

func (r *ocr3reportingPluginClient) ShouldTransmitAcceptedReport(ctx context.Context, u uint64, ri ocr3types.ReportWithInfo[any]) (bool, error) {
	reply, err := r.grpc.ShouldTransmitAcceptedReport(ctx, &pb.OCR3ShouldTransmitAcceptedReportRequest{
		SegNr: u,
		Ri:    &pb.OCR3ReportWithInfo{Report: ri.Report},
	})
	if err != nil {
		return false, err
	}
	return reply.ShouldTransmit, nil
}

func (r *ocr3reportingPluginClient) Close() error {
	ctx, cancel := r.stopCtx()
	defer cancel()

	_, err := r.grpc.Close(ctx, &emptypb.Empty{})
	return err
}

func newOCR3ReportingPluginClient(b *brokerExt, cc grpc.ClientConnInterface) *ocr3reportingPluginClient {
	return &ocr3reportingPluginClient{b.withName("OCR3ReportingPluginClient"), pb.NewOCR3ReportingPluginClient(cc)}
}

var _ pb.OCR3ReportingPluginServer = (*ocr3reportingPluginServer)(nil)

type ocr3reportingPluginServer struct {
	pb.UnimplementedOCR3ReportingPluginServer

	impl ocr3types.ReportingPlugin[any]
}

func (r *ocr3reportingPluginServer) Query(ctx context.Context, request *pb.OCR3QueryRequest) (*pb.OCR3QueryReply, error) {
	oc := outcomeContext(request.OutcomeContext)
	q, err := r.impl.Query(ctx, oc)
	if err != nil {
		return nil, err
	}
	return &pb.OCR3QueryReply{Query: q}, nil
}

func (r *ocr3reportingPluginServer) Observation(ctx context.Context, request *pb.OCR3ObservationRequest) (*pb.OCR3ObservationReply, error) {
	o, err := r.impl.Observation(ctx, outcomeContext(request.OutcomeContext), request.Query)
	if err != nil {
		return nil, err
	}
	return &pb.OCR3ObservationReply{Observation: o}, nil
}

func (r *ocr3reportingPluginServer) ValidateObservation(ctx context.Context, request *pb.OCR3ValidateObservationRequest) (*emptypb.Empty, error) {
	ao, err := ocr3attributedObservation(request.Ao)
	if err != nil {
		return nil, err
	}
	err = r.impl.ValidateObservation(outcomeContext(request.OutcomeContext), request.Query, ao)
	return new(emptypb.Empty), err
}

func (r *ocr3reportingPluginServer) ObservationQuorum(ctx context.Context, request *pb.OCR3ObservationQuorumRequest) (*pb.OCR3ObservationQuorumReply, error) {
	oq, err := r.impl.ObservationQuorum(outcomeContext(request.OutcomeContext), request.Query)
	if err != nil {
		return nil, err
	}
	return &pb.OCR3ObservationQuorumReply{Quorum: int32(oq)}, nil
}

func (r *ocr3reportingPluginServer) Outcome(ctx context.Context, request *pb.OCR3OutcomeRequest) (*pb.OCR3OutcomeReply, error) {
	aos, err := ocr3attributedObservations(request.Ao)
	if err != nil {
		return nil, err
	}
	o, err := r.impl.Outcome(outcomeContext(request.OutcomeContext), request.Query, aos)
	if err != nil {
		return nil, err
	}
	return &pb.OCR3OutcomeReply{
		Outcome: o,
	}, nil
}

func (r *ocr3reportingPluginServer) Reports(ctx context.Context, request *pb.OCR3ReportsRequest) (*pb.OCR3ReportsReply, error) {
	ri, err := r.impl.Reports(request.SeqNr, request.Outcome)
	if err != nil {
		return nil, err
	}
	return &pb.OCR3ReportsReply{
		ReportWithInfo: pbReportsWithInfo(ri),
	}, nil
}

func (r *ocr3reportingPluginServer) ShouldAcceptAttestedReport(ctx context.Context, request *pb.OCR3ShouldAcceptAttestedReportRequest) (*pb.OCR3ShouldAcceptAttestedReportReply, error) {
	sa, err := r.impl.ShouldAcceptAttestedReport(ctx, request.SegNr, ocr3types.ReportWithInfo[any]{
		Report: request.Ri.Report,
	})
	if err != nil {
		return nil, err
	}
	return &pb.OCR3ShouldAcceptAttestedReportReply{
		ShouldAccept: sa,
	}, nil
}

func (r *ocr3reportingPluginServer) ShouldTransmitAcceptedReport(ctx context.Context, request *pb.OCR3ShouldTransmitAcceptedReportRequest) (*pb.OCR3ShouldTransmitAcceptedReportReply, error) {
	st, err := r.impl.ShouldTransmitAcceptedReport(ctx, request.SegNr, ocr3types.ReportWithInfo[any]{
		Report: request.Ri.Report,
	})
	if err != nil {
		return nil, err
	}
	return &pb.OCR3ShouldTransmitAcceptedReportReply{
		ShouldTransmit: st,
	}, nil
}

func (r *ocr3reportingPluginServer) Close(ctx context.Context, empty *emptypb.Empty) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, r.impl.Close()
}

func pbOutccomeContext(oc ocr3types.OutcomeContext) *pb.OCR3OutcomeContext {
	return &pb.OCR3OutcomeContext{
		SeqNr:           oc.SeqNr,
		PreviousOutcome: oc.PreviousOutcome,
		Epoch:           oc.Epoch,
		Round:           oc.Round,
	}
}

func pbAtributesObservation(ao libocr.AttributedObservation) *pb.OCR3AttributedObservation {
	return &pb.OCR3AttributedObservation{
		Observation: ao.Observation,
		Observer:    uint32(ao.Observer),
	}
}

func pbAtributesObservations(aos []libocr.AttributedObservation) (pbaos []*pb.OCR3AttributedObservation) {
	for _, ao := range aos {
		pbaos = append(pbaos, pbAtributesObservation(ao))
	}

	return pbaos
}

func outcomeContext(oc *pb.OCR3OutcomeContext) ocr3types.OutcomeContext {
	return ocr3types.OutcomeContext{
		SeqNr:           oc.SeqNr,
		PreviousOutcome: oc.PreviousOutcome,
		Epoch:           oc.Epoch,
		Round:           oc.Round,
	}
}

func ocr3attributedObservation(pbo *pb.OCR3AttributedObservation) (o libocr.AttributedObservation, err error) {
	o.Observation = pbo.Observation
	if pbo.Observer > math.MaxUint8 {
		err = ErrUint8Bounds{Name: "Observer", U: pbo.Observer}
		return
	}
	o.Observer = commontypes.OracleID(pbo.Observer)
	return
}

func ocr3attributedObservations(pbo []*pb.OCR3AttributedObservation) (o []libocr.AttributedObservation, err error) {
	for _, ao := range pbo {
		a, err := ocr3attributedObservation(ao)
		if err != nil {
			return nil, err
		}
		o = append(o, a)
	}
	return
}

func pbReportsWithInfo(rwi []ocr3types.ReportWithInfo[any]) (ri []*pb.OCR3ReportWithInfo) {
	for _, r := range rwi {
		ri = append(ri, &pb.OCR3ReportWithInfo{
			Report: r.Report,
		})
	}
	return
}

func reportsWithInfo(ri []*pb.OCR3ReportWithInfo) (rwi []ocr3types.ReportWithInfo[any]) {
	for _, r := range ri {
		rwi = append(rwi, ocr3types.ReportWithInfo[any]{
			Report: r.Report,
		})
	}
	return
}
