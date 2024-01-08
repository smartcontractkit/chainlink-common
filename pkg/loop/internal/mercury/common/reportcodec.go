package mercury_common

import (
	"context"

	mercury_v3_internal "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/mercury/v3"
	mercury_pb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/mercury"
	mercury_v3_pb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/mercury/v3"
)

// The point of this is to translate between the well-versioned services and the
// mercury provider api

var _ mercury_pb.ReportCodecV3Server = (*reportCodecV3Server)(nil)

// reportCodecV3Server implements mercury_pb.ReportCodecV3Server by wrapping [mercury_v3_internal.ReportCodecServer]
type reportCodecV3Server struct {
	mercury_pb.UnimplementedReportCodecV3Server

	impl *mercury_v3_internal.ReportCodecServer
}

// mustEmbedUnimplementedReportCodecV3Server implements mercury_pb.ReportCodecV3Server.
func (*reportCodecV3Server) mustEmbedUnimplementedReportCodecV3Server() {
	panic("unimplemented")
}

func NewReportCodecV3Server(impl *mercury_v3_internal.ReportCodecServer) mercury_pb.ReportCodecV3Server {
	return &reportCodecV3Server{impl: impl}
}

func (r *reportCodecV3Server) BuildReport(ctx context.Context, request *mercury_v3_pb.BuildReportRequest) (*mercury_v3_pb.BuildReportReply, error) {
	return r.impl.BuildReport(ctx, request)
}

func (r *reportCodecV3Server) MaxReportLength(ctx context.Context, request *mercury_v3_pb.MaxReportLengthRequest) (*mercury_v3_pb.MaxReportLengthReply, error) {
	return r.impl.MaxReportLength(ctx, request)
}

func (r *reportCodecV3Server) ObservationTimestampFromReport(ctx context.Context, request *mercury_v3_pb.ObservationTimestampFromReportRequest) (*mercury_v3_pb.ObservationTimestampFromReportReply, error) {
	return r.impl.ObservationTimestampFromReport(ctx, request)
}
