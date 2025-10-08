package orgresolver

import (
	"context"

	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/services/orgresolver"
)

var _ orgresolver.OrgResolver = (*Client)(nil)

type Client struct {
	services.Service
	grpc pb.OrgResolverClient
}

func (c *Client) Get(ctx context.Context, owner string) (string, error) {
	resp, err := c.grpc.Get(ctx, &pb.GetOrganizationRequest{Owner: owner})
	if err != nil {
		return "", err
	}
	return resp.OrganizationId, nil
}

func NewClient(lggr logger.Logger, cc grpc.ClientConnInterface) *Client {
	return &Client{Service:services.Config{Name:}.NewService(lggr), grpc: pb.NewOrgResolverClient(cc)}
}
