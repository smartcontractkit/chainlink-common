package sdk

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type Trigger[T proto.Message] interface {
	NewT() T
	Id() string
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
