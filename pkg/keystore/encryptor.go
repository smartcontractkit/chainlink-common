package keystore

import "context"

type EncryptRequest struct {
	KeyName string
	Data    []byte
}

type EncryptResponse struct {
	EncryptedData []byte
}

type DecryptRequest struct {
	KeyName       string
	EncryptedData []byte
}

type DecryptResponse struct {
	Data []byte
}

type Encryptor interface {
	Encrypt(ctx context.Context, req EncryptRequest) (EncryptResponse, error)
	Decrypt(ctx context.Context, req DecryptRequest) (DecryptResponse, error)
}
