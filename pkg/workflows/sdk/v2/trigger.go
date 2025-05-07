package sdk

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// Trigger represents a trigger in the workflow engine.
// Implementations should come from generated code.
// Methods are meant to be used by the DonRunner or NodeRunner.
type Trigger[T proto.Message] interface {
	NewT() T
	CapabilityID() string
	ConfigAsAny() *anypb.Any
	Method() string
}

type DonTrigger[T proto.Message] interface {
	Trigger[T]
	IsDonTrigger()
}

type NodeTrigger[T proto.Message] interface {
	Trigger[T]
}
