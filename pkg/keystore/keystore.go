package keystore

import "context"

type KeyType string

const (
	// For simplicity, just use a single string to represent the key type
	// including curve and algo. These 2 types should cover the vast majority of use cases.
	// But we can add more types as needed.
	Ed25519   KeyType = "ed25519"
	Secp256k1 KeyType = "secp256k1"
)

type KeyInfo struct {
	Name      string
	KeyType   KeyType
	PublicKey []byte
}

type Writer interface {
	// Name is a user supplied name of the key. It is used to identify the key in the keystore.
	// Clients define the naming scheme relevant for them.
	CreateKey(ctx context.Context, name string, keyType KeyType) (KeyInfo, error)
	DeleteKey(ctx context.Context, name string) error
	ImportKey(ctx context.Context, name string, keyType KeyType, data []byte) ([]byte, error)
}

type Reader interface {
	ListKeys(ctx context.Context) ([]KeyInfo, error)
	GetKey(ctx context.Context, name string) (KeyInfo, error)
}

type Signer interface {
	Sign(ctx context.Context, name string, data []byte) ([]byte, error)
	Verify(ctx context.Context, name string, data []byte, signature []byte) (bool, error)
}

type Keystore interface {
	Writer
	Reader
	Signer
}
