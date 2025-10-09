package orgresolver

import (
	"context"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/services/orgresolver"
)

var _ pb.OrgResolverServer = (*Server)(nil)

type Server struct {
	pb.UnimplementedOrgResolverServer
	impl orgresolver.OrgResolver
}

func NewServer(impl orgresolver.OrgResolver) *Server {
	return &Server{impl: impl}
}

func (s *Server) Get(ctx context.Context, req *pb.GetOrganizationRequest) (*pb.GetOrganizationResponse, error) {
	orgID, err := s.impl.Get(ctx, req.Owner)
	if err != nil {
		return nil, err
	}
	return &pb.GetOrganizationResponse{OrganizationId: orgID}, nil
}
