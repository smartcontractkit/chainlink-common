package chaincapabilities

import (
	"context"
	"fmt"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
)

// ReadContract calls the EVM method, passing encodedParams directly.
func (c *Client) ReadContract(ctx context.Context, method string, encodedParams []byte) ([]byte, error) {
	req := &pb.ReadContractRequest{
		Method:        method,
		EncodedParams: encodedParams,
	}

	resp, err := c.grpc.ReadContract(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Result, nil
}

// ReadContract handles the EVM RPC call by passing the raw bytes directly.
func (s *Server) ReadContract(ctx context.Context, req *pb.ReadContractRequest) (*pb.ReadContractReply, error) {
	result, err := s.impl.ReadContract(ctx, req.Method, req.EncodedParams)
	if err != nil {
		return nil, fmt.Errorf("ReadContract: %w", err)
	}
	return &pb.ReadContractReply{Result: result}, nil
}
