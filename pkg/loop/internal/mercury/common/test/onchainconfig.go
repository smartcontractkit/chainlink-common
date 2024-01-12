package mercury_common_test

import (
	"fmt"
	"math/big"

	mercury_types "github.com/smartcontractkit/chainlink-common/pkg/types/mercury"
)

type OnChainConfigCodecParameters struct {
	Encoded []byte
	Decoded mercury_types.OnchainConfig
}

var StaticOnChainConfigCodecFixtures = OnChainConfigCodecParameters{
	Encoded: []byte("on chain config to be encoded"),
	Decoded: mercury_types.OnchainConfig{
		Min: big.NewInt(1),
		Max: big.NewInt(100),
	},
}

type StaticOnchainConfigCodec struct{}

var _ mercury_types.OnchainConfigCodec = StaticOnchainConfigCodec{}

func (StaticOnchainConfigCodec) Encode(c mercury_types.OnchainConfig) ([]byte, error) {
	if c != StaticOnChainConfigCodecFixtures.Decoded {
		return nil, fmt.Errorf("expected onchainconfig to be %v, got %v", StaticOnChainConfigCodecFixtures.Decoded, c)
	}
	return StaticOnChainConfigCodecFixtures.Encoded, nil
}

func (StaticOnchainConfigCodec) Decode([]byte) (mercury_types.OnchainConfig, error) {
	return StaticOnChainConfigCodecFixtures.Decoded, nil
}
