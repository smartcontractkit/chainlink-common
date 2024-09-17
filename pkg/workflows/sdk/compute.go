package sdk

//go:generate go run ./gen

type ComputeOutput[T any] struct {
	Value T
}

type ComputeOutputCap[T any] interface {
	CapDefinition[ComputeOutput[T]]
	Value() CapDefinition[T]
}

type computeOutputCap[T any] struct {
	CapDefinition[ComputeOutput[T]]
}

func (c *computeOutputCap[T]) Value() CapDefinition[T] {
	return AccessField[ComputeOutput[T], T](c.CapDefinition, "Value")
}

var _ ComputeOutputCap[struct{}] = &computeOutputCap[struct{}]{}
