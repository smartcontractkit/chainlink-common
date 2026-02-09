package scrypt

import keystore2 "github.com/ethereum/go-ethereum/accounts/keystore"

type ScryptParams struct {
	N int
	P int
}

var (
	DefaultScryptParams = ScryptParams{
		N: keystore2.StandardScryptN,
		P: keystore2.StandardScryptP,
	}
	FastScryptParams ScryptParams = ScryptParams{
		N: 1 << 14,
		P: 1,
	}
)
