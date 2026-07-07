package main

import (
	"fmt"

	"google.golang.org/protobuf/types/known/anypb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basicaction"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basictrigger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/internal/rawsdk"
	"github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
)

func main() {
	request := rawsdk.GetRequest()
	switch request.Request.(type) {
	case *sdk.ExecuteRequest_Trigger:
		input := &basicaction.Inputs{InputThing: true}
		err := rawsdk.DoRequestErr("basic-test-action@1.0.0", "PerformAction", sdk.Mode_MODE_DON, input)
		if err != nil {
			rawsdk.SendError(err)
		}
		rawsdk.SendResponse("should have errored out...")
	case *sdk.ExecuteRequest_PreHook:
		rawsdk.SendRestrictions(&sdk.Restrictions{
			Capabilities: &sdk.CapabilityRestrictions{MaxTotalCalls: 0},
		})
	case *sdk.ExecuteRequest_Subscribe:
		rawsdk.SendSubscription(&sdk.TriggerSubscriptionRequest{
			Subscriptions: []*sdk.TriggerSubscription{
				{
					Id: "basic-test-trigger@1.0.0",
					Payload: rawsdk.Must(anypb.New(&basictrigger.Config{
						Name:   "first-trigger",
						Number: 100,
					})),
					Method:  "Trigger",
					PreHook: true,
				},
			},
		})
	default:
		rawsdk.SendError(fmt.Errorf("unexpected request type: %T", request.Request))
	}
}
