package contexts

import (
	"context"

	"google.golang.org/grpc/stats"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

const (
	headerKeyOrg      = "X-Chainlink-Org"
	headerKeyOwner    = "X-Chainlink-Owner"
	headerKeyWorkflow = "X-Chainlink-Workflow"
)

// NewCREStatsHandler returns a stats.Handler which transfers contexts.CRE context values over GRPC calls via headers.
func NewCREStatsHandler(lggr logger.Logger) stats.Handler {
	return &creStatsHandler{lggr: logger.Named(lggr, "CREStatsHandler")}
}

type creStatsHandler struct {
	lggr logger.Logger
}

func (s *creStatsHandler) TagRPC(ctx context.Context, info *stats.RPCTagInfo) context.Context {
	// Insert empty CRE to be mutated in HandleRPC based on headers.
	return WithCRE(ctx, CRE{})
}

func (s *creStatsHandler) HandleRPC(ctx context.Context, rpcStats stats.RPCStats) {
	switch h := rpcStats.(type) {
	case *stats.OutHeader:
		// Set the headers from context values, if present.
		cre := Value[*CRE](ctx, creCtxKey)
		if cre.Org != "" {
			h.Header.Set(headerKeyOrg, cre.Org)
		}
		if cre.Owner != "" {
			h.Header.Set(headerKeyOwner, cre.Owner)
		}
		if cre.Workflow != "" {
			h.Header.Set(headerKeyWorkflow, cre.Workflow)
		}

	case *stats.InHeader:
		// Fill the empty CRE from headers, if present.
		cre := Value[*CRE](ctx, creCtxKey) // mutate instead of replace
		if vs := h.Header.Get(headerKeyOrg); len(vs) == 1 {
			cre.Org = vs[0]
		} else if len(vs) > 1 {
			s.lggr.Errorw("GRPC header contains multiple orgs", "orgs", vs)
		}
		if vs := h.Header.Get(headerKeyOwner); len(vs) == 1 {
			cre.Owner = vs[0]
		} else if len(vs) > 1 {
			s.lggr.Errorw("GRPC header contains multiple owners", "owners", vs)
		}
		if vs := h.Header.Get(headerKeyWorkflow); len(vs) == 1 {
			cre.Workflow = vs[0]
		} else if len(vs) > 1 {
			s.lggr.Errorw("GRPC header contains multiple workflows", "workflows", vs)
		}
	}
}

func (s *creStatsHandler) TagConn(ctx context.Context, info *stats.ConnTagInfo) context.Context {
	return ctx
}

func (s *creStatsHandler) HandleConn(ctx context.Context, connStats stats.ConnStats) {}
