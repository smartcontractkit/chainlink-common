package mercury_common

import (
	"context"

	ocr2plus_types "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	mercury_v1_internal "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/mercury/v1"
	mercury_v2_internal "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/mercury/v2"
	mercury_v3_internal "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/mercury/v3"
	mercury_pb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/mercury"
	mercury_v1_pb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/mercury/v1"
	mercury_v2_pb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/mercury/v2"
	mercury_v3_pb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/mercury/v3"
	mercury_v3_types "github.com/smartcontractkit/chainlink-common/pkg/types/mercury/v3"
)

// The point of this is to translate between the well-versioned gRPC api in [pkg/loop/internal/pb/mercury] and the
// mercury provider [pkg/types/provider_mercury.go] which is not versioned.

var _ mercury_pb.ReportCodecV3Server = (*reportCodecV3Server)(nil)

// reportCodecV3Server implements mercury_pb.ReportCodecV3Server by wrapping [mercury_v3_internal.ReportCodecServer]
type reportCodecV3Server struct {
	mercury_pb.UnimplementedReportCodecV3Server

	impl *mercury_v3_internal.ReportCodecServer
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

var _ mercury_v3_types.ReportCodec = (*reportCodecV3Client)(nil)

type reportCodecV3Client struct {
	//mercury_pb.UnimplementedReportCodecV3Client

	impl *mercury_v3_internal.ReportCodecClient
}

func NewReportCodecV3Client(impl *mercury_v3_internal.ReportCodecClient) mercury_v3_types.ReportCodec {
	return &reportCodecV3Client{impl: impl}
}

func (r *reportCodecV3Client) BuildReport(fields mercury_v3_types.ReportFields) (ocr2plus_types.Report, error) {
	return r.impl.BuildReport(fields)
}

func (r *reportCodecV3Client) MaxReportLength(n int) (int, error) {
	return r.impl.MaxReportLength(n)
}

func (r *reportCodecV3Client) ObservationTimestampFromReport(report ocr2plus_types.Report) (uint32, error) {
	return r.impl.ObservationTimestampFromReport(report)
}

var _ mercury_pb.ReportCodecV2Server = (*reportCodecV2Server)(nil)

// reportCodecV2Server implements mercury_pb.ReportCodecV2Server by wrapping [mercury_v2_internal.ReportCodecServer]
type reportCodecV2Server struct {
	mercury_pb.UnimplementedReportCodecV2Server

	impl *mercury_v2_internal.ReportCodecServer
}

func NewReportCodecV2Server(impl *mercury_v2_internal.ReportCodecServer) mercury_pb.ReportCodecV2Server {
	return &reportCodecV2Server{impl: impl}
}

func (r *reportCodecV2Server) BuildReport(ctx context.Context, request *mercury_v2_pb.BuildReportRequest) (*mercury_v2_pb.BuildReportReply, error) {
	return r.impl.BuildReport(ctx, request)
}

func (r *reportCodecV2Server) MaxReportLength(ctx context.Context, request *mercury_v2_pb.MaxReportLengthRequest) (*mercury_v2_pb.MaxReportLengthReply, error) {
	return r.impl.MaxReportLength(ctx, request)
}

func (r *reportCodecV2Server) ObservationTimestampFromReport(ctx context.Context, request *mercury_v2_pb.ObservationTimestampFromReportRequest) (*mercury_v2_pb.ObservationTimestampFromReportReply, error) {
	return r.impl.ObservationTimestampFromReport(ctx, request)
}

var _ mercury_pb.ReportCodecV1Server = (*reportCodecV1Server)(nil)

// reportCodecV1Server implements mercury_pb.ReportCodecV1Server by wrapping [mercury_v1_internal.ReportCodecServer]
type reportCodecV1Server struct {
	mercury_pb.UnimplementedReportCodecV1Server

	impl *mercury_v1_internal.ReportCodecServer
}

func NewReportCodecV1Server(impl *mercury_v1_internal.ReportCodecServer) mercury_pb.ReportCodecV1Server {
	return &reportCodecV1Server{impl: impl}
}

func (r *reportCodecV1Server) BuildReport(ctx context.Context, request *mercury_v1_pb.BuildReportRequest) (*mercury_v1_pb.BuildReportReply, error) {
	return r.impl.BuildReport(ctx, request)
}

func (r *reportCodecV1Server) MaxReportLength(ctx context.Context, request *mercury_v1_pb.MaxReportLengthRequest) (*mercury_v1_pb.MaxReportLengthReply, error) {
	return r.impl.MaxReportLength(ctx, request)
}

func (r *reportCodecV1Server) CurrentBlockNumFromReport(ctx context.Context, request *mercury_v1_pb.CurrentBlockNumFromReportRequest) (*mercury_v1_pb.CurrentBlockNumFromReportResponse, error) {
	return r.impl.CurrentBlockNumFromReport(ctx, request)
}
