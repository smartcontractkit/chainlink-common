package solana

const (
	PublicKeyLength = 32
)

// represents solana-style AccountsMeta
type AccountMeta struct {
	PublicKey  [PublicKeyLength]byte
	IsWritable bool
	IsSigner   bool
}

type AccountMetaSlice []*AccountMeta
