package keystore

import "context"

type CreateKeyRequest struct {
	Name    string
	KeyType KeyType
}

type CreateKeyResponse struct {
	KeyInfo KeyInfo
}

type DeleteKeyRequest struct {
	Name string
}

type DeleteKeyResponse struct{}

type ImportKeyRequest struct {
	Name    string
	KeyType KeyType
	Data    []byte
}

type ImportKeyResponse struct {
	PublicKey []byte
}

type ExportKeyRequest struct {
	Name string
}

type ExportKeyResponse struct {
	Data []byte
}

type Admin interface {
	CreateKey(ctx context.Context, req CreateKeyRequest) (CreateKeyResponse, error)
	DeleteKey(ctx context.Context, req DeleteKeyRequest) (DeleteKeyResponse, error)
	ImportKey(ctx context.Context, req ImportKeyRequest) (ImportKeyResponse, error)
	ExportKey(ctx context.Context, req ExportKeyRequest) (ExportKeyResponse, error)
}
