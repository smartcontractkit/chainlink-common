package chains

import (
	"context"
	"math/big"

	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
)

type InstanceID string

type ReadEntity struct {
	Namespace  string
	Identifier string
}

type ReadEntityInstance struct {
	ReadEntity
	InstanceID InstanceID
}

type WriteOperation struct {
	Namespace string
}

type WriteOperationInstance struct {
	WriteOperation
	InstanceID InstanceID
}

// univocally identifies a chain
type ChainID uint

// Can be any struct since it's defined by the chain specific implementation. It contains all the information necessary for deployment, read and write operations for a particular product
type ChainConfig any

// Can be any struct since it's defined by the product specific implementation. When creating a Chain from a product, the Chain will have a ChainConfig that described how to process the data in ProductConfig.
type ProductConfig any

// Represents the state of a product deployment with the set of read entities and write operations instances created during deployment
type Deployment struct {
	readEntityInstances     map[ReadEntity]InstanceID
	writeOperationInstances map[WriteOperation]InstanceID
}

func (d Deployment) getReadEntityInstance(entity ReadEntity) ReadEntityInstance {
	instance := d.readEntityInstances[entity]
	return ReadEntityInstance{entity, instance}
}

func (d Deployment) getWriteOperationInstance(entity WriteOperation) WriteOperationInstance {
	instance := d.writeOperationInstances[entity]
	return WriteOperationInstance{entity, instance}
}

// Clients can use this interface to discover a set of ChainFactory and rely in ChainFactory.SupportsChain to get an instance of a Chain
// TODO: define discovery mechanism
type ChainFactory interface {
	// True if supports the chain with ID chainID
	SupportsChain(chainID ChainID) bool

	// Instantiates a new Chain with the given config
	CreateChain(chainID ChainID, chainConfig ChainConfig) Chain
}

// Abstract interface for all chains. An instance of a Chain gets created using a ChainConfig that's configured for a product.
type Chain interface {
	// Deploys the product
	Deploy(productID string, ProductConfig ProductConfig) Deployment

	// Bind functionality - associate a previous deployment
	Bind(productID string, deployment Deployment) (Deployment, error)

	// Read ReadEntity
	GetLatestValue(ctx context.Context, instance ReadEntityInstance, confidenceLevel primitives.ConfidenceLevel, params, entity any) error

	// Batch ReadEntity
	BatchGetLatestValues(ctx context.Context, request BatchGetLatestValuesRequest) (BatchGetLatestValuesResult, error)

	// Query ReadEntity
	QueryKey(ctx context.Context, instance ReadEntityInstance, filter query.KeyFilter, limitAndSort query.LimitAndSort, sequenceDataType any) ([]Sequence, error)

	// Write based on operations with input parameters
	Write(ctx context.Context, instance WriteOperationInstance, args any, transactionID string, meta *TxMeta) error

	// Get transaction status created during Write
	GetTransactionStatus(ctx context.Context, transactionID string) (TransactionStatus, error)

	// Get Fee Components
	GetFeeComponents(ctx context.Context) (*ChainFeeComponents, error)
}

type chain struct{}

func (c chain) Deploy(productID string, ProductConfig any) Deployment {
	// Lookup in the chain config the deployment information using the productID, deploy on-chain components and configure using ProductConfig
	return Deployment{}
}

// BatchGetLatestValuesRequest string is contract name.
type (
	BatchGetLatestValuesRequest map[ReadEntityInstance]ContractBatch
	ContractBatch               []BatchRead
	BatchRead                   struct {
		Params    any
		ReturnVal any
	}
)

type (
	BatchGetLatestValuesResult map[ReadEntityInstance]ContractBatchResults
	ContractBatchResults       []BatchReadResult
	BatchReadResult            struct {
		returnValue any
		err         error
	}
)

type Sequence struct {
	// This way we can retrieve past/future sequences (EVM log events) very granularly, but still hide the chain detail.
	Cursor string
	Head
	Data any
}

type Head struct {
	Height string
	Hash   []byte
	// Timestamp is in Unix time
	Timestamp uint64
}

type TxMeta struct {
	// Used for Keystone Workflows
	WorkflowExecutionID *string
	// An optional maximum gas limit for the transaction. If not set the ChainWriter implementation will be responsible for
	// setting a gas limit for the transaction.  If it is set and the ChainWriter implementation does not support setting
	// this value per transaction it will return ErrSettingTransactionGasLimitNotSupported
	GasLimit *big.Int
}

// TransactionStatus are the status we expect every TXM to support and that can be returned by StatusForUUID.
type TransactionStatus int

const (
	Unknown TransactionStatus = iota
	Pending
	Unconfirmed
	Finalized
	Failed
	Fatal
)

// ChainFeeComponents contains the different cost components of executing a transaction.
type ChainFeeComponents struct {
	// The cost of executing transaction in the chain's EVM (or the L2 environment).
	ExecutionFee *big.Int

	// The cost associated with an L2 posting a transaction's data to the L1.
	DataAvailabilityFee *big.Int
}
