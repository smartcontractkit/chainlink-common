package main

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basictrigger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/internal/rawsdk"
	"github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
)

func main() {
	req := rawsdk.GetRequest()
	requirements := &sdk.Requirements{Tee: &sdk.Tee{Type: &sdk.Tee_TypeSelection{TypeSelection: &sdk.TeeTypeSelection{Types: []*sdk.TeeTypeAndRegions{{Type: sdk.TeeType_TEE_TYPE_AWS_NITRO, Regions: []string{"us-west-2"}}}}}}}
	bytes, err := proto.Marshal(requirements)
	if err != nil {
		rawsdk.SendError(err)
	}
	rawsdk.Requirements(rawsdk.BufferToPointerLen(bytes))
	subscription := &sdk.TriggerSubscriptionRequest{
		Subscriptions: []*sdk.TriggerSubscription{
			{
				Id: "basic-test-trigger@1.0.0",
				Payload: rawsdk.Must(anypb.New(&basictrigger.Config{
					Name:   "first-trigger",
					Number: 100,
				})),
				Method: "Trigger",
			},
		},
	}
	switch req.GetRequest().(type) {
	case *sdk.ExecuteRequest_Subscribe:
		rawsdk.SendSubscription(subscription)
	}

	rawsdk.SendResponse(0)
}
