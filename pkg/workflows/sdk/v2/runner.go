package sdk

import (
	"io"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type RunnerBase interface {
	LogWriter() io.Writer
	Config() []byte
}

type DonRunner interface {
	RunnerBase

	Run(args *WorkflowArgs[DonRuntime])
}

// NOTE that some triggers in node mode need callbacks into the WASM when setting up (like WebSocket trigger).
// This is not in this interface yet.

type NodeRunner interface {
	RunnerBase

	Run(args *WorkflowArgs[NodeRuntime])
}

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
