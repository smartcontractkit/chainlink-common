package codec

import (
	"reflect"
)

type ElementExtractorLocation string

const (
	FirstElementLocation  = ElementExtractorLocation(FirstElementTransform)
	MiddleElementLocation = ElementExtractorLocation(MiddleElementTransform)
	LastElementLocation   = ElementExtractorLocation(LastElementTransform)
)

type ElementExtractor struct {
	FieldInfo         map[string]ElementExtractorLocation
	outputToInputType map[reflect.Type]reflect.Type
	inputToOutputType map[reflect.Type]reflect.Type
}

func (e *ElementExtractor) RetypeForInput(outputType reflect.Type) (reflect.Type, error) {
	//TODO implement me
	panic("implement me")
}

func (e *ElementExtractor) TransformInput(input any) (any, error) {
	//TODO implement me
	panic("implement me")
}

func (e *ElementExtractor) TransformOutput(output any) (any, error) {
	//TODO implement me
	panic("implement me")
}
