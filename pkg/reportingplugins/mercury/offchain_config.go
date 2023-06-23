package mercury

type OffchainConfig struct{}

func DecodeOffchainConfig(b []byte) (o OffchainConfig, err error) {
	return
}

func (c OffchainConfig) Encode() []byte {
	return []byte{}
}
