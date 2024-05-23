package hashlib

// Hash contains all supported hash formats.
// Add additional hash types e.g. [20]byte as needed here.
type Hash interface {
	[32]byte
}

type Ctx[H Hash] interface {
	Hash(l []byte) H
	HashInternal(a, b H) H
	ZeroHash() H
}
