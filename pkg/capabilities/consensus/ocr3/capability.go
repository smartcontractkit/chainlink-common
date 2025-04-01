package ocr3

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jonboulle/clockwork"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/requests"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/types"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/metering"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

const (
	ocrCapabilityID = "offchain_reporting@1.0.0"

	methodStartRequest = "start_request"
	methodSendResponse = "send_response"
	methodHeader       = "method"
	transmissionHeader = "transmission"
	terminateHeader    = "terminate"
)

var info = capabilities.MustNewCapabilityInfo(
	ocrCapabilityID,
	capabilities.CapabilityTypeConsensus,
	"OCR3 consensus exposed as a capability.",
)

type capability struct {
	services.Service
	eng *services.Engine

	capabilities.CapabilityInfo
	capabilities.Validator[config, inputs, requests.Response]

	reqHandler *requests.Handler

	requestTimeout     time.Duration
	requestTimeoutLock sync.RWMutex

	clock clockwork.Clock

	aggregatorFactory types.AggregatorFactory
	aggregators       map[string]types.Aggregator

	encoderFactory types.EncoderFactory
	encoders       map[string]types.Encoder

	callbackChannelBufferSize int

	registeredWorkflowsIDs map[string]bool
	mu                     sync.RWMutex
}

var _ CapabilityIface = (*capability)(nil)
var _ capabilities.ConsensusCapability = (*capability)(nil)

func NewCapability(s *requests.Store, clock clockwork.Clock, requestTimeout time.Duration, aggregatorFactory types.AggregatorFactory, encoderFactory types.EncoderFactory, lggr logger.Logger,
	callbackChannelBufferSize int) *capability {
	o := &capability{
		CapabilityInfo:    info,
		Validator:         capabilities.NewValidator[config, inputs, requests.Response](capabilities.ValidatorArgs{Info: info}),
		clock:             clock,
		requestTimeout:    requestTimeout,
		aggregatorFactory: aggregatorFactory,
		aggregators:       map[string]types.Aggregator{},
		encoderFactory:    encoderFactory,
		encoders:          map[string]types.Encoder{},

		callbackChannelBufferSize: callbackChannelBufferSize,
		registeredWorkflowsIDs:    map[string]bool{},
	}
	o.Service, o.eng = services.Config{
		Name: "OCR3CapabilityClient",
		NewSubServices: func(l logger.Logger) []services.Service {
			o.reqHandler = requests.NewHandler(lggr, s, clock, requestTimeout)
			return []services.Service{o.reqHandler}
		},
	}.NewServiceEngine(lggr)
	return o
}

func (o *capability) RegisterToWorkflow(ctx context.Context, request capabilities.RegisterToWorkflowRequest) error {
	c, err := o.ValidateConfig(request.Config)
	if err != nil {
		return err
	}

	o.mu.Lock()
	defer o.mu.Unlock()
	agg, err := o.aggregatorFactory(c.AggregationMethod, *c.AggregationConfig, o.eng)
	if err != nil {
		return err
	}
	o.aggregators[request.Metadata.WorkflowID] = agg

	encoder, err := o.encoderFactory(c.Encoder, c.EncoderConfig, o.eng)
	if err != nil {
		return err
	}
	o.encoders[request.Metadata.WorkflowID] = encoder
	o.registeredWorkflowsIDs[request.Metadata.WorkflowID] = true
	return nil
}

func (o *capability) GetAggregator(workflowID string) (types.Aggregator, error) {
	agg, ok := o.aggregators[workflowID]
	if !ok {
		return nil, fmt.Errorf("no aggregator found for workflowID %s", workflowID)
	}

	return agg, nil
}

func (o *capability) GetEncoderByWorkflowID(workflowID string) (types.Encoder, error) {
	enc, ok := o.encoders[workflowID]
	if !ok {
		return nil, fmt.Errorf("no encoder found for workflowID %s", workflowID)
	}

	return enc, nil
}

func (o *capability) GetEncoderByName(encoderName string, config *values.Map) (types.Encoder, error) {
	return o.encoderFactory(encoderName, config, o.eng)
}

func (o *capability) GetRegisteredWorkflowsIDs() []string {
	o.mu.RLock()
	defer o.mu.RUnlock()

	workflows := make([]string, 0, len(o.registeredWorkflowsIDs))
	for wf := range o.registeredWorkflowsIDs {
		workflows = append(workflows, wf)
	}
	return workflows
}

func (o *capability) UnregisterWorkflowID(workflowID string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	delete(o.registeredWorkflowsIDs, workflowID)
}

func (o *capability) UnregisterFromWorkflow(ctx context.Context, request capabilities.UnregisterFromWorkflowRequest) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	delete(o.registeredWorkflowsIDs, request.Metadata.WorkflowID)
	delete(o.aggregators, request.Metadata.WorkflowID)
	delete(o.encoders, request.Metadata.WorkflowID)
	return nil
}

func (o *capability) setRequestTimeout(timeout time.Duration) {
	o.requestTimeoutLock.Lock()
	defer o.requestTimeoutLock.Unlock()
	o.requestTimeout = timeout
}

// Execute enqueues a new consensus request, passing it to the reporting plugin as needed.
// IMPORTANT: OCR3 only exposes signatures via the contractTransmitter, which is located
// in a separate process to the reporting plugin LOOPP. However, only the reporting plugin
// LOOPP is able to transmit responses back to the workflow engine. As a workaround to this, we've implemented a custom contract transmitter which fetches this capability from the
// registry and calls Execute with the response, setting "method = `methodSendResponse`".
func (o *capability) Execute(ctx context.Context, r capabilities.CapabilityRequest) (capabilities.CapabilityResponse, error) {
	m := struct {
		Method       string
		Transmission map[string]any
		Terminate    bool
	}{
		Method: methodStartRequest,
	}
	err := r.Inputs.UnwrapTo(&m)
	if err != nil {
		o.eng.Warnf("could not unwrap method from CapabilityRequest, using default: %v", err)
	}

	switch m.Method {
	case methodSendResponse:
		inputs, err := values.NewMap(m.Transmission)
		if err != nil {
			return capabilities.CapabilityResponse{}, fmt.Errorf("failed to create map for response inputs: %w", err)
		}
		o.eng.Debugw("Execute - sending response", "workflowExecutionID", r.Metadata.WorkflowExecutionID, "inputs", inputs, "terminate", m.Terminate)
		var responseErr error
		if m.Terminate {
			o.eng.Debugw("Execute - terminating execution", "workflowExecutionID", r.Metadata.WorkflowExecutionID)
			responseErr = capabilities.ErrStopExecution
		}
		out := requests.Response{
			WorkflowExecutionID: r.Metadata.WorkflowExecutionID,
			Value:               inputs,
			Err:                 responseErr,
		}
		o.reqHandler.SendResponse(ctx, out)

		// Return a dummy response back to the caller
		// This allows the transmitter to block on a response before
		// returning from Transmit()
		return capabilities.CapabilityResponse{}, nil
	case methodStartRequest:
		// Receives and stores an observation to do consensus on
		// Receives an aggregation method; at this point the method has been validated
		// Returns the consensus result over a channel
		inputs, err := o.ValidateInputs(r.Inputs)
		if err != nil {
			return capabilities.CapabilityResponse{}, err
		}
		inputLenBytes := byteSizeOfMap(r.Inputs)

		config, err := o.ValidateConfig(r.Config)
		if err != nil {
			return capabilities.CapabilityResponse{}, err
		}

		ch, err := o.queueRequestForProcessing(ctx, r.Metadata, inputs, config)
		if err != nil {
			return capabilities.CapabilityResponse{}, err
		}

		response := <-ch
		outputLenBytes := byteSizeOfMap(response.Value)
		return capabilities.CapabilityResponse{
			Value: response.Value,
			Metadata: capabilities.ResponseMetadata{
				Metering: []capabilities.MeteringNodeDetail{
					{SpendUnit: metering.PayloadUnit.Name, SpendValue: fmt.Sprintf("%d", inputLenBytes+outputLenBytes)},
				},
			},
		}, response.Err
	}

	return capabilities.CapabilityResponse{}, fmt.Errorf("unknown method: %s", m.Method)
}

// queueRequestForProcessing queues a request for processing by the worker
// goroutine by adding the request to its store.
//
// When a request is queued, a timer is started to ensure that the request does not exceed its expiry time.
func (o *capability) queueRequestForProcessing(
	ctx context.Context,
	metadata capabilities.RequestMetadata,
	i *inputs,
	c *config,
) (<-chan requests.Response, error) {
	callbackCh := make(chan requests.Response, o.callbackChannelBufferSize)

	// Use the capability-level request timeout unless the request's config specifies
	// its own timeout, in which case we'll use that instead. This allows the workflow spec
	// to configure more granular timeouts depending on the circumstances.
	o.requestTimeoutLock.RLock()
	requestTimeout := o.requestTimeout
	if c.RequestTimeoutMS != 0 {
		requestTimeout = time.Duration(c.RequestTimeoutMS) * time.Millisecond
	}
	o.requestTimeoutLock.RUnlock()

	r := &requests.Request{
		StopCh:                   make(chan struct{}),
		CallbackCh:               callbackCh,
		WorkflowExecutionID:      metadata.WorkflowExecutionID,
		WorkflowID:               metadata.WorkflowID,
		WorkflowOwner:            metadata.WorkflowOwner,
		WorkflowName:             metadata.WorkflowName,
		ReportID:                 c.ReportID,
		WorkflowDonID:            metadata.WorkflowDonID,
		WorkflowDonConfigVersion: metadata.WorkflowDonConfigVersion,
		Observations:             i.Observations,
		OverriddenEncoderName:    i.EncoderName,
		OverriddenEncoderConfig:  i.EncoderConfig,
		KeyID:                    c.KeyID,
		ExpiresAt:                o.clock.Now().Add(requestTimeout),
	}

	o.eng.Debugw("Execute - adding to store", "workflowID", r.WorkflowID, "workflowExecutionID", r.WorkflowExecutionID, "observations", r.Observations)

	o.reqHandler.SendRequest(ctx, r)
	return callbackCh, nil
}

// byteSizeOfMap is a utility to get the wire-size
// of a values.Map.
func byteSizeOfMap(m *values.Map) int {
	if m == nil {
		return 0
	}
	pbVal := values.Proto(m)
	size := proto.Size(pbVal)
	return size
}
