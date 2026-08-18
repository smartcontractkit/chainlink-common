package codec

import (
	"reflect"
)

// Modifier allows you to modify the off-chain type to be used on-chain, and vice-versa.
// A modifier is set up by retyping the on-chain type to a type used off-chain.
type Modifier interface {
	// RetypeToOffChain will retype the onChainType to its correlated offChainType. The itemType should be empty for an
	// expected whole struct. A dot-separated string can be provided when path traversal is supported on the modifier
	// to retype a nested field.
	//
	// For most modifiers, RetypeToOffChain must be called first with the entire struct to be retyped/modified before
	// any other transformations or path traversal can function.
	RetypeToOffChain(onChainType reflect.Type, itemType string) (reflect.Type, error)

	// TransformToOnChain transforms a type returned from AdjustForInput into the outputType.
	// You may also pass a pointer to the type returned by AdjustForInput to get a pointer to outputType.
	//
	// Modifiers should also optionally provide support for path traversal using itemType. In the case of using path
	// traversal, the offChainValue should be the field value being modified as identified by itemType.
	TransformToOnChain(offChainValue any, itemType string) (any, error)

	// TransformToOffChain is the reverse of TransformForOnChain input.
	// It is used to send back the object after it has been decoded
	//
	// Modifiers should also optionally provide support for path traversal using itemType. In the case of using path
	// traversal, the onChainValue should be the field value being modified as identified by itemType.
	TransformToOffChain(onChainValue any, itemType string) (any, error)
}
