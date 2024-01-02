package codec

import (
	"reflect"
)

// Modifier allows you to modify the off-chain type to be used on-chain, and vice-versa.
// A modifier is set up by retyping the on-chain type to a type used off-chain.
type Modifier interface {
	RetypeForOffChain(onChainType reflect.Type, itemType string) (reflect.Type, error)

	// TransformForOnChain transforms a type returned from AdjustForInput into the outputType.
	// You may also pass a pointer to the type returned by AdjustForInput to get a pointer to outputType.
	TransformForOnChain(offChainValue any, itemType string) (any, error)

	// TransformForOffChain is the reverse of TransformForOnChain input.
	// It is used to send back the object after it has been decoded
	TransformForOffChain(onChainValue any, itemType string) (any, error)
}
