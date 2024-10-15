package keystore

import (
	"context"

	"github.com/smartcontractkit/chainlink-common/pkg/services"
)

// KeystoreMethods This interface contains all the functionalities of the GRPC layer of the LOOPP keystore
type KeystoreMethods interface {
	services.Service
	Sign(ctx context.Context, keyID []byte, data []byte) ([]byte, error)
	SignBatch(ctx context.Context, keyID []byte, data [][]byte) ([][]byte, error)
	Verify(ctx context.Context, keyID []byte, data []byte) (bool, error)
	VerifyBatch(ctx context.Context, keyID []byte, data [][]byte) ([]bool, error)

	List(ctx context.Context, tags []string) ([][]byte, error)

	RunUDF(ctx context.Context, udfName string, keyID []byte, data []byte) ([]byte, error)

	Import(ctx context.Context, keyType string, data []byte, tags []string) ([]byte, error)
	Export(ctx context.Context, keyID []byte) ([]byte, error)

	Create(ctx context.Context, keyType string, tags []string) ([]byte, error)
	Delete(ctx context.Context, keyID []byte) error

	AddTag(ctx context.Context, keyID []byte, tag string) error
	RemoveTag(ctx context.Context, keyID []byte, tag string) error
	ListTags(ctx context.Context, keyID []byte) ([]string, error)
}
