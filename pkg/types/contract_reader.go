package types

import (
	"context"

	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
)

// Errors exposed to product plugins
const (
	ErrInvalidType                 = InvalidArgumentError("invalid type")
	ErrInvalidConfig               = InvalidArgumentError("invalid configuration")
	ErrContractReaderConfigMissing = UnimplementedError("ContractReader entry missing from RelayConfig")
	ErrInternal                    = InternalError("internal error")
	ErrNotFound                    = NotFoundError("not found")
)

/*
type ContractReader interface {
	services.Service
	GetLatestValue(ctx context.Context, readIdentifier string, confidenceLevel primitives.ConfidenceLevel, params, returnVal any) error

	BatchGetLatestValues(ctx context.Context, request BatchGetLatestValuesRequest) (BatchGetLatestValuesResult, error)

}
*/

// ContractReader defines essential read operations a chain should implement for reading contract values and events.
type ContractReader struct {
	services.Service
}

// GetLatestValue gets the latest value with a certain confidence level that maps to blockchain finality....
// The params argument can be any object which maps a set of generic parameters into chain specific parameters defined in RelayConfig.
// It must encode as an object via [json.Marshal] and [github.com/fxamacker/cbor/v2.Marshal].
// Typically, would be either a struct with field names mapping to arguments, or anonymous map such as `map[string]any{"baz": 42, "test": true}}`
//
// returnVal must [json.Unmarshal] and and [github.com/fxamacker/cbor/v2.Marshal] as an object.
//
// Example use:
//
//	 type ProductParams struct {
//			ID int `json:"id"`
//	 }
//	 type ProductReturn struct {
//			Foo string `json:"foo"`
//			Bar *big.Int `json:"bar"`
//	 }
//	 func do(ctx context.Context, cr ContractReader) (resp ProductReturn, err error) {
//			err = cr.GetLatestValue(ctx, "FooContract", "GetProduct", primitives.Finalized, ProductParams{ID:1}, &resp)
//			return
//	 }
//
// Note that implementations should ignore extra fields in params that are not expected in the call to allow easier
// use across chains and contract versions.
// Similarly, when using a struct for returnVal, fields in the return value that are not on-chain will not be set.
func (ContractReader) GetLatestValue(ctx context.Context, readIdentifier string, confidenceLevel primitives.ConfidenceLevel, params, returnVal any) error {
	return UnimplementedError("ContractReader.GetLatestValue unimplemented")
}

// BatchGetLatestValues batches get latest value calls based on request, which is grouped by contract names that each have a slice of BatchRead.
// BatchGetLatestValuesRequest params and returnVal follow same rules as GetLatestValue params and returnVal arguments, with difference in how response is returned.
// BatchGetLatestValuesResult response is grouped by contract names, which contain read results that maintain the order from the request.
// Contract call errors are returned in the Err field of BatchGetLatestValuesResult.
func (ContractReader) BatchGetLatestValues(ctx context.Context, request BatchGetLatestValuesRequest) (BatchGetLatestValuesResult, error) {
	return nil, UnimplementedError("ContractReader.BatchGetLatestValues unimplemented")
}

// Bind will add provided bindings and will return an error if the contract is not known by the ContractReader, or if
// the Address is invalid. Any provided binding that already exists should result in a noop.
func (ContractReader) Bind(ctx context.Context, bindings []BoundContract) error {
	return UnimplementedError("ContractReader.Bind unimplemented")
}

// Unbind will remove all provided bindings.
func (ContractReader) Unbind(ctx context.Context, bindings []BoundContract) error {
	return UnimplementedError("ContractReader.Unbind unimplemented")
}

// QueryKey provides fetching chain agnostic events (Sequence) with general querying capability.
func (ContractReader) QueryKey(ctx context.Context, boundContract BoundContract, filter query.KeyFilter, limitAndSort query.LimitAndSort, sequenceDataType any) ([]Sequence, error) {
	return nil, UnimplementedError("ContractReader.QueryKey unimplemented")
}

// mustEmbedUnimplementedContractReaderServer surely serves some purpose, but I don't totally understand it.
func (ContractReader) mustEmbedUnimplementedContractReaderServer() {}

// Sneaky functions below are not part of the original ContractReader interface.

func (ContractReader) Start(context.Context) error {
	return UnimplementedError("ContractReader.Start unimplemented")
}

func (ContractReader) Close() error {
	return UnimplementedError("ContractReader.Close unimplemented")
}

func (ContractReader) HealthReport() map[string]error {
	panic(UnimplementedError("ContractReader.HealthReport unimplemented"))
}

func (ContractReader) Name() string {
	panic(UnimplementedError("ContractReader.Name unimplemented"))
}

func (ContractReader) Ready() error {
	return UnimplementedError("ContractReader.Ready unimplemented")
}

// BatchGetLatestValuesRequest string is contract name.
type BatchGetLatestValuesRequest map[BoundContract]ContractBatch
type ContractBatch []BatchRead
type BatchRead struct {
	ReadName  string
	Params    any
	ReturnVal any
}

type BatchGetLatestValuesResult map[BoundContract]ContractBatchResults
type ContractBatchResults []BatchReadResult
type BatchReadResult struct {
	ReadName    string
	returnValue any
	err         error
}

// GetResult returns an error if this specific read from the batch failed, otherwise returns the result in format that was provided in the request.
func (brr *BatchReadResult) GetResult() (any, error) {
	if brr.err != nil {
		return brr.returnValue, brr.err
	}

	return brr.returnValue, nil
}

func (brr *BatchReadResult) SetResult(returnValue any, err error) {
	brr.returnValue, brr.err = returnValue, err
}

type Head struct {
	Height string
	Hash   []byte
	// Timestamp is in Unix time
	Timestamp uint64
}

type Sequence struct {
	// This way we can retrieve past/future sequences (EVM log events) very granularly, but still hide the chain detail.
	Cursor string
	Head
	Data any
}

type BoundContract struct {
	Address string
	Name    string
}

func (bc BoundContract) ReadIdentifier(readName string) string {
	return bc.Address + "-" + bc.Name + "-" + readName
}

func (bc BoundContract) String() string {
	return bc.Address + "-" + bc.Name
}
