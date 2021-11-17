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
type Keyrings = map[string]keyring.Keyring
type Signers = map[string]keyring.Signer

//	- This: keyring.Service (subsystem), manages multiple keyring.Keyring/s (backend, signing_algo)
//	- Deps:
//		- *gin.Engine
//		- *keyring.Keyrings (many)
//	- Ports:
//		- *keyring.Signers (one per keystore)
//	- Drivers:
//		- HTTP endpoint: /keys - manages config keys (CRUD + I/E)
//	- Services: none
//	- Emits:
//		- e.keyring.run_init
//		- e.keyring.running
//		- e.keyring.health
//		- e.keyring.key_created
//		- e.keyring.key_deleted
//		- e.keyring.key_imported
type Service struct {
	router   *gin.Engine
	keyrings Keyrings
	Signers  Signers
}

func NewService(router *gin.Engine, keyrings Keyrings) *Service {
	signers := make(Signers, len(keyrings))
	for k, v := range keyrings {
		signers[k] = &signer{v}
	}

	return &Service{
		router: router,
		// full keyring access should be limited to this service
		keyrings: keyrings,
		// the service exposes a map of available keyring.Signer/s
		Signers: signers,
	}
}

type signer struct {
	keyring keyring.Keyring
}

func (s signer) Sign(uid string, msg []byte) ([]byte, types.PubKey, error) {
	return s.Sign(uid, msg)
}

func (s signer) SignByAddress(address sdk.Address, msg []byte) ([]byte, types.PubKey, error) {
	return s.SignByAddress(address, msg)
}
