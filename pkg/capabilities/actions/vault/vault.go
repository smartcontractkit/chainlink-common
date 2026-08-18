//go:generate protoc --go_out=../../.. --go_opt=paths=source_relative  --proto_path=../../.. capabilities/actions/vault/messages.proto
package vault

const (
	CapabilityID = "vault@1.0.0"
	// Deprecated - use the types in core instead.
	MethodGetSecrets = "vault.secrets.get"
)

type GetPublicKeyRequest struct{}

type GetPublicKeyResponse struct {
	PublicKey string `json:"publicKey"`
}
