package ethkey

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/google/uuid"

	commonkeystore "github.com/smartcontractkit/chainlink-common/keystore"
)

func (key KeyV2) ToEncryptedJSON(password string, scryptParams commonkeystore.ScryptParams) (export []byte, err error) {
	// DEV: uuid is derived directly from the address, since it is not stored internally
	id, err := uuid.FromBytes(key.Address.Bytes()[:16])
	if err != nil {
		return nil, fmt.Errorf("could not generate ethkey UUID: %w", err)
	}
	dKey := &keystore.Key{
		Id:         id,
		Address:    key.Address,
		PrivateKey: key.getPK(),
	}
	return keystore.EncryptKey(dKey, password, scryptParams.N, scryptParams.P)
}
