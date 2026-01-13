package workflows

import (
	"crypto/sha256"
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

func GenerateWorkflowOwnerAddress(prefix string, ownerKey string) ([]byte, error) {
	// CREATE2 proposed in EIP-1014:
	// keccak256(0xff ++ address ++ salt ++ keccak256(init_code))[12:]
	// CREATE2-style address derivation inspired by the above:
	// ownerAddress = keccak256(0xff ++ bytes.repeat(0x0, 84) ++ keccak256(prefix ++ ownerKey))[12:]

	outerHash := sha3.NewLegacyKeccak256()

	// Write 0xff byte
	_, err := outerHash.Write([]byte{0xff})
	if err != nil {
		return nil, err
	}

	// Write 84 zero bytes because preimage for the final hashing round is always exactly 85 bytes
	zeroBytes := make([]byte, 84)
	_, err = outerHash.Write(zeroBytes)
	if err != nil {
		return nil, err
	}

	// Creation of the nested hash
	nestedHash := sha3.NewLegacyKeccak256()

	// Write prefix
	_, err = nestedHash.Write([]byte(prefix))
	if err != nil {
		return nil, err
	}

	// Write ownerKey
	_, err = nestedHash.Write([]byte(ownerKey))
	if err != nil {
		return nil, err
	}

	// Write the nested hash within the outer hash
	_, err = outerHash.Write(nestedHash.Sum(nil))
	if err != nil {
		return nil, err
	}

	// Return the last 20 bytes (EVM compatible address)
	return outerHash.Sum(nil)[12:], nil
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
