package confidentialrelay

const (
	MethodSecretsGet     = "confidential.secrets.get"
	MethodCapabilityExec = "confidential.capability.execute"

	DomainSecretsGet     = "ConfidentialSecretsGet"
	DomainCapabilityExec = "ConfidentialCapabilityExecute"
)

// SecretIdentifier identifies a secret by key and namespace.
type SecretIdentifier struct {
	Key       string `json:"key"`
	Namespace string `json:"namespace"`
}

// SecretsRequestParams is the JSON-RPC params for "confidential.secrets.get".
type SecretsRequestParams struct {
	WorkflowID          string             `json:"workflow_id"`
	WorkflowOwner       string             `json:"workflow_owner"`                // Ethereum address (hex, 0x-prefixed)
	WorkflowExecutionID string             `json:"workflow_execution_id"`         // 32 bytes, hex-encoded
	Secrets             []SecretIdentifier `json:"secrets"`
	EnclavePublicKey    string             `json:"enclave_public_key"`
	Attestation         string             `json:"attestation,omitempty"`
}

// SecretEntry is a single secret in the relay DON's response.
type SecretEntry struct {
	ID              SecretIdentifier `json:"id"`
	Ciphertext      string           `json:"ciphertext"`
	EncryptedShares []string         `json:"encrypted_shares"`
}

// SecretsResponseResult is the JSON-RPC result for "confidential.secrets.get".
// The enclave uses its own config for MasterPublicKey and threshold (config.T),
// so the relay handler only returns the encrypted shares per secret.
type SecretsResponseResult struct {
	Secrets []SecretEntry `json:"secrets"`
}

// CapabilityRequestParams is the JSON-RPC params for "confidential.capability.execute".
type CapabilityRequestParams struct {
	WorkflowID   string `json:"workflow_id"`
	CapabilityID string `json:"capability_id"`
	Payload      string `json:"payload"`
	Attestation  string `json:"attestation,omitempty"`
}

// CapabilityResponseResult is the JSON-RPC result for "confidential.capability.execute".
type CapabilityResponseResult struct {
	Payload string `json:"payload,omitempty"`
	Error   string `json:"error,omitempty"`
}
