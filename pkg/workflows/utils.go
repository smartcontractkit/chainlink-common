package workflows

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
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

func GenerateWorkflowIDFromStrings(owner string, workflow []byte, config []byte, secretsURL string) (string, error) {
	ownerb, err := hex.DecodeString(owner)
	if err != nil {
		return "", err
	}

	if len(ownerb) != 20 {
		return "", fmt.Errorf("invalid owner length: got %d, expected 20", len(ownerb))
	}

	wid, err := GenerateWorkflowID([20]byte(ownerb), workflow, config, secretsURL)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(wid[:]), nil
}

func GenerateWorkflowID(owner [20]byte, workflow []byte, config []byte, secretsURL string) ([32]byte, error) {
	s := sha256.New()
	_, err := s.Write(owner[:])
	if err != nil {
		return [32]byte{}, err
	}
	_, err = s.Write([]byte(workflow))
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
	return [32]byte(s.Sum(nil)), nil
}
