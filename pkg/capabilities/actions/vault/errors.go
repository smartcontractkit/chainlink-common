package vault

import "errors"

// ErrSecretVersionSkew is returned for direct GetSecrets requests when the Vault
// nodes did not agree on a single ciphertext for a secret. The request can be
// retried under a new request ID.
var ErrSecretVersionSkew = errors.New("vault: secret version skew across vault nodes")
