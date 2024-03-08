package ccip

import (
	"encoding/hex"
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/crypto/sha3"
)

// Address is the generic address type used across CCIP plugins.
//
// NOTE: JSON codec had to be overridden for CCIP backwards compatibility.
// Before this generic address type, CCIP was using common.Address from go-ethereum library which marshals
// to lower-case and prints as EIP55. We have to maintain this behavior to keep nodes that run different
// versions to come in consensus.
type Address string

func (a *Address) UnmarshalJSON(bytes []byte) error {
	vStr := strings.Trim(string(bytes), `"`)

	eip55, err := EIP55(vStr)
	if err != nil {
		*a = Address(vStr)
	} else {
		*a = Address(eip55)
	}

	return nil
}

func (a Address) MarshalJSON() ([]byte, error) {
	return []byte(`"` + strings.ToLower(string(a)) + `"`), nil
}

func (a Address) MarshalText() (text []byte, err error) {
	return []byte(strings.ToLower(string(a))), nil
}

func (a *Address) UnmarshalText(text []byte) error {
	vStr := string(text)

	eip55, err := EIP55(vStr)
	if err != nil {
		*a = Address(vStr)
	} else {
		*a = Address(eip55)
	}

	return nil
}

type Hash [32]byte

func (h Hash) String() string {
	return "0x" + hex.EncodeToString(h[:])
}

type TxMeta struct {
	BlockTimestampUnixMilli int64
	BlockNumber             uint64
	TxHash                  string
	LogIndex                uint64
}

// EIP55 converts the provided address to EIP55.
//
// If the ith digit is a letter (ie. it’s one of abcdef) print it in uppercase if the 4*ith bit of the hash of the
// lowercase hexadecimal address is 1 otherwise print it in lowercase.
func EIP55(addr string) (string, error) {
	addr = strings.ToLower(strings.TrimPrefix(addr, "0x"))
	if len(addr) != 40 {
		return "", fmt.Errorf("address is not the correct length")
	}

	keccak256 := sha3.NewLegacyKeccak256()
	keccak256.Write([]byte(addr))
	addrHash := hex.EncodeToString(keccak256.Sum(nil))

	addrEIP55 := "0x"
	for i, c := range addr {
		isAbcdef := c >= 'a' && c <= 'f'
		isDigit := c >= '0' && c <= '9'

		if !isAbcdef && !isDigit {
			return "", fmt.Errorf("address contains a character that is not evm specific")
		}

		if i >= len(addrHash) {
			return "", fmt.Errorf("invalid address hash")
		}

		if isAbcdef && addrHash[i] >= '8' {
			addrEIP55 += string(unicode.ToUpper(c))
			continue
		}
		addrEIP55 += string(c)
	}

	return addrEIP55, nil
}
