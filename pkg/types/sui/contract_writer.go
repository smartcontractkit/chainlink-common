package sui

import (
	"errors"
	"fmt"
)

type SuiPTBCommandType string

const (
	SuiPTBCommandMoveCall SuiPTBCommandType = "move_call"
	SuiPTBCommandPublish  SuiPTBCommandType = "publish"
	SuiPTBCommandTransfer SuiPTBCommandType = "transfer"
)

type PTBCommandDependency struct {
	CommandIndex uint16
	ResultIndex  *uint16
}

// PointerTag defines the structured format for pointer tags used in chain reader.
// Pointer tags specify how to derive object IDs from pointer objects stored on-chain.
type PointerTag struct {
	// Module name containing the pointer object (e.g. "state_object", "offramp", "counter")
	Module string `json:"module"`
	// PointerName is the object type to search for (e.g. "CCIPObjectRefPointer", "OffRampStatePointer")
	PointerName string `json:"pointerName"`
	// FieldName is OPTIONAL and NOT USED by the implementation. The parent field name is automatically
	// looked up from the global common.PointerConfigs registry based on the PointerName.
	// This field exists for backward compatibility or future implementations to override static code but is currently ignored.
	FieldName string `json:"fieldName,omitempty"`
	// DerivationKey is the key used to derive the child object ID from the parent object ID (e.g. "CCIPObjectRef", "CCIP_OWNABLE")
	DerivationKey string `json:"derivationKey"`
	// PackageID is the package ID for the Pointer object if it differs from the calling contract's package ID
	// This is used for cross-package pointer dependencies (e.g. offramp package depending on CCIP package CCIPObjectRef)
	// If empty, the calling contract's package ID is used
	PackageID string `json:"packageId,omitempty"`
}

func (p PointerTag) Validate() error {
	if p.Module == "" {
		return errors.New("PointerTag.Module is required")
	}
	if p.PointerName == "" {
		return errors.New("PointerTag.PointerName is required")
	}
	// FieldName is optional - it's looked up from common.PointerConfigs
	if p.DerivationKey == "" {
		return errors.New("PointerTag.DerivationKey is required")
	}
	return nil
}

// SuiFunctionParam defines a parameter for a Sui function call
type SuiFunctionParam struct {
	// Name of the parameter
	Name string
	// PointerTag (optional) specify how to derive object IDs from pointer objects stored on-chain.
	PointerTag *PointerTag
	// Type of the parameter (e.g., "u64", "String", "vector<u8>", "ptb_dependency")
	Type string
	// IsMutable specifies if the object is mutable or not (optional - defaults to true)
	IsMutable *bool
	// IsGeneric specifies if the parameter is a generic argument
	GenericType *string
	// Whether the parameter is required
	Required bool
	// Default value to use if not provided
	DefaultValue any
	// Result from a previous PTB Command (optional). It is used for expressive construction of PTB commands
	PTBDependency *PTBCommandDependency
	// GenericDependency maps to internal helpers for fetching an unknown generic type required by the parameter
	GenericDependency *string
}

var (
	PTBChainWriterModuleName = "cll://component=cw/type=ptb_builder"
	CCIPExecute              = "execute"
	CCIPCommit               = "commit"
)

type ChainWriterConfig struct {
	Modules map[string]*ChainWriterModule
}

type ChainWriterModule struct {
	// The module name (optional). When not provided, the key in the map under which this module
	// is stored is used.
	Name      string
	ModuleID  string
	Functions map[string]*ChainWriterFunction
}

type ChainWriterPTBCommand struct {
	Type SuiPTBCommandType
	// The package ID to call (optional). This may not be needed in the case
	// that the type of PTB command does not require it (e.g. Publish).
	PackageId *string            `json:"package_id,omitempty"`
	ModuleId  *string            `json:"module_id,omitempty"`
	Function  *string            `json:"function,omitempty"`
	TypeArgs  []string           `json:"type_args,omitempty"`
	Params    []SuiFunctionParam `json:"params,omitempty"`
}

// GetParamKey returns the key for a parameter in the PTB command in a map of arguments.
// The key is a string that uniquely identifies the parameter within the map of arguments.
// The key is formatted as follows:
// "packageId::moduleId::functionName::parameterName"
// This format allows associating specific argument values with their target
// Move function call and parameter name within a potentially complex PTB.
func (c ChainWriterPTBCommand) GetParamKey(paramName string) string {
	return fmt.Sprintf("%s.%s.%s.%s", *c.PackageId, *c.ModuleId, *c.Function, paramName)
}

// PrerequisiteObject represents a structure defining requirements or dependencies needed before constructing the PTB.
// These requirements refer to object details that need to be fetched with `SuiX_GetOwnedObjects` and then populated
// into the arguments map provided for PTB construction.
//
// The usage flow is that a request is made to get all the owned objects by "OwnerId" and then picking the one
// that matches the Tag
type PrerequisiteObject struct {
	OwnerId *string
	Name    string // the key under which the value is inserted in the args, must match one of the arg names used in the PTB commands
	Tag     string
	SetKeys bool // optionally set the keys of the object in the arg map instead of name
}

type ChainWriterFunction struct {
	// The function name (optional). When not provided, the key in the map under which this function
	// is stored is used.
	Name string
	// The public key of the account that will sign and submit the transaction.
	PublicKey []byte
	// The values that need to be loaded into the args by making SuiX_GetOwnedObjects calls
	PrerequisiteObjects []PrerequisiteObject
	// Mapping of logical names to package/module addresses
	// e.g. ccip_package -> 0x123
	// 		ccip_offramp -> 0x321
	AddressMappings map[string]string

	Params []SuiFunctionParam
	// The set of PTB commands to run as part of this function call.
	// This field is used in replacement of `Params` above.
	PTBCommands []ChainWriterPTBCommand
}

type Arguments struct {
	Args     map[string]any
	ArgTypes map[string]string // Maps argument name to its generic type
}

type ChainWriterSignal struct{}
