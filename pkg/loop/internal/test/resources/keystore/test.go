package keystore_test

import (
	libocr "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

const (
	account = libocr.Account("testaccount")
)

var (
	encoded = []byte{5: 11}
	signed  = []byte{13: 37}

	DefaultKeystoreTestConfig = StaticKeystoreConfig{
		Account: string(account),
		Encoded: encoded,
		Signed:  signed,
	}
)
