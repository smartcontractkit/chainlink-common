package secrets

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"golang.org/x/crypto/nacl/box"
)

// this matches the secrets config file by the users, see the secretsConfig.yaml file
type SecretsConfig struct {
	SecretsNames map[string][]string `yaml:"secretsNames"`
}

// this is the payload that will be encrypted
type SecretPayloadToEncrypt struct {
	WorkflowOwner string            `json:"workflowOwner"`
	Secrets       map[string]string `json:"secrets"`
}

// this holds the mapping of secret name (e.g. API_KEY) to the local environment variable name which points to the raw secret
type AssignedSecrets struct {
	WorkflowSecretName string `json:"workflowSecretName"`
	LocalEnvVarName    string `json:"localEnvVarName"`
}

// this is the metadata that will be stored in the encrypted secrets file
type Metadata struct {
	WorkflowOwner            string                       `json:"workflowOwner"`
	CapabilitiesRegistry     string                       `json:"capabilitiesRegistry"`
	DonId                    string                       `json:"donId"`
	DateEncrypted            string                       `json:"dateEncrypted"`
	NodePublicEncryptionKeys map[string]string            `json:"nodePublicEncryptionKeys"`
	EnvVarsAssignedToNodes   map[string][]AssignedSecrets `json:"envVarsAssignedToNodes"`
}

// this is the result of the encryption, will be used by the DON
type EncryptedSecretsResult struct {
	EncryptedSecrets map[string]string `json:"encryptedSecrets"`
	Metadata         Metadata          `json:"metadata"`
}

func ContainsP2pId(p2pId [32]byte, p2pIds [][32]byte) bool {
	for _, id := range p2pIds {
		if id == p2pId {
			return true
		}
	}
	return false
}

func EncryptSecretsForNodes(
	workflowOwner string,
	secrets map[string][]string,
	encryptionPublicKeys map[string][32]byte, // map of p2pIds to the node's CSA (Ed25519) key.
	config SecretsConfig,
) (map[string]string, map[string][]AssignedSecrets, error) {
	encryptedSecrets := make(map[string]string)
	secretsEnvVarsByNode := make(map[string][]AssignedSecrets) // Only used for metadata
	i := 0

	// Encrypt secrets for each node
	for p2pId, encryptionPublicKey := range encryptionPublicKeys {
		secretsPayload := SecretPayloadToEncrypt{
			WorkflowOwner: workflowOwner,
			Secrets:       make(map[string]string),
		}

		for secretName, secretValues := range secrets {
			// Assign secrets to nodes in a round-robin fashion
			secretValue := secretValues[i%len(secretValues)]
			secretsPayload.Secrets[secretName] = secretValue
		}

		// Marshal the secrets payload into JSON
		secretsJSON, err := json.Marshal(secretsPayload)
		if err != nil {
			return nil, nil, err
		}

		// Encrypt secrets payload
		encrypted, err := box.SealAnonymous(nil, secretsJSON, &encryptionPublicKey, rand.Reader)
		if err != nil {
			return nil, nil, err
		}
		encryptedSecrets[p2pId] = base64.StdEncoding.EncodeToString(encrypted)

		// Generate metadata showing which nodes were assigned which environment variables
		for secretName, envVarNames := range config.SecretsNames {
			secretsEnvVarsByNode[p2pId] = append(secretsEnvVarsByNode[p2pId], AssignedSecrets{
				WorkflowSecretName: secretName,
				LocalEnvVarName:    envVarNames[i%len(envVarNames)],
			})
		}

		i++
	}

	return encryptedSecrets, secretsEnvVarsByNode, nil
}

type X25519Key interface {
	Decrypt(box []byte) ([]byte, error)
	PublicKey() [32]byte
	PublicKeyString() string
}

func DecryptSecretsForNode(
	result EncryptedSecretsResult,
	key X25519Key,
	workflowOwner string,
) (map[string]string, error) {
	var foundP2pId string
	for p2pId, pubKey := range result.Metadata.NodePublicEncryptionKeys {
		if pubKey == key.PublicKeyString() {
			foundP2pId = p2pId
			break
		}
	}

	if foundP2pId == "" {
		return nil, fmt.Errorf("cannot find public key %s in nodePublicEncryptionKeys list", key.PublicKeyString())
	}

	bundle, ok := result.EncryptedSecrets[foundP2pId]
	if !ok {
		return nil, fmt.Errorf("cannot find secrets blob for node with public key %s", key.PublicKeyString())
	}

	bundleBytes, err := base64.StdEncoding.DecodeString(bundle)
	if err != nil {
		return nil, fmt.Errorf("cannot base64 decode bundle into bytes: %w", err)
	}

	payloadBytes, err := key.Decrypt(bundleBytes)
	if err != nil {
		return nil, fmt.Errorf("cannot decrypt box: %w", err)
	}

	var payload SecretPayloadToEncrypt
	err = json.Unmarshal(payloadBytes, &payload)
	if err != nil {
		return nil, err
	}

	if payload.WorkflowOwner != workflowOwner {
		return nil, fmt.Errorf("invalid secrets bundle: got owner %s, expected %s", payload.WorkflowOwner, workflowOwner)
	}

	return payload.Secrets, nil
}

func ValidateEncryptedSecrets(secretsData []byte, encryptionPublicKeys map[string][32]byte, workflowOwner string) error {
	var encryptedSecrets EncryptedSecretsResult
	err := json.Unmarshal(secretsData, &encryptedSecrets)
	if err != nil {
		return fmt.Errorf("failed to parse encrypted secrets JSON: %w", err)
	}

	if encryptedSecrets.Metadata.WorkflowOwner != workflowOwner {
		return fmt.Errorf("the workflow owner in the encrypted secrets metadata: %s does not match the input workflow owner: %s", encryptedSecrets.Metadata.WorkflowOwner, workflowOwner)
	}

	// Verify that the encryptedSecrets values are all valid base64 strings
	for _, encryptedSecret := range encryptedSecrets.EncryptedSecrets {
		_, err := base64.StdEncoding.DecodeString(encryptedSecret)
		if err != nil {
			return fmt.Errorf("the encrypted secrets JSON payload contains encrypted secrets which are not in base64 format: %w", err)
		}
	}

	// Check that the p2pIds keys in encryptedSecrets.EncryptedSecrets match the keys in encryptionPublicKeys
	for p2pId := range encryptedSecrets.Metadata.NodePublicEncryptionKeys {
		if _, ok := encryptedSecrets.EncryptedSecrets[p2pId]; !ok {
			return fmt.Errorf("no encrypted secret found for node with p2pId: %s. Ensure secrets have been correctly encrypted for this DON", p2pId)
		}
	}

	// Check that the encryptionPublicKey values in the encryptedSecrets metadata match the keys in encryptionPublicKeys
	for p2pId, keyFromMetadata := range encryptedSecrets.Metadata.NodePublicEncryptionKeys {
		encryptionPublicKey, ok := encryptionPublicKeys[p2pId]
		if !ok {
			return fmt.Errorf("encryption key not found for node with p2pId: %s. Ensure secrets have been correctly encrypted for this DON", p2pId)
		}

		if keyFromMetadata != hex.EncodeToString(encryptionPublicKey[:]) {
			return fmt.Errorf("the encryption public key in the encrypted secrets metadata does not match the one in the workflow registry. Ensure secrets have been correctly encrypted for this DON")
		}
	}

	return nil
}
