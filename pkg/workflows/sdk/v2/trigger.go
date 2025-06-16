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

type Trigger[M proto.Message, T any] interface {
	baseTrigger[M]
	Adapt(m M) (T, error)
}
