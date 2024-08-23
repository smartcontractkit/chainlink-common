package testutils

import (
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows"
)

type computeCapability struct {
	sdk      workflows.Sdk
	callback func(sdk workflows.Sdk, request capabilities.CapabilityRequest) capabilities.CapabilityResponse
}

func (c *computeCapability) Run(request capabilities.CapabilityRequest) capabilities.CapabilityResponse {
	return c.callback(c.sdk, request)
}

func (c *computeCapability) ID() string {
	return "internal!!custom_compute@1.0.0"
}

var _ CapabilityMock = &computeCapability{}
