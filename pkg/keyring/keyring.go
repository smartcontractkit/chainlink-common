package keyring

import (
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gin-gonic/gin"
)

var (
	_ keyring.Signer = (*signer)(nil)
)

// TODO: suggestions? (will be used to store Chainlink CSA keys, so not only network specific)
// Schema: <net>.<use>.<label>/<backend> (@see keyring_test.go for examples)
type Keyrings = map[string]keyring.Keyring // TODO: pointers to Keyring, to avoid backend db (mem) copying?
type KeyringsPub = map[string]keyringPub
type Signers = map[string]keyring.Signer

// keyringPub is providing keyring types.PubKey/s access.
type keyringPub struct {
	keyring keyring.Keyring
}

// Get gets the public key with a user key.
func (p keyringPub) Get(uid string) (types.PubKey, error) {
	// sign an empty msg and extract the types.PubKey
	_, pk, err := p.keyring.Sign(uid, []byte{})
	if err != nil {
		return nil, err
	}

	return pk, nil
}

// GetByAddress gets the public key with a user key providing the address.
func (p keyringPub) GetByAddress(address sdk.Address, msg []byte) (types.PubKey, error) {
	// sign an empty msg and extract the types.PubKey
	_, pk, err := p.keyring.SignByAddress(address, []byte{})
	if err != nil {
		return nil, err
	}

	return pk, nil
}

type signer struct {
	keyring keyring.Keyring
}

func (s signer) Sign(uid string, msg []byte) ([]byte, types.PubKey, error) {
	return s.keyring.Sign(uid, msg)
}

func (s signer) SignByAddress(address sdk.Address, msg []byte) ([]byte, types.PubKey, error) {
	return s.keyring.SignByAddress(address, msg)
}

//	- This: keyring.Service (subsystem), manages multiple keyring.Keyring/s (backend, signing_algo)
//	- Deps:
//		- *gin.Engine
//		- *keyring.Keyrings (many)
//	- Ports:
//		- keyring.KeyringsPub (one per keystore)
//		- keyring.Signers     (one per keystore)
//	- Drivers:
//		- HTTP endpoint: /keys - manages config keys (CRUD + I/E)
//		- GRPC endpoint: /keys - manages config keys (CRUD + I/E)
//	- Services: none
//	- Emits:
//		- e.keyring.run_init
//		- e.keyring.running
//		- e.keyring.health
//		- e.keyring.key_created
//		- e.keyring.key_deleted
//		- e.keyring.key_imported
type Service struct {
	router      *gin.Engine
	keyrings    Keyrings
	KeyringsPub KeyringsPub
	Signers     Signers
}

func NewService(router *gin.Engine, keyrings Keyrings) *Service {
	n := len(keyrings)
	keyringsPub := make(KeyringsPub, n)
	signers := make(Signers, n)
	for k, v := range keyrings {
		keyringsPub[k] = keyringPub{v}
		signers[k] = signer{v}
	}

	return &Service{
		router: router,
		// full keyring access (write) should be limited to this service
		keyrings: keyrings,
		// exposes a map of available keyringPub/s (public keyrings)
		KeyringsPub: keyringsPub,
		// exposes a map of available keyring.Signer/s
		Signers: signers,
	}
}
