package hex

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
)

// EnsurePrefix adds the prefix (0x) to a given hex string.
func EnsurePrefix(str string) string {
	if !strings.HasPrefix(str, "0x") {
		str = "0x" + str
	}
	return str
}

// ToBig parses the given hex string or panics if it is invalid.
func ToBig(s string) *big.Int {
	n, ok := new(big.Int).SetString(s, 16)
	if !ok {
		panic(fmt.Errorf(`failed to convert "%s" as hex to big.Int`, s))
	}
	return n
}

// RemovePrefix removes the prefix (0x) of a given hex string.
func RemovePrefix(str string) string {
	if HasPrefix(str) {
		return str[2:]
	}
	return str
}

// HasPrefix returns true if the string starts with 0x.
func HasPrefix(str string) bool {
	return len(str) >= 2 && str[0] == '0' && (str[1] == 'x' || str[1] == 'X')
}

// TryParse parses the given hex string to bytes,
// it can return error if the hex string is invalid.
// Follows the semantic of ethereum's FromHex.
func TryParse(s string) (b []byte, err error) {
	if !HasPrefix(s) {
		err = errors.New("hex string must have 0x prefix")
	} else {
		s = s[2:]
		if len(s)%2 == 1 {
			s = "0" + s
		}
		b, err = hex.DecodeString(s)
	}
	return
}
