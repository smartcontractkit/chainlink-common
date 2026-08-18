package host

import "context"

type EncryptionKeyFetcher interface {
	GetEncryptionKeys(ctx context.Context) ([]string, error)
}
