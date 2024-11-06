package keystore

import (
	"context"

	"github.com/smartcontractkit/chainlink-common/pkg/services"
)

// GRPCService This interface contains all the functionalities of the GRPC layer of the LOOPP keystore
type GRPCService interface {
	services.Service
	Sign(ctx context.Context, keyID []byte, data []byte) ([]byte, error)
	SignBatch(ctx context.Context, keyID []byte, data [][]byte) ([][]byte, error)
	Verify(ctx context.Context, keyID []byte, data []byte) (bool, error)
	VerifyBatch(ctx context.Context, keyID []byte, data [][]byte) ([]bool, error)

	ListKeys(ctx context.Context, tags []string) ([][]byte, error)

	RunUDF(ctx context.Context, name string, keyID []byte, data []byte) ([]byte, error)

	ImportKey(ctx context.Context, keyType string, data []byte, tags []string) ([]byte, error)
	ExportKey(ctx context.Context, keyID []byte) ([]byte, error)

	CreateKey(ctx context.Context, keyType string, tags []string) ([]byte, error)
	DeleteKey(ctx context.Context, keyID []byte) error

	AddTag(ctx context.Context, keyID []byte, tag string) error
	RemoveTag(ctx context.Context, keyID []byte, tag string) error
	ListTags(ctx context.Context, keyID []byte) ([]string, error)
}
