package enclaverelay

const (
	MethodSecretsGet     = "enclave.secrets.get"
	MethodCapabilityExec = "enclave.capability.execute"

	DomainSecretsGet     = "EnclaveSecretsGet"
	DomainCapabilityExec = "EnclaveCapabilityExecute"
)

// SecretIdentifier identifies a secret by key and namespace.
type SecretIdentifier struct {
	Key       string `json:"key"`
	Namespace string `json:"namespace"`
}

// SecretsRequestParams is the JSON-RPC params for "enclave.secrets.get".
type SecretsRequestParams struct {
	WorkflowID       string             `json:"workflow_id"`
	Secrets          []SecretIdentifier `json:"secrets"`
	EnclavePublicKey string             `json:"enclave_public_key"`
	Attestation      string             `json:"attestation,omitempty"`
}

// SecretEntry is a single secret in the relay DON's response.
type SecretEntry struct {
	ID              SecretIdentifier `json:"id"`
	Ciphertext      string           `json:"ciphertext"`
	EncryptedShares []string         `json:"encrypted_shares"`
}

// SecretsResponseResult is the JSON-RPC result for "enclave.secrets.get".
type SecretsResponseResult struct {
	Secrets         []SecretEntry `json:"secrets"`
	MasterPublicKey string        `json:"master_public_key"`
	Threshold       int           `json:"threshold"`
}

// CapabilityRequestParams is the JSON-RPC params for "enclave.capability.execute".
type CapabilityRequestParams struct {
	WorkflowID   string `json:"workflow_id"`
	CapabilityID string `json:"capability_id"`
	Payload      string `json:"payload"`
	Attestation  string `json:"attestation,omitempty"`
}

// CapabilityResponseResult is the JSON-RPC result for "enclave.capability.execute".
type CapabilityResponseResult struct {
	Payload string `json:"payload,omitempty"`
	Error   string `json:"error,omitempty"`
}
