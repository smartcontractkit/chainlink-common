package aptos

import (
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

type ContractReaderConfig struct {
	IsLoopPlugin bool
	Modules      map[string]*ContractReaderModule
}

type ContractReaderModule struct {
	// The module name (optional). When not provided, the key in the map under which this module
	// is stored is used.
	Name      string
	Functions map[string]*ContractReaderFunction
	Events    map[string]*ContractReaderEvent
}

type ContractReaderFunction struct {
	// The function name (optional). When not provided, the key in the map under which this function
	// is stored is used.
	Name   string
	Params []FunctionParam

	ResultFieldRenames  map[string]RenamedField
	ResultTupleToStruct []string
	ResultUnwrapStruct  []string
}

type FunctionParam struct {
	// The function parameter name.
	Name string
	// The function parameter Move type.
	Type string
	// True if this is a required parameter, false otherwise.
	Required bool
	// If this is not a required parameter and it is not provided, this default value will be used.
	DefaultValue any
}

type ContractReaderEvent struct {
	// The struct where the event handle is defined.
	EventHandleStructName string

	// The name of the event handle field.
	// This field can be defined as path to the nested
	// struct that stores the event, e.g. "token_pool_state.burned_events"
	EventHandleFieldName string

	// The event account address.
	// This field can be defined in several ways:
	// - Empty string, which means the event account address is the address of the bound contract.
	// - An exact address hex string (eg. 0x1234 or 1234) containing the events.
	// - A fully qualified function name (eg. 0x1234::my_contract::get_event_address) which
	//   takes no parameters and returns the actual event account address.
	// - A name containing the module name and function name components
	//   (eg. my_first_contract::get_event_address) stored at the address of the bound contract,
	//   which takes no parameters and returns the actual event account address.
	EventAccountAddress string

	// Renames of event field names (optional). When not provided, the field names are used as-is.
	EventFieldRenames map[string]RenamedField

	// Renames provided filters to match the event field names (optional). When not provided, the filters are used as-is.
	EventFilterRenames map[string]string
}

type RenamedField struct {
	// The new field name (optional). This does not need to be provided if this field does not need
	// to be renamed.
	NewName string

	// Rename sub-fields. This assumes that the event field value is a struct or a map with string keys.
	SubFieldRenames map[string]RenamedField
}

type SequenceWithMetadata struct {
	Sequence  types.Sequence
	TxVersion uint64
	TxHash    string
}
