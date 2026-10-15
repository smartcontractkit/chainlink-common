package vault

import (
	"errors"
	"strings"
)

const SecretVersionSkewMessage = "vault: secret version skew across vault nodes"

// ErrSecretVersionSkew is returned for direct GetSecrets requests when the Vault
// nodes did not agree on a single ciphertext for a secret. The request can be
// retried under a new request ID.
var ErrSecretVersionSkew = errors.New(SecretVersionSkewMessage)

func IsSecretVersionSkew(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrSecretVersionSkew) || strings.Contains(err.Error(), SecretVersionSkewMessage)
}
