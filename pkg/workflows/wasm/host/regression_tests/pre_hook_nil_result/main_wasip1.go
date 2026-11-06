package main

import (
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basictrigger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/internal/rawsdk"
	"github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
)

// The pre-hook exits without send_response, so the host hands back a nil result.
func main() {
	switch rawsdk.GetRequest().Request.(type) {
	case *sdk.ExecuteRequest_Subscribe:
		rawsdk.SendSubscription(&sdk.TriggerSubscriptionRequest{
			Subscriptions: []*sdk.TriggerSubscription{
				{
					Id:      "basic-test-trigger@1.0.0",
					Payload: rawsdk.Must(anypb.New(&basictrigger.Config{Name: "t", Number: 1})),
					Method:  "Trigger",
					PreHook: true,
				},
			},
		})
	case *sdk.ExecuteRequest_Trigger:
		rawsdk.SendResponse("ok")
	}
}
