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

func (k *keystore) Encrypt(ctx context.Context, req EncryptRequest) (EncryptResponse, error) {
	return EncryptResponse{}, nil
}

func (k *keystore) Decrypt(ctx context.Context, req DecryptRequest) (DecryptResponse, error) {
	return DecryptResponse{}, nil
}
