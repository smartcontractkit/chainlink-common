package types

import (
	"context"
	"fmt"
)

type UnimplementedSolanaChainReader struct{}

var _ SolanaChainReader = UnimplementedSolanaChainReader{}

func (u UnimplementedSolanaChainReader) FindProgramAddress(_ [][]byte, _ [32]byte) ([32]byte, error) {
	return [32]byte{}, fmt.Errorf("not implemented")
}

func (u UnimplementedSolanaChainReader) GetAccountData(_ context.Context, _ [32]byte) (SolanaAccount, error) {
	return SolanaAccount{}, fmt.Errorf("not implemented")
}

func (u UnimplementedSolanaChainReader) GetMultipleAccountData(_ context.Context, _ [][32]byte) ([]SolanaAccount, error) {
	return []SolanaAccount{}, fmt.Errorf("not implemented")
}
