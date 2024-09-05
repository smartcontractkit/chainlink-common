package auth

import (
	"encoding/gob"
	"errors"
	"os"
	"path"
)

// Token represents the token used to authorize the requests
// to access protected resources on the provider's backend.
type Token interface {
	Save(token string) error
	Get() (string, error)
	Delete() error
}

// TokenStore is the on-disk cache that stores the token
type tokenStore struct {
	rootPath string
}

// Save saves the supplied token to the tokenStore
func (t tokenStore) Save(token string) error {
	f, err := os.Create(t.cachePath())
	if err != nil {
		return err
	}
	encoder := gob.NewEncoder(f)
	if err = encoder.Encode(token); err != nil {
		return err
	}
	err = f.Sync()
	if err != nil {
		return nil
	}
	return f.Close()
}

// Get returns any saved tokens
func (t tokenStore) Get() (string, error) {
	f, err := os.Open(t.cachePath())
	if err != nil {
		// If it doesn't exist, it means that the user has never been authenticated
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", errors.New("unable to read cookie cache")
	}

	decoder := gob.NewDecoder(f)
	var token string
	if err = decoder.Decode(&token); err != nil {
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	return token, nil
}

// Delete clears the cached token
func (t tokenStore) Delete() error {
	return os.Remove(t.cachePath())
}

func (t tokenStore) cachePath() string {
	return path.Join(t.rootPath, "/authToken.cache")
}

// newTokenStore creates a new token store in the same directory
// as the provided configuration file path.
func NewTokenStore(cfgFilePath string) *tokenStore {
	return &tokenStore{
		rootPath: path.Dir(cfgFilePath),
	}
}
