package workflows

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
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
