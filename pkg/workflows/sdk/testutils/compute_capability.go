package testutils

import (
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
)

type computeCapability struct {
	sdk      sdk.Runtime
	callback func(sdk sdk.Runtime, request capabilities.CapabilityRequest) capabilities.CapabilityResponse
}

func (c *computeCapability) Run(request capabilities.CapabilityRequest) capabilities.CapabilityResponse {
	return c.callback(c.sdk, request)
}

func (c *computeCapability) ID() string {
	return "__internal__custom_compute@1.0.0"
}

var _ CapabilityMock = &computeCapability{}
