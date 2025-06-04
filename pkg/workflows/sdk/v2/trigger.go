package sdk

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type baseTrigger[T proto.Message] interface {
	NewT() T
	CapabilityID() string
	ConfigAsAny() *anypb.Any
	Method() string
}

// Trigger represents a trigger in the workflow engine.
// Implementations should come from generated code.
// Methods are meant to be used by the Runner
type Trigger[T proto.Message] interface {
	baseTrigger[T]
	IsTrigger()
}
