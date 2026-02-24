package contexts

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

const (
	headerKeyOrg                  = "X-Chainlink-Org"
	headerKeyOwner                = "X-Chainlink-Owner"
	headerKeyWorkflow             = "X-Chainlink-Workflow"
	headerKeyTransmissionSchedule = "X-Chainlink-Transmission-Schedule"
)

// CREUnaryInterceptor is a [grpc.UnaryInterceptor] that converts CRE context values to GRPC metadata.
func CREUnaryInterceptor(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	ctx = appendOutgoingMetadata(ctx)
	return invoker(ctx, method, req, reply, cc, opts...)
}

// CREStreamInterceptor is a [grpc.StreamInterceptor] that converts CRE context values to GRPC metadata.
func CREStreamInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	ctx = appendOutgoingMetadata(ctx)
	return streamer(ctx, desc, cc, method, opts...)
}

func appendOutgoingMetadata(ctx context.Context) context.Context {
	cre := CREValue(ctx)
	var kvs []string
	if cre.Org != "" {
		kvs = append(kvs, headerKeyOrg, cre.Org)
	}
	if cre.Owner != "" {
		kvs = append(kvs, headerKeyOwner, cre.Owner)
	}
	if cre.Workflow != "" {
		kvs = append(kvs, headerKeyWorkflow, cre.Workflow)
	}
	if ts := TransmissionScheduleValue(ctx); ts != "" {
		kvs = append(kvs, headerKeyTransmissionSchedule, ts)
	}
	if len(kvs) > 0 {
		ctx = metadata.AppendToOutgoingContext(ctx, kvs...)
	}
	return ctx
}

var _ grpc.UnaryServerInterceptor = (&CREServerInterceptor{}).UnaryServerInterceptor
var _ grpc.StreamServerInterceptor = (&CREServerInterceptor{}).StreamServerInterceptor

// CREServerInterceptor has methods that implement [grpc.UnaryServerInterceptor] and [grpc.StreamServerInterceptor]
type CREServerInterceptor struct {
	lggr logger.SugaredLogger
}

func NewCREServerInterceptor(lggr logger.Logger) *CREServerInterceptor {
	return &CREServerInterceptor{logger.Sugared(lggr).Named("CREServerInterceptor")}
}

// UnaryServerInterceptor is a [grpc.UnaryServerInterceptor] that converts GRPC metadata to CRE context values.
func (i *CREServerInterceptor) UnaryServerInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	ctx = i.extractIncomingMetadata(ctx)
	return handler(ctx, req)
}

// StreamServerInterceptor is a [grpc.StreamServerInterceptor] that converts GRPC metadata to CRE context values.
func (i *CREServerInterceptor) StreamServerInterceptor(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	return handler(srv, &wrappedStream{ServerStream: ss, ctx: i.extractIncomingMetadata(ss.Context())})
}

func (i *CREServerInterceptor) extractIncomingMetadata(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}
	var cre CRE
	if vs := md.Get(headerKeyOrg); len(vs) == 1 {
		cre.Org = vs[0]
	} else if len(vs) > 1 {
		i.lggr.Criticalw("GRPC header contains multiple orgs", "orgs", vs)
	}
	if vs := md.Get(headerKeyOwner); len(vs) == 1 {
		cre.Owner = vs[0]
	} else if len(vs) > 1 {
		i.lggr.Criticalw("GRPC header contains multiple owners", "owners", vs)
	}
	if vs := md.Get(headerKeyWorkflow); len(vs) == 1 {
		cre.Workflow = vs[0]
	} else if len(vs) > 1 {
		i.lggr.Criticalw("GRPC header contains multiple workflows", "workflows", vs)
	}
	if vs := md.Get(headerKeyTransmissionSchedule); len(vs) == 1 {
		ctx = WithTransmissionSchedule(ctx, vs[0])
	} else if len(vs) > 1 {
		i.lggr.Criticalw("GRPC header contains multiple transmission schedules", "schedules", vs)
	}
	return WithCRE(ctx, cre)
}

// wrappedStream overrides [grpc.ServerStream.Context] to return ctx
type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedStream) Context() context.Context { return w.ctx }
