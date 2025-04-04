package chaincapabilities

import (
	"context"
	"fmt"
	"math/big"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

// FindProgramAddress calls the corresponding RPC, passing seeds and programID.
func (c *Client) FindProgramAddress(seed [][]byte, programID [32]byte) ([32]byte, error) {
	var out [32]byte

	req := &pb.FindProgramAddressRequest{
		Seeds:     seed,
		ProgramId: programID[:],
	}

	resp, err := c.grpc.FindProgramAddress(context.Background(), req)
	if err != nil {
		return out, err
	}

	if len(resp.Address) != 32 {
		return out, fmt.Errorf("invalid address length returned")
	}

	copy(out[:], resp.Address)
	return out, nil
}

// GetAccountData calls the gRPC method to fetch a single Solana account.
func (c *Client) GetAccountData(ctx context.Context, programID [32]byte) (types.SolanaAccount, error) {
	req := &pb.GetAccountDataRequest{
		ProgramId: programID[:],
	}

	resp, err := c.grpc.GetAccountData(ctx, req)
	if err != nil {
		return types.SolanaAccount{}, err
	}

	return protoToSolanaAccount(resp.Account), nil
}

// GetMultipleAccountData calls the gRPC method to fetch multiple Solana accounts.
func (c *Client) GetMultipleAccountData(ctx context.Context, programIDs [][32]byte) ([]types.SolanaAccount, error) {
	var pids [][]byte
	for _, pid := range programIDs {
		p := make([]byte, 32)
		copy(p, pid[:])
		pids = append(pids, p)
	}

	req := &pb.GetMultipleAccountDataRequest{
		ProgramIds: pids,
	}

	resp, err := c.grpc.GetMultipleAccountData(ctx, req)
	if err != nil {
		return nil, err
	}

	results := make([]types.SolanaAccount, len(resp.Accounts))
	for i, a := range resp.Accounts {
		results[i] = protoToSolanaAccount(a)
	}

	return results, nil
}

// Helper: convert the proto SolanaAccount to the native Go type.
func protoToSolanaAccount(in *pb.SolanaAccount) types.SolanaAccount {
	if in == nil {
		return types.SolanaAccount{}
	}

	var owner [32]byte
	copy(owner[:], in.Owner)

	var rentEpoch big.Int
	// Parse the rent epoch from string; if parsing fails, rentEpoch defaults to 0.
	rentEpoch.SetString(in.RentEpoch, 10)

	return types.SolanaAccount{
		Lamports:   in.Lamports,
		Owner:      owner,
		Data:       in.Data,
		Executable: in.Executable,
		RentEpoch:  &rentEpoch,
	}
}

// FindProgramAddress handles the RPC call for finding a program address.
func (s *Server) FindProgramAddress(_ context.Context, req *pb.FindProgramAddressRequest) (*pb.FindProgramAddressReply, error) {
	var programID [32]byte
	copy(programID[:], req.ProgramId)

	addr, err := s.impl.FindProgramAddress(req.Seeds, programID)
	if err != nil {
		return nil, err
	}
	return &pb.FindProgramAddressReply{Address: addr[:]}, nil
}

// GetAccountData handles the RPC call for retrieving a single Solana account.
func (s *Server) GetAccountData(ctx context.Context, req *pb.GetAccountDataRequest) (*pb.GetAccountDataReply, error) {
	var programID [32]byte
	copy(programID[:], req.ProgramId)

	acc, err := s.impl.GetAccountData(ctx, programID)
	if err != nil {
		return nil, err
	}

	return &pb.GetAccountDataReply{
		Account: solanaAccountToProto(acc),
	}, nil
}

// GetMultipleAccountData handles the RPC call for retrieving multiple Solana accounts.
func (s *Server) GetMultipleAccountData(ctx context.Context, req *pb.GetMultipleAccountDataRequest) (*pb.GetMultipleAccountDataReply, error) {
	var programIDs [][32]byte
	for _, pid := range req.ProgramIds {
		var arr [32]byte
		copy(arr[:], pid)
		programIDs = append(programIDs, arr)
	}

	accounts, err := s.impl.GetMultipleAccountData(ctx, programIDs)
	if err != nil {
		return nil, err
	}

	resp := &pb.GetMultipleAccountDataReply{
		Accounts: make([]*pb.SolanaAccount, len(accounts)),
	}
	for i, acc := range accounts {
		resp.Accounts[i] = solanaAccountToProto(acc)
	}
	return resp, nil
}

// Helper: convert native SolanaAccount to the proto message.
func solanaAccountToProto(acc types.SolanaAccount) *pb.SolanaAccount {
	return &pb.SolanaAccount{
		Lamports:   acc.Lamports,
		Owner:      acc.Owner[:],
		Data:       acc.Data,
		Executable: acc.Executable,
		RentEpoch:  acc.RentEpoch.String(),
	}
}
