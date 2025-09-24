package keystore

type KeyType string

type KeyInfo struct {
	Name      string
	KeyType   KeyType
	PublicKey []byte
}

type Keystore interface {
	Admin
	Reader
	Signer
	Encryptor
	KeyExchanger
}
