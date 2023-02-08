package mercury

import (
	"math/big"

	"github.com/smartcontractkit/libocr/bigbigendian"
)

// Bounds on an int192
const byteWidthInt192 = 24
const bitWidthInt192 = byteWidthInt192 * 8

var one *big.Int = big.NewInt(1)

// 2**191-1
func MaxValueInt192() *big.Int {
	result := MinValueInt192()
	result.Abs(result)
	result.Sub(result, one)
	return result
}

// -2**191
func MinValueInt192() *big.Int {
	result := &big.Int{}
	result.Lsh(one, bitWidthInt192-1)
	result.Neg(result)
	return result
}

// Encodes a value using 24-byte big endian two's complement representation. This function never panics.
func EncodeValueInt192(i *big.Int) ([]byte, error) {
	return bigbigendian.SerializeSigned(byteWidthInt192, i)
}

// Decodes a value using 24-byte big endian two's complement representation. This function never panics.
func DecodeValueInt192(s []byte) (*big.Int, error) {
	return bigbigendian.DeserializeSigned(byteWidthInt192, s)
}
