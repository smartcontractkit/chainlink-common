package stellar

type ContractWriterConfig struct {
	Contracts map[string]*ContractWriterContract `json:"contracts"`
}

type ContractWriterContract struct {
	// The contract name (optional). When not provided, the key in the map under which this contract
	// is stored is used.
	Name      string                             `json:"name,omitempty"`
	Functions map[string]*ContractWriterFunction `json:"functions"`
}

type ContractWriterFunction struct {
	FromAddress string `json:"fromAddress"`
}
