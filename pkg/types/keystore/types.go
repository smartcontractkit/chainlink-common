package keystore

import "context"

// Keystore This interface is exposed to keystore consumers
type Keystore interface {
	Sign(ctx context.Context, keyID []byte, data []byte) ([]byte, error)
	SignBatch(ctx context.Context, keyID []byte, data [][]byte) ([][]byte, error)
	Verify(ctx context.Context, keyID []byte, data []byte) (bool, error)
	VerifyBatch(ctx context.Context, keyID []byte, data [][]byte) ([]bool, error)

	ListKeys(ctx context.Context, tags []string) ([][]byte, error)

	// RunUDF executes a user-defined function (UDF) on the keystore.
	// This method is designed to provide flexibility by allowing users to define custom
	// logic that can be executed without breaking the existing interface. While it enables
	// future extensibility, developers should ensure that UDF implementations are safe
	// and do not compromise the security of the keystore or the integrity of the data.
	RunUDF(ctx context.Context, name string, keyID []byte, data []byte) ([]byte, error)
}

// Management Core node exclusive
type Management interface {
	AddPolicy(ctx context.Context, policy []byte) (string, error)
	RemovePolicy(ctx context.Context, policyID string) error
	ListPolicy(ctx context.Context) []byte

	ImportKey(ctx context.Context, keyType string, data []byte, tags []string) ([]byte, error)
	ExportKey(ctx context.Context, keyID []byte) ([]byte, error)

	CreateKey(ctx context.Context, keyType string, tags []string) ([]byte, error)
	DeleteKey(ctx context.Context, keyID []byte) error

	AddTag(ctx context.Context, keyID []byte, tag string) error
	RemoveTag(ctx context.Context, keyID []byte, tag string) error
	ListTags(ctx context.Context, keyID []byte) ([]string, error)
}
