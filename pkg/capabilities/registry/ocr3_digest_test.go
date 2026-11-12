package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	capabilitiespb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/pb"
)

func goldenDigestInputs() OCR3DigestInputs {
	return OCR3DigestInputs{
		ChainID:         11155111,
		RegistryAddress: "0x5FbDB2315678afecb367f032d93F642f64180aa3",
		DonID:           1,
		CapabilityID:    "consensus@1.0.0-alpha",
	}
}

func goldenOCR3Config() *capabilitiespb.OCR3Config {
	return &capabilitiespb.OCR3Config{
		Signers:               [][]byte{{0x01, 0x02, 0x03}, {0x04, 0x05, 0x06}, {0x07, 0x08, 0x09}, {0x0a, 0x0b, 0x0c}},
		Transmitters:          [][]byte{{0x11, 0x12}, {0x13, 0x14}, {0x15, 0x16}, {0x17, 0x18}},
		F:                     1,
		OnchainConfig:         []byte("onchain"),
		OffchainConfigVersion: 30,
		OffchainConfig:        []byte("offchain"),
		ConfigCount:           7,
	}
}

// The expected values were produced by the original implementation in
// chainlink's core/services/ocr/capregconfig. They must never change: nodes on
// different versions have to agree on the digest.
func TestOCR3ConfigDigest_Golden(t *testing.T) {
	with := func(f func(*OCR3DigestInputs)) OCR3DigestInputs {
		in := goldenDigestInputs()
		f(&in)
		return in
	}

	tests := []struct {
		name string
		in   OCR3DigestInputs
		key  string
		cfg  *capabilitiespb.OCR3Config
		want string
	}{
		{"base", goldenDigestInputs(), "__default__", goldenOCR3Config(),
			"000ea04ca28fc10c13978a86621f9482af243786610f399190c5c08a45c6a5ef"},
		{"chain id", with(func(in *OCR3DigestInputs) { in.ChainID = 1 }), "__default__", goldenOCR3Config(),
			"000ee7db1958a3f5d1bd21cea04d357aad06fd572049e9a5f980616548ff5354"},
		{"don id", with(func(in *OCR3DigestInputs) { in.DonID = 2 }), "__default__", goldenOCR3Config(),
			"000ea7ff5916d3c441445b6e21ac763b9a532476745ccf8a80e190bb3b971643"},
		{"capability id", with(func(in *OCR3DigestInputs) { in.CapabilityID = "evm@1.0.0" }), "__default__", goldenOCR3Config(),
			"000ea0d42d080cb69028adf86864ed1a67567669cf3c2c0fa4c1507c344e847f"},
		{"ocr config key", goldenDigestInputs(), "secondary", goldenOCR3Config(),
			"000eb2866a27a4b97be0f290ce16e5c0c2e146126683eef561df40aec311303a"},
		{"empty config", goldenDigestInputs(), "__default__", &capabilitiespb.OCR3Config{},
			"000e700fc0b71b1ee7be224c2020223aa4234d14627bd78ca5ac28ab7e42a63f"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := OCR3ConfigDigest(tt.in, tt.key, tt.cfg)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.Hex())
			assert.True(t, ocrtypes.ConfigDigestPrefixKeystoneOCR3Capability.IsPrefixOf(got))
		})
	}
}

func TestOCR3ConfigDigest_Limits(t *testing.T) {
	cfg := goldenOCR3Config()
	cfg.F = 256
	_, err := OCR3ConfigDigest(goldenDigestInputs(), "__default__", cfg)
	require.ErrorContains(t, err, "f value too large")

	cfg = goldenOCR3Config()
	cfg.Signers = make([][]byte, 256)
	_, err = OCR3ConfigDigest(goldenDigestInputs(), "__default__", cfg)
	require.ErrorContains(t, err, "too many signers")
}

func TestDeepCopyRegistryMetadata_KeepsOCR3DigestInputs(t *testing.T) {
	in := goldenDigestInputs()
	md := &RegistryMetadata{
		IDsToDONs: map[DonID]DON{
			1: {CapabilityConfigurations: map[string]CapabilityConfiguration{
				in.CapabilityID: {Config: []byte{0x01}, OCR3DigestInputs: in},
			}},
		},
	}

	cp := DeepCopyRegistryMetadata(md)
	assert.Equal(t, in, cp.IDsToDONs[1].CapabilityConfigurations[in.CapabilityID].OCR3DigestInputs)
}
