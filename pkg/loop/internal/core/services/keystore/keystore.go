package keystore

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

var _ core.Keystore = (*Client)(nil)

type Client struct {
	grpc pb.KeystoreClient
}

func (k Client) Accounts(ctx context.Context) ([]string, error) {
	accts, err := k.grpc.Accounts(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}

	return accts.Accounts, nil
}

func (k Client) Sign(ctx context.Context, account string, data []byte) ([]byte, error) {
	resp, err := k.grpc.Sign(ctx, &pb.SignRequest{Account: account, Data: data})
	if err != nil {
		return nil, fmt.Errorf("failed to sign data for account: %s: %w", account, err)
	}

	return resp.SignedData, nil
}

func NewClient(cc grpc.ClientConnInterface) *Client {
	return &Client{pb.NewKeystoreClient(cc)}
}

var _ pb.KeystoreServer = (*Server)(nil)

type Server struct {
	pb.UnimplementedKeystoreServer
	impl core.Keystore
}

func NewServer(impl core.Keystore) *Server {
	return &Server{impl: impl}
}

func (s Server) Accounts(ctx context.Context, req *emptypb.Empty) (*pb.AccountsReply, error) {
	accts, err := s.impl.Accounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}
	return &pb.AccountsReply{Accounts: accts}, nil
}

func (s Server) Sign(ctx context.Context, req *pb.SignRequest) (*pb.SignReply, error) {
	signedData, err := s.impl.Sign(ctx, req.Account, req.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to sign data for account: %s: %w", req.Account, err)
	}

	return &pb.SignReply{SignedData: signedData}, nil
}
