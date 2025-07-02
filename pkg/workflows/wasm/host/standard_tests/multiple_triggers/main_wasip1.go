package main

import (
	"fmt"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/actionandtrigger"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basictrigger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/internal/rawsdk"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func main() {
	switch request := rawsdk.GetRequest().Request.(type) {
	case *pb.ExecuteRequest_Subscribe:
		subscribe()
	case *pb.ExecuteRequest_Trigger:
		trigger(request)
	}
}

func subscribe() {
	subscription := &pb.TriggerSubscriptionRequest{
		Subscriptions: []*pb.TriggerSubscription{
			{
				Id: "basic-test-trigger@1.0.0",
				Payload: rawsdk.Must(anypb.New(&basictrigger.Config{
					Name:   "first-trigger",
					Number: 100,
				})),
				Method: "Trigger",
			},
			{
				Id: "basic-test-action-trigger@1.0.0",
				Payload: rawsdk.Must(anypb.New(&actionandtrigger.Config{
					Name:   "second-trigger",
					Number: 150,
				})),
				Method: "Trigger",
			},
			{
				Id: "basic-test-trigger@1.0.0",
				Payload: rawsdk.Must(anypb.New(&basictrigger.Config{
					Name:   "third-trigger",
					Number: 200,
				})),
				Method: "Trigger",
			},
		},
	}

	rawsdk.SendSubscription(subscription)
}

func trigger(request *pb.ExecuteRequest_Trigger) {
	switch request.Trigger.Id {
	case 0, 2:
		proveTrigger(request.Trigger, &basictrigger.Outputs{})
	case 1:
		proveTrigger(request.Trigger, &actionandtrigger.TriggerEvent{})
	default:
		panic("invalid trigger id")
	}
}

func proveTrigger(trigger *pb.Trigger, outputs interface {
	GetCoolOutput() string
	proto.Message
}) {
	if err := trigger.Payload.UnmarshalTo(outputs); err != nil {
		panic(err)
	}

	response := fmt.Sprintf("called %v with %v", trigger.Id, outputs.GetCoolOutput())
	rawsdk.SendResponse(response)
}
