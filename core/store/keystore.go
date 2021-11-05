package store

import (
	"fmt"

	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/chainlink/core/services/keystore"
	"github.com/smartcontractkit/chainlink/core/services/keystore/keys/ocrkey"
	"github.com/smartcontractkit/chainlink/core/services/keystore/keys/p2pkey"
	"github.com/smartcontractkit/chainlink/core/utils"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
	"golang.org/x/crypto/curve25519"
	"gorm.io/gorm"
)

// Note: this file could be significantly simplified if OCR2 keystore functionality works

func KeystoreInit(db *gorm.DB, password string) (keystore.Master, map[string]string, error) {
	log := logger.Default.Named("keystore")
	keys := keystore.New(db, utils.DefaultScryptParams)
	if err := keys.Unlock(password); err != nil {
		return keys, map[string]string{}, err
	}

	// // OCR2 keystore (having issues with it...temporary solution by wrapping OCR keys to meet expected interface)
	// key, err := keys.OCR2().Create() // properly stores to DB
	// logger.Info(key, err)
	// ocr2All, err := keys.OCR2().GetAll() // works if keys are created, but doesn't properly fetch from the DB...
	// logger.Info(ocr2All, err)

	// create OCR key if necessary
	var ocrKey ocrkey.KeyV2
	ocrKeys, err := keys.OCR().GetAll()
	if err != nil {
		return keys, map[string]string{}, err
	}
	if len(ocrKeys) == 0 {
		log.Info("No OCR key found. Creating new OCR key")
		ocrKey, err = keys.OCR().Create()
		if err != nil {
			return keys, map[string]string{}, err
		}
	} else {
		ocrKey = ocrKeys[0]
	}

	// create P2P key if necessary
	var p2pKey p2pkey.KeyV2
	p2pKeys, err := keys.P2P().GetAll()
	if err != nil {
		return keys, map[string]string{}, err
	}
	if len(p2pKeys) == 0 {
		log.Info("No P2P key found. Creating new P2P key")
		p2pKey, err = keys.P2P().Create()
		if err != nil {
			return keys, map[string]string{}, err
		}
	} else {
		p2pKey = p2pKeys[0]
	}

	// print OCR key + P2P key information
	fmt.Printf("OCR Key Bundle:\n  ID: %s\n  Config Public Key: %s\n  Offchain Public Key: %s\n", ocrKey.ID(), ocrkey.ConfigPublicKey(ocrKey.PublicKeyConfig()), ocrKey.OffChainSigning.PublicKey())
	fmt.Printf("P2P Key Bundle:\n  Peer ID: %s\n  Public Key: %s\n", p2pKey.ID(), p2pKey.PublicKeyHex())

	// returns a map of keys (later used to retrieve keys from relay endpoint)
	return keys, map[string]string{
		"OCRKeyID":             ocrKey.ID(),
		"OCRConfigPublicKey":   ocrkey.ConfigPublicKey(ocrKey.PublicKeyConfig()).String(),
		"OCROffchainPublicKey": ocrKey.OffChainSigning.PublicKey().String(),
		"P2PID":                p2pKey.ID(),
		"P2PPublicKey":         p2pKey.PublicKeyHex(),
	}, nil
}

type OCR2KeyWrapper struct {
	ocr ocrkey.KeyV2
}

func NewOCR2KeyWrapper(key ocrkey.KeyV2) OCR2KeyWrapper {
	return OCR2KeyWrapper{key}
}

func (o OCR2KeyWrapper) OffchainSign(msg []byte) (signature []byte, err error) {
	return o.ocr.SignOffChain(msg)
}

func (o OCR2KeyWrapper) ConfigDiffieHellman(point [curve25519.PointSize]byte) (sharedPoint [curve25519.PointSize]byte, err error) {
	out, err := o.ocr.ConfigDiffieHellman(&point)
	if err != nil {
		return [curve25519.PointSize]byte{}, err
	}
	return *out, err
}

func (o OCR2KeyWrapper) OffchainPublicKey() types.OffchainPublicKey {
	return types.OffchainPublicKey(o.ocr.PublicKeyOffChain())
}

func (o OCR2KeyWrapper) ConfigEncryptionPublicKey() types.ConfigEncryptionPublicKey {
	return o.ocr.PublicKeyConfig()
}
