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
// material is kept for API compatibility with other keystore key types; starknet.go
// v0.17 key generation uses crypto/rand internally.
func GenerateKey(material io.Reader) (k Key, err error) {
	_ = material

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
		padLen := privateKeyLen - len(a)
		out := make([]byte, privateKeyLen)
		copy(out[padLen:], a)
		return out
	}

	// return original if length is >= to specified length
	return a
}
