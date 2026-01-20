package workflows

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"strings"

	"golang.org/x/crypto/sha3"
)

func EncodeExecutionID(workflowID, eventID string) (string, error) {
	s := sha256.New()
	_, err := s.Write([]byte(workflowID))
	if err != nil {
		return "", err
	}

	_, err = s.Write([]byte(eventID))
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(s.Sum(nil)), nil
}

func GenerateWorkflowIDFromStrings(owner string, name string, workflow []byte, config []byte, secretsURL string) (string, error) {
	ownerWithoutPrefix := owner
	if strings.HasPrefix(owner, "0x") {
		ownerWithoutPrefix = owner[2:]
	}

	ownerb, err := hex.DecodeString(ownerWithoutPrefix)
	if err != nil {
		return "", err
	}

	wid, err := GenerateWorkflowID(ownerb, name, workflow, config, secretsURL)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(wid[:]), nil
}

var (
	versionByte = byte(0)
)

func GenerateWorkflowID(owner []byte, name string, workflow []byte, config []byte, secretsURL string) ([32]byte, error) {
	s := sha256.New()
	_, err := s.Write(owner)
	if err != nil {
		return [32]byte{}, err
	}
	_, err = s.Write([]byte(name))
	if err != nil {
		return [32]byte{}, err
	}
	_, err = s.Write(workflow)
	if err != nil {
		return [32]byte{}, err
	}
	_, err = s.Write([]byte(config))
	if err != nil {
		return [32]byte{}, err
	}
	_, err = s.Write([]byte(secretsURL))
	if err != nil {
		return [32]byte{}, err
	}

	sha := [32]byte(s.Sum(nil))
	sha[0] = versionByte

	return sha, nil
}

// CREATE2-style address derivation with domain separation and collision resistance:
// ownerAddress = keccak256(0xff ++ bytes.repeat(0x0, 84) ++
// "Chainlink Runtime Environment GenerateWorkflowOwnerAddress\x00" ++
// len(prefix).to_bytes(8, byteorder='big') ++ prefix ++
// len(ownerKey).to_bytes(8, byteorder='big') ++ ownerKey)[:20]
func GenerateWorkflowOwnerAddress(prefix string, ownerKey string) ([]byte, error) {
	hash := sha3.NewLegacyKeccak256()

	// Write 0xff byte
	_, err := hash.Write([]byte{0xff})
	if err != nil {
		return nil, err
	}

	// Write 84 zero bytes because preimage for the final hashing round is always exactly 85 bytes
	zeroBytes := make([]byte, 84)
	_, err = hash.Write(zeroBytes)
	if err != nil {
		return nil, err
	}

	// Write domain separator string
	domainSeparator := "Chainlink Runtime Environment GenerateWorkflowOwnerAddress\x00"
	_, err = hash.Write([]byte(domainSeparator))
	if err != nil {
		return nil, err
	}

	// Write length-prefixed prefix as big endian
	prefixLen := uint64(len(prefix))
	err = binary.Write(hash, binary.BigEndian, prefixLen)
	if err != nil {
		return nil, err
	}

	// Write prefix
	_, err = hash.Write([]byte(prefix))
	if err != nil {
		return nil, err
	}

	// Write length-prefixed ownerKey as big endian
	ownerKeyLen := uint64(len(ownerKey))
	err = binary.Write(hash, binary.BigEndian, ownerKeyLen)
	if err != nil {
		return nil, err
	}

	// Write ownerKey
	_, err = hash.Write([]byte(ownerKey))
	if err != nil {
		return nil, err
	}

	// Return the first 20 bytes (Ethereum address)
	fullHash := hash.Sum(nil)
	return fullHash[:20], nil
}

// HashTruncateName returns the SHA-256 hash of the workflow name truncated to the first 10 bytes.
func HashTruncateName(name string) string {
	// Compute SHA-256 hash of the input string
	hash := sha256.Sum256([]byte(name))

	// Encode as hex to ensure UTF8
	var hashBytes []byte = hash[:]
	resultHex := hex.EncodeToString(hashBytes)

	// Truncate to 10 bytes
	truncated := []byte(resultHex)[:10]
	return string(truncated)
}
