package v2

import (
	"context"
	"encoding/json"

	"github.com/invopop/jsonschema"
	jsonvalidate "github.com/santhosh-tekuri/jsonschema/v5"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

type CapabilityResponse[O any] struct {
	Value O
	Err   error
}

type CapabilityRequest[I, C any] struct {
	Metadata capabilities.RequestMetadata
	Config   C
	Inputs   I
}

type RegisterToWorkflowRequest[C any] struct {
	Metadata capabilities.RegistrationMetadata
	Config   C
}

type UnregisterFromWorkflowRequest[C any] struct {
	Metadata capabilities.RegistrationMetadata
	Config   C
}

type Capability[I, O, C any] interface {
	RegisterToWorkflow(ctx context.Context, req RegisterToWorkflowRequest[C]) error
	UnregisterFromWorkflow(ctx context.Context, req UnregisterFromWorkflowRequest[C]) error
	Execute(ctx context.Context, callback chan<- CapabilityResponse[O], request CapabilityRequest[I, C]) error
}

type capability[I, O, C any] struct {
	inner Capability[I, O, C]
}

func (c *capability[I, O, C]) RegisterToWorkflow(ctx context.Context, req capabilities.RegisterToWorkflowRequest) error {
	var conf C
	err := c.validate(&conf, req.Config)
	if err != nil {
		return err
	}

	err = req.Config.UnwrapTo(&conf)
	if err != nil {
		return err
	}
	regReq := RegisterToWorkflowRequest[C]{
		Metadata: req.Metadata,
		Config:   conf,
	}
	return c.inner.RegisterToWorkflow(ctx, regReq)
}

func (c *capability[I, O, C]) UnregisterFromWorkflow(ctx context.Context, req capabilities.UnregisterFromWorkflowRequest) error {
	var conf C
	err := c.validate(&conf, req.Config)
	if err != nil {
		return err
	}

	err = req.Config.UnwrapTo(&conf)
	if err != nil {
		return err
	}
	regReq := UnregisterFromWorkflowRequest[C]{
		Metadata: req.Metadata,
		Config:   conf,
	}
	return c.inner.UnregisterFromWorkflow(ctx, regReq)
}

func (c *capability[I, O, C]) validate(str any, m *values.Map) error {
	sch := jsonschema.Reflect(str)
	schemab, err := json.Marshal(sch)
	if err != nil {
		return err
	}

	mapping, err := values.Unwrap(m)
	if err != nil {
		return err
	}

	schema, err := jsonvalidate.CompileString("<uriPrefix>", string(schemab))
	if err != nil {
		return err
	}
	return schema.Validate(mapping)
}

func (c *capability[I, O, C]) Execute(ctx context.Context, callback chan<- capabilities.CapabilityResponse, request capabilities.CapabilityRequest) error {
	tcb := make(chan CapabilityResponse[O])
	go c.forwardResponses(ctx, callback, tcb)

	var conf C
	err := c.validate(&conf, request.Config)
	if err != nil {
		return err
	}

	err = request.Config.UnwrapTo(&conf)
	if err != nil {
		return err
	}

	var inp I
	err = c.validate(&inp, request.Inputs)
	if err != nil {
		return err
	}

	err = request.Inputs.UnwrapTo(&inp)
	if err != nil {
		return err
	}

	treq := CapabilityRequest[I, C]{
		Metadata: request.Metadata,
		Config:   conf,
		Inputs:   inp,
	}

	return c.inner.Execute(ctx, tcb, treq)
}

func (c *capability[I, O, C]) forwardResponses(ctx context.Context, callback chan<- capabilities.CapabilityResponse, typedCallback chan CapabilityResponse[O]) {
	for {
		select {
		case <-ctx.Done():
			return
		case resp, isOpen := <-typedCallback:
			if !isOpen {
				close(callback)
				return
			}

			v, err := values.Wrap(resp.Value)
			if err != nil {
				callback <- capabilities.CapabilityResponse{
					Err: err,
				}
			}

			callback <- capabilities.CapabilityResponse{
				Value: v,
			}
		}
	}
}

func NewCapability[I, O, C any](cap Capability[I, O, C]) capabilities.CallbackExecutable {
	return &capability[I, O, C]{
		inner: cap,
	}
}
