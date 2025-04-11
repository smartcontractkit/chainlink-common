// Code generated by github.com/smartcontractkit/chainlink-common/pkg/capabilities/protoc, DO NOT EDIT.

package server

import (
	"context"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/protoc/pkg/testdata/fixtures/capabilities/basicaction"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/testutils"

	"github.com/stretchr/testify/mock"
)

type BasicActionCapability struct {
	m mock.Mock
}

// TODO register?
func OnBasicAction(func(ctx context.Context, input *basicaction.Inputs /* TODO this isn't right */, config *basicaction.Inputs) (*basicaction.Outputs, error)) testutils.MockCapabilityCall[basicaction.Outputs] {
	return &testutils.MockCapabilityCall[basicaction.Outputs]{
		Call: c.m.On("OnPerformAction", mock.AnythingOfType(ctx), mock.AnythingOfType(input)),
	}
}

// TODO register if needed...
