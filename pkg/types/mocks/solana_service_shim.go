package mocks

import (
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	solana "github.com/smartcontractkit/chainlink-common/pkg/types/chains/solana"
)

// SolanaServiceShim wraps the mockery-generated SolanaService so it satisfies types.SolanaService.
// The generated mock cannot implement solana.Client.mustEmbedUnimplementedClient (unexported, defined in chains/solana).
type SolanaServiceShim struct {
	*SolanaService
	solana.ClientMustEmbed
}

var _ types.SolanaService = (*SolanaServiceShim)(nil)

// WrapSolanaService returns a SolanaService that delegates RPC calls to m.
func WrapSolanaService(m *SolanaService) types.SolanaService {
	if m == nil {
		return nil
	}
	return &SolanaServiceShim{
		SolanaService:       m,
		ClientMustEmbed: solana.ClientMustEmbed{},
	}
}
