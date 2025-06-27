package aptos

type FeeStrategy int

const (
	DeprioritizedFeeStrategy FeeStrategy = -1
	DefaultFeeStrategy       FeeStrategy = 0
	PrioritizedFeeStrategy   FeeStrategy = 1
)

type ChainWriterConfig struct {
	Modules     map[string]*ChainWriterModule
	FeeStrategy FeeStrategy
}

type ChainWriterModule struct {
	// The module name (optional). When not provided, the key in the map under which this module
	// is stored is used.
	Name      string
	Functions map[string]*ChainWriterFunction
}

type ChainWriterFunction struct {
	// The function name (optional). When not provided, the key in the map under which this function
	// is stored is used.
	Name string
	// The public key of the sending account.
	PublicKey string
	// The account address (optional). When not provided, the address is calculated
	// from the public key.
	FromAddress string
	Params      []FunctionParam
}
