package orgresolver

import (
	"context"

	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/services/orgresolver"
)

var _ orgresolver.OrgResolver = (*Client)(nil)

type Client struct {
	grpc pb.OrgResolverClient
}

func (c *Client) Get(ctx context.Context, owner string) (string, error) {
	resp, err := c.grpc.Get(ctx, &pb.GetOrganizationRequest{Owner: owner})
	if err != nil {
		return "", err
	}
	return resp.OrganizationId, nil
}

func (c *Client) Start(_ context.Context) error {
	return nil
}

func (c *Client) HealthReport() map[string]error {
	return map[string]error{c.Name(): nil}
}

func (c *Client) Close() error {
	return nil
}

func (c *Client) Name() string {
	return "OrgResolverClient"
}

func (c *Client) Ready() error {
	return nil
}

func NewClient(cc grpc.ClientConnInterface) *Client {
	return &Client{grpc: pb.NewOrgResolverClient(cc)}
}
