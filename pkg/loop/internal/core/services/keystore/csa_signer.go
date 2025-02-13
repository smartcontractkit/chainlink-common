package keystore

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

var _ core.Keystore = (*Client)(nil)

type Client struct {
	grpc pb.CSASignerClient
}

func (k Client) Accounts(ctx context.Context) (accounts []string, err error) {
	accountsResponse, err := k.grpc.Accounts(ctx, &pb.AccountsCSARequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}

	accounts = accountsResponse.GetAccounts()

	return accounts, nil
}

func (k Client) Sign(ctx context.Context, account string, data []byte) (signed []byte, err error) {
	signResponse, err := k.grpc.Sign(ctx, &pb.SignCSARequest{
		Id:   account,
		Data: data,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to sign: %w", err)
	}

	signed = signResponse.GetSigned()

	return signed, err
}

func NewClient(cc grpc.ClientConnInterface) *Client {
	return &Client{
		pb.NewCSASignerClient(cc),
	}
}

var _ pb.CSASignerServer = (*Server)(nil)

type Server struct {
	pb.UnimplementedCSASignerServer
	impl core.Keystore
}

func NewServer(impl core.Keystore) *Server {
	return &Server{impl: impl}
}

func (s Server) Accounts(ctx context.Context, req *pb.AccountsCSARequest) (*pb.AccountsCSAResponse, error) {
	accounts, err := s.impl.Accounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}
	return &pb.AccountsCSAResponse{Accounts: accounts}, nil
}

func (s Server) Sign(ctx context.Context, req *pb.SignCSARequest) (*pb.SignCSAResponse, error) {
	signed, err := s.impl.Sign(ctx, req.GetId(), req.GetData())
	if err != nil {
		return nil, fmt.Errorf("failed to sign: %w", err)
	}

	return &pb.SignCSAResponse{Signed: signed}, nil
}
