package codec

import "reflect"

type Modifier interface {
	AdjustForInput(outputType reflect.Type) (reflect.Type, error)

	// TransformInput transforms a type returned from AdjustForInput into the outputType.
	// You may also pass a pointer to the type returned by AdjustForInput to get a pointer to outputType.
	// This function is allowed to return a value that points the same address as input for efficiency.
	TransformInput(input any) (any, error)
}
