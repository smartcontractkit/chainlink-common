package solana

type SubmitTransactionRequest struct {
	Receiver           PublicKey
	EncodedTransaction string // base64 encoded transaction
}

// TransactionStatus is the result of the transaction sent to the chain
type TransactionStatus int

const (
	// Transaction processing failed due to a network issue, RPC issue, or other fatal error
	TxFatal TransactionStatus = iota
	// Transaction was sent successfully to the chain but the program execution was aborted
	TxAborted
	// Transaction was sent successfully to the chain, program was succesfully executed and mined into a block.
	TxSuccess
)

type ComputeConfig struct {
	// Default to nil. If not specified the value configured in GasEstimator will be used
	ComputeLimit *uint32
	// Default to nil. If not specified the value configured in GasEstimator will be used
	ComputeMaxPrice *uint64
}

type SubmitTransactionReply struct {
	Signature      Signature
	IdempotencyKey string
	Status         TransactionStatus
}
