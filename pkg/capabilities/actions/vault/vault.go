//go:generate protoc --go_out=../../.. --go_opt=paths=source_relative  --proto_path=../../.. capabilities/actions/vault/messages.proto
package vault

const (
	CapabilityID     = "vault@1.0.0"
	MethodGetSecrets = "GetSecrets"
)
