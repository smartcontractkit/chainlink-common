package mercury_common_test

import (
	"context"
	"fmt"
	"math/big"
	"reflect"

	mercury_types "github.com/smartcontractkit/chainlink-common/pkg/types/mercury"
)

// OnchainConfigCodecImpl is a static implementation of OnchainConfigCodec for testing
var OnchainConfigCodecImpl = staticOnchainConfigCodec{
	onChainConfigCodecParameters: StaticOnChainConfigCodecFixtures,
}

type OnchainConfigCodecEvaluator interface {
	mercury_types.OnchainConfigCodec
	// Evaluate runs the other OnchainConfigCodec and checks that
	// the results are equal to this one
	Evaluate(ctx context.Context, other mercury_types.OnchainConfigCodec) error
}

type onChainConfigCodecParameters struct {
	Encoded []byte
	Decoded mercury_types.OnchainConfig
}

var StaticOnChainConfigCodecFixtures = onChainConfigCodecParameters{
	Encoded: []byte("on chain config to be encoded"),
	Decoded: mercury_types.OnchainConfig{
		Min: big.NewInt(1),
		Max: big.NewInt(100),
	},
}

type staticOnchainConfigCodec struct {
	onChainConfigCodecParameters
}

var _ OnchainConfigCodecEvaluator = staticOnchainConfigCodec{}

func (staticOnchainConfigCodec) Encode(c mercury_types.OnchainConfig) ([]byte, error) {
	if !reflect.DeepEqual(c, StaticOnChainConfigCodecFixtures.Decoded) {
		return nil, fmt.Errorf("expected OnchainConfig %v but got %v", StaticOnChainConfigCodecFixtures.Decoded, c)
	}

	return StaticOnChainConfigCodecFixtures.Encoded, nil
}

func (staticOnchainConfigCodec) Decode([]byte) (mercury_types.OnchainConfig, error) {
	return StaticOnChainConfigCodecFixtures.Decoded, nil
}

func (staticOnchainConfigCodec) Evaluate(ctx context.Context, other mercury_types.OnchainConfigCodec) error {
	encoded, err := other.Encode(StaticOnChainConfigCodecFixtures.Decoded)
	if err != nil {
		return fmt.Errorf("failed to encode: %w", err)
	}
	if !reflect.DeepEqual(encoded, StaticOnChainConfigCodecFixtures.Encoded) {
		return fmt.Errorf("expected encoded %x but got %x", StaticOnChainConfigCodecFixtures.Encoded, encoded)
	}

	decoded, err := other.Decode(StaticOnChainConfigCodecFixtures.Encoded)
	if err != nil {
		return fmt.Errorf("failed to decode: %w", err)
	}
	if !reflect.DeepEqual(decoded, StaticOnChainConfigCodecFixtures.Decoded) {
		return fmt.Errorf("expected decoded %v but got %v", StaticOnChainConfigCodecFixtures.Decoded, decoded)
	}

	return nil
}
