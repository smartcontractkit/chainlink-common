package ocrutil

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	capabilitiespb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/pb"
)

func TestRegistryConfigDigest(t *testing.T) {
	cfg := &capabilitiespb.OCR3Config{
		Signers:               [][]byte{{1}, {2}},
		Transmitters:          [][]byte{{0xab}, {0xcd}},
		F:                     1,
		OnchainConfig:         []byte("onchain"),
		OffchainConfigVersion: 30,
		OffchainConfig:        []byte("offchain"),
		ConfigCount:           3,
	}

	digest, err := RegistryConfigDigest(1337, "0xregistry", "consensus@1.0.0-alpha", 1, capabilitiespb.OCR3ConfigDefaultKey, cfg)
	require.NoError(t, err)
	assert.Equal(t, ocrtypes.ConfigDigestPrefixKeystoneOCR3Capability, ocrtypes.ConfigDigestPrefixFromConfigDigest(digest))

	other, err := RegistryConfigDigest(1337, "0xregistry", "consensus@1.0.0-alpha", 2, capabilitiespb.OCR3ConfigDefaultKey, cfg)
	require.NoError(t, err)
	assert.NotEqual(t, digest, other, "the DON is part of the digest")

	t.Run("from the registry's contract config", func(t *testing.T) {
		cc := ocrtypes.ContractConfig{
			Signers:               []ocrtypes.OnchainPublicKey{{1}, {2}},
			Transmitters:          []ocrtypes.Account{ocrtypes.Account(hex.EncodeToString([]byte{0xab})), "0xcd"},
			F:                     1,
			OnchainConfig:         []byte("onchain"),
			OffchainConfigVersion: 30,
			OffchainConfig:        []byte("offchain"),
			ConfigCount:           3,
		}

		withDigest, err := WithRegistryConfigDigest(1337, "0xregistry", "consensus@1.0.0-alpha", 1, capabilitiespb.OCR3ConfigDefaultKey, cc)
		require.NoError(t, err)
		assert.Equal(t, digest, withDigest.ConfigDigest)
	})
}
