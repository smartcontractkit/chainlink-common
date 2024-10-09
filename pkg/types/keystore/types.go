package keystore

import "context"

type Keystore interface {
	Sign(ctx context.Context, keyID []byte, data []byte) ([]byte, error)
	SignBatch(ctx context.Context, keyID []byte, data [][]byte) ([][]byte, error)
	Verify(ctx context.Context, keyID []byte, data []byte) (bool, error)
	VerifyBatch(ctx context.Context, keyID []byte, data [][]byte) ([]bool, error)

	Get(ctx context.Context, tags []string) ([][]byte, error)

	RunUDF(ctx context.Context, udfName string, keyID []byte, data []byte) ([]byte, error)
}

// Management Core node exclusive
type Management interface {
	AddPolicy(ctx context.Context, policy []byte) (string, error)
	RemovePolicy(ctx context.Context, policyID string) error
	ListPolicy(ctx context.Context) []byte

	Import(ctx context.Context, keyType string, data []byte, tags []string) ([]byte, error)
	Export(ctx context.Context, keyID []byte) ([]byte, error)

	Create(ctx context.Context, keyType string, tags []string) ([]byte, error)
	Delete(ctx context.Context, keyID []byte) error

	AddTag(ctx context.Context, keyID []byte, tag string) error
	RemoveTag(ctx context.Context, keyID []byte, tag string) error
	ListTags(ctx context.Context, keyID []byte) ([]string, error)
}
