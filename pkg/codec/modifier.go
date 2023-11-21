package codec

import (
	"reflect"
)

type Modifier interface {
	RetypeForInput(outputType reflect.Type) (reflect.Type, error)

	// TransformInput transforms a type returned from AdjustForInput into the outputType.
	// You may also pass a pointer to the type returned by AdjustForInput to get a pointer to outputType.
	TransformInput(input any) (any, error)

	// TransformOutput is the reverse of transform input.
	// It is used to send back the object after it has been decoded
	TransformOutput(output any) (any, error)
}
