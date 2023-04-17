package loop

import "math/big"

// BigToBytes returns an encoded [*big.Int] or nil when nil.
// The encoding uses [*big.Int.Bytes] and a leading sign byte.
func BigToBytes(b *big.Int) []byte {
	if b == nil {
		return nil
	}
	var neg byte
	if b.Sign() < 0 {
		neg = 1
	}
	return append([]byte{neg}, b.Bytes()...)
}

// BigFromBytes returns a decoded [*big.Int] or nil.
func BigFromBytes(b []byte) *big.Int {
	if len(b) == 0 {
		return nil
	}
	neg := b[0] == 1
	i := new(big.Int)
	i.SetBytes(b[1:])
	if neg {
		i = i.Neg(i)
	}
	return i
}
