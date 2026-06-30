package starkkey

import (
	"errors"
	"io"
	"math/big"

	"github.com/NethermindEth/starknet.go/curve"

	"github.com/smartcontractkit/chainlink-common/keystore/internal"
)

// GenerateKey creates a Starknet key pair using starknet.go's GetRandomKeys.
//
// The io.Reader parameter is kept for API compatibility with other keystore key types;
// starknet.go v0.17 key generation uses crypto/rand internally.
func GenerateKey(_ io.Reader) (k Key, err error) {
	priv, x, y, err := curve.GetRandomKeys()
	if err != nil {
		return k, err
	}

	k.signFn = func(hash *big.Int) (x, y *big.Int, err error) {
		return curve.Sign(hash, priv)
	}

	k.pub.X = x
	k.pub.Y = y
	if k.pub.X == nil || k.pub.Y == nil {
		return k, errors.New("key gen is not on stark curve")
	}
	k.raw = internal.NewRaw(padBytes(priv.Bytes()))

	return k, nil
}

// pad bytes to privateKeyLen
func padBytes(a []byte) []byte {
	if len(a) < privateKeyLen {
		pad := make([]byte, privateKeyLen-len(a))
		return append(pad, a...)
	}

	// return original if length is >= to specified length
	return a
}
