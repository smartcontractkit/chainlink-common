package sdk

import "github.com/smartcontractkit/chainlink-common/pkg/capabilities"

type RuntimeV2 interface {
	CallCapability(call CapabilityCallPromise) (int32, error)
	AwaitCapabilities(calls ...CapabilityCallPromise) error
}

// weakly-typed, for the runtime to fulfill
type CapabilityCallPromise interface {
	CallInfo() (ref int32, capId string, request capabilities.CapabilityRequest)
	Fulfill(response capabilities.CapabilityResponse, err error)
}
