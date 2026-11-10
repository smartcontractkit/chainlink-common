package ocrutil

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"strings"

	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	capabilitiespb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/pb"
)

// RegistryConfigDigest is the digest of an OCR config stored in CRE's CapabilitiesRegistry contract. It includes the
// DON ID and capability ID so that each protocol instance's digest is unique.
func RegistryConfigDigest(
	chainID uint64,
	registryAddress string,
	capabilityID string,
	donID uint32,
	ocrConfigKey string,
	cfg *capabilitiespb.OCR3Config,
) (ocrtypes.ConfigDigest, error) {
	buf := []byte{}

	// 1. Chain ID (8 bytes, big-endian)
	var chainIDBytes [8]byte
	binary.BigEndian.PutUint64(chainIDBytes[:], chainID)
	buf = append(buf, chainIDBytes[:]...)

	// 2. Registry address (length-prefixed)
	buf = appendLengthPrefixed(buf, []byte(registryAddress))

	// 3. DON ID (4 bytes, big-endian)
	var donIDBytes [4]byte
	binary.BigEndian.PutUint32(donIDBytes[:], donID)
	buf = append(buf, donIDBytes[:]...)

	// 4. Capability ID (length-prefixed)
	buf = appendLengthPrefixed(buf, []byte(capabilityID))

	// 5. OCR config key (length-prefixed)
	buf = appendLengthPrefixed(buf, []byte(ocrConfigKey))

	// 6. Config count (8 bytes, big-endian)
	var configCountBytes [8]byte
	binary.BigEndian.PutUint64(configCountBytes[:], cfg.ConfigCount)
	buf = append(buf, configCountBytes[:]...)

	// 7. Number of signers (1 byte)
	if len(cfg.Signers) > math.MaxUint8 {
		return ocrtypes.ConfigDigest{}, fmt.Errorf("too many signers: %d > %d", len(cfg.Signers), math.MaxUint8)
	}
	buf = append(buf, uint8(len(cfg.Signers))) //#nosec G115

	// 8. Each signer (length-prefixed)
	for _, signer := range cfg.Signers {
		buf = appendLengthPrefixed(buf, signer)
	}

	// 9. Each transmitter (length-prefixed)
	for _, transmitter := range cfg.Transmitters {
		buf = appendLengthPrefixed(buf, transmitter)
	}

	// 10. F (1 byte)
	if cfg.F > math.MaxUint8 {
		return ocrtypes.ConfigDigest{}, fmt.Errorf("f value too large: %d > %d", cfg.F, math.MaxUint8)
	}
	buf = append(buf, uint8(cfg.F)) //#nosec G115

	// 11. Onchain config (length-prefixed)
	buf = appendLengthPrefixed(buf, cfg.OnchainConfig)

	// 12. Offchain config version (8 bytes, big-endian)
	var offchainVersionBytes [8]byte
	binary.BigEndian.PutUint64(offchainVersionBytes[:], cfg.OffchainConfigVersion)
	buf = append(buf, offchainVersionBytes[:]...)

	// 13. Offchain config (length-prefixed)
	buf = appendLengthPrefixed(buf, cfg.OffchainConfig)

	// Hash and create digest with prefix in first 2 bytes
	hash := sha256.Sum256(buf)
	var digest ocrtypes.ConfigDigest
	binary.BigEndian.PutUint16(digest[:2], uint16(ocrtypes.ConfigDigestPrefixKeystoneOCR3Capability))
	copy(digest[2:], hash[2:])

	return digest, nil
}

func appendLengthPrefixed(buf []byte, data []byte) []byte {
	var lenBytes [4]byte
	binary.BigEndian.PutUint32(lenBytes[:], uint32(len(data))) //#nosec G115 - data length will never exceed uint32 max in practice
	buf = append(buf, lenBytes[:]...)
	buf = append(buf, data...)
	return buf
}

// WithRegistryConfigDigest is cc, an OCR config read from the capabilities registry, with its digest set. The registry
// gives transmitters as hex, as common's registry reads encode them.
func WithRegistryConfigDigest(chainID uint64, registryAddress string, capabilityID string, donID uint32, ocrConfigKey string,
	cc ocrtypes.ContractConfig) (ocrtypes.ContractConfig, error) {
	signers := make([][]byte, len(cc.Signers))
	for i, s := range cc.Signers {
		signers[i] = s
	}

	transmitters := make([][]byte, len(cc.Transmitters))
	for i, t := range cc.Transmitters {
		b, err := hex.DecodeString(strings.TrimPrefix(string(t), "0x"))
		if err != nil {
			return ocrtypes.ContractConfig{}, fmt.Errorf("transmitter %q is not hex: %w", t, err)
		}

		transmitters[i] = b
	}

	digest, err := RegistryConfigDigest(chainID, registryAddress, capabilityID, donID, ocrConfigKey, &capabilitiespb.OCR3Config{
		Signers:               signers,
		Transmitters:          transmitters,
		F:                     uint32(cc.F),
		OnchainConfig:         cc.OnchainConfig,
		OffchainConfigVersion: cc.OffchainConfigVersion,
		OffchainConfig:        cc.OffchainConfig,
		ConfigCount:           cc.ConfigCount,
	})
	if err != nil {
		return ocrtypes.ContractConfig{}, err
	}

	cc.ConfigDigest = digest
	return cc, nil
}
