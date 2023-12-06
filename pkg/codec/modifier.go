package codec

import (
	"reflect"
)

type Modifier interface {
	RetypeForOffChain(onChainType reflect.Type, itemType string) (reflect.Type, error)

	// TransformForOnChain transforms a type returned from AdjustForInput into the outputType.
	// You may also pass a pointer to the type returned by AdjustForInput to get a pointer to outputType.
	TransformForOnChain(offChainValue any, itemType string) (any, error)

	// TransformForOffChain is the reverse of TransformForOnChain input.
	// It is used to send back the object after it has been decoded
	TransformForOffChain(onChainValue any, itemType string) (any, error)
}
