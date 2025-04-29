package wasm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/actionandtrigger"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basictrigger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/v2/pb"
)

const anyExecutionId = "execId"

var anyConfig = []byte("config")
var anyMaxResponseSize = uint64(2048)

var subscribeRequest = &pb.ExecuteRequest{
	Id:              anyExecutionId,
	Config:          anyConfig,
	MaxResponseSize: anyMaxResponseSize,
	Request:         &pb.ExecuteRequest_Subscribe{Subscribe: &emptypb.Empty{}},
}

const anyTriggerId = "triggerId"

var anyTrigger = &basictrigger.Outputs{CoolOutput: "tirggerOutput"}
var anyExecuteRequest = &pb.ExecuteRequest{
	Id:              anyExecutionId,
	Config:          anyConfig,
	MaxResponseSize: anyMaxResponseSize,
	Request: &pb.ExecuteRequest_Trigger{
		Trigger: &pb.Trigger{
			Id:      anyTriggerId,
			Payload: mustAny(anyTrigger),
		},
	},
}

func TestRunner_Config(t *testing.T) {
	r := runner[sdk.DonRuntime]{}
	r.config = []byte("test")
	assert.Equal(t, "test", string(r.config))
	s := subscriber[sdk.DonRuntime]{}
	s.config = []byte("test")
	assert.Equal(t, "test", string(r.config))
}

func TestRunner_LogWriter(t *testing.T) {
	r := runner[sdk.DonRuntime]{}
	assert.IsType(t, &writer{}, r.LogWriter())
	s := subscriber[sdk.DonRuntime]{}
	assert.IsType(t, &writer{}, s.LogWriter())
}

func TestRunner_Run(t *testing.T) {
	runWorkflow := func(runner sdk.DonRunner) {

	}

	t.Run("gathers subscriptions", func(t *testing.T) {
		assert.Fail(t, "not written yet")
	})

	t.Run("makes callback with correct runner", func(t *testing.T) {
		assert.Fail(t, "not written yet")
	})
}

func mustAny(msg proto.Message) *anypb.Any {
	a, err := anypb.New(msg)
	if err != nil {
		panic(err)
	}
	return a
}
