package bindingsmock

import (
	"math/big"

	evmmock "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/chain-capabilities/evm/capabilitymock"
)

type IReserverManagerMock struct {
	address    []byte
	original   *evmmock.ClientCapability
	underlying *evmmock.ClientCapability

	TotalSupply func() *big.Int
	// Other methods would be generated as well
}
