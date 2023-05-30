package starknet

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/smartcontractkit/chainlink-relay/pkg/loop/internal/pb"
)

const (
	maxPointByteLen = 32 // stark curve max is 252 bits
	signatureLen    = 2 * maxPointByteLen
)

type Signature struct {
	x, y *pb.BigInt
}

// encodes starknet doublet of big.Ints into []byte slice
// the first 32 bytes are the padded bytes of x
// the second 32 bytes are the padded bytes of y
func (s *Signature) Bytes() ([]byte, error) {
	buf := new(bytes.Buffer)
	n, err := buf.Write(padBytes(s.x.Int().Bytes(), maxPointByteLen))
	if err != nil {
		return nil, fmt.Errorf("error writing 'x' component of signature: %w", err)
	}
	if n != maxPointByteLen {
		return nil, fmt.Errorf("unexpected write length of 'x' component of signature: wrote %d expected %d", n, maxPointByteLen)
	}

	n, err = buf.Write(padBytes(s.y.Int().Bytes(), maxPointByteLen))
	if err != nil {
		return nil, fmt.Errorf("error writing 'y' component of signature: %w", err)
	}
	if n != maxPointByteLen {
		return nil, fmt.Errorf("unexpected write length of 'y' component of signature: wrote %d expected %d", n, maxPointByteLen)
	}

	if buf.Len() != signatureLen {
		return nil, fmt.Errorf("error in signature length")
	}
	return buf.Bytes(), nil
}

func (s *Signature) Ints() (x *big.Int, y *big.Int, err error) {
	if s.x == nil || s.y == nil {
		return nil, nil, fmt.Errorf("signature uninitialized")
	}

	return s.x.Int(), s.y.Int(), nil
}

// b is expected to encode x,y components in accordance with [signature.Bytes]
func SignatureFromBytes(b []byte) (*Signature, error) {

	if len(b) != signatureLen {
		return nil, fmt.Errorf("expected signature length %d got %d", signatureLen, len(b))
	}

	x := b[:maxPointByteLen]
	y := b[maxPointByteLen:]

	return SignatureFromBigInts(
		new(big.Int).SetBytes(x),
		new(big.Int).SetBytes(y))
}

// x,y must be non-negative numbers
func SignatureFromBigInts(x *big.Int, y *big.Int) (*Signature, error) {

	if x.Cmp(big.NewInt(0)) < 0 || y.Cmp(big.NewInt(0)) < 0 {
		return nil, fmt.Errorf("Cannot create signature from negative values (x,y), (%v, %v)", x, y)
	}

	return &Signature{
		x: pb.NewBigIntFromInt(x),
		y: pb.NewBigIntFromInt(y),
	}, nil

}

// pad bytes to specific length
func padBytes(a []byte, length int) []byte {
	if len(a) < length {
		pad := make([]byte, length-len(a))
		return append(pad, a...)
	}

	// return original if length is >= to specified length
	return a
}
