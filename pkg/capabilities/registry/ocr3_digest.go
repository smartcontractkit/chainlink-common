package registry

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"

	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	capabilitiespb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/pb"
)

// OCR3DigestInputs identifies where an OCR3 config lives in the CapabilitiesRegistry.
// These values are part of the config digest so that identical configs used by
// different protocol instances still get distinct digests.
type OCR3DigestInputs struct {
	ChainID         uint64
	RegistryAddress string
	DonID           uint32
	CapabilityID    string
}

func (in OCR3DigestInputs) complete() bool {
	return in.ChainID != 0 && in.RegistryAddress != "" && in.DonID != 0 && in.CapabilityID != ""
}

// OCR3ConfigDigest computes the offchain digest of an OCR3 config stored in the
// CapabilitiesRegistry. Every node in a DON must compute the same digest for the
// same config, so any change to this encoding is a breaking protocol change.
func OCR3ConfigDigest(in OCR3DigestInputs, ocrConfigKey string, cfg *capabilitiespb.OCR3Config) (ocrtypes.ConfigDigest, error) {
	buf := binary.BigEndian.AppendUint64(nil, in.ChainID)
	buf = appendLengthPrefixed(buf, []byte(in.RegistryAddress))
	buf = binary.BigEndian.AppendUint32(buf, in.DonID)
	buf = appendLengthPrefixed(buf, []byte(in.CapabilityID))
	buf = appendLengthPrefixed(buf, []byte(ocrConfigKey))
	buf = binary.BigEndian.AppendUint64(buf, cfg.ConfigCount)

	if len(cfg.Signers) > math.MaxUint8 {
		return ocrtypes.ConfigDigest{}, fmt.Errorf("too many signers: %d > %d", len(cfg.Signers), math.MaxUint8)
	}
	buf = append(buf, uint8(len(cfg.Signers))) //#nosec G115 - bounds checked above
	for _, signer := range cfg.Signers {
		buf = appendLengthPrefixed(buf, signer)
	}
	for _, transmitter := range cfg.Transmitters {
		buf = appendLengthPrefixed(buf, transmitter)
	}

	if cfg.F > math.MaxUint8 {
		return ocrtypes.ConfigDigest{}, fmt.Errorf("f value too large: %d > %d", cfg.F, math.MaxUint8)
	}
	buf = append(buf, uint8(cfg.F)) //#nosec G115 - bounds checked above
	buf = appendLengthPrefixed(buf, cfg.OnchainConfig)
	buf = binary.BigEndian.AppendUint64(buf, cfg.OffchainConfigVersion)
	buf = appendLengthPrefixed(buf, cfg.OffchainConfig)

	hash := sha256.Sum256(buf)
	var digest ocrtypes.ConfigDigest
	binary.BigEndian.PutUint16(digest[:2], uint16(ocrtypes.ConfigDigestPrefixKeystoneOCR3Capability))
	copy(digest[2:], hash[2:])
	return digest, nil
}

func appendLengthPrefixed(buf []byte, data []byte) []byte {
	buf = binary.BigEndian.AppendUint32(buf, uint32(len(data))) //#nosec G115 - config fields never approach 4GiB
	return append(buf, data...)
}
