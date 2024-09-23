package ocr3

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jonboulle/clockwork"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/beholder/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/requests"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/types"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
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
	services.StateMachine
	capabilities.CapabilityInfo
	capabilities.Validator[config, inputs, requests.Response]

	reqHandler *requests.Handler
	stopCh     services.StopChan
	wg         sync.WaitGroup
	lggr       logger.Logger

	requestTimeout time.Duration
	clock          clockwork.Clock

	aggregatorFactory types.AggregatorFactory
	aggregators       map[string]types.Aggregator

	encoderFactory types.EncoderFactory
	encoders       map[string]types.Encoder

	callbackChannelBufferSize int

	registeredWorkflowsIDs map[string]bool
	mu                     sync.RWMutex
}

var _ capabilityIface = (*capability)(nil)
var _ capabilities.ConsensusCapability = (*capability)(nil)

func newCapability(s *requests.Store, clock clockwork.Clock, requestTimeout time.Duration, aggregatorFactory types.AggregatorFactory, encoderFactory types.EncoderFactory, lggr logger.Logger,
	callbackChannelBufferSize int) *capability {
	o := &capability{
		CapabilityInfo:    info,
		Validator:         capabilities.NewValidator[config, inputs, requests.Response](capabilities.ValidatorArgs{Info: info}),
		reqHandler:        requests.NewHandler(lggr, s, clock, requestTimeout),
		clock:             clock,
		requestTimeout:    requestTimeout,
		stopCh:            make(chan struct{}),
		lggr:              logger.Named(lggr, "OCR3CapabilityClient"),
		aggregatorFactory: aggregatorFactory,
		aggregators:       map[string]types.Aggregator{},
		encoderFactory:    encoderFactory,
		encoders:          map[string]types.Encoder{},

		callbackChannelBufferSize: callbackChannelBufferSize,
		registeredWorkflowsIDs:    map[string]bool{},
	}
	return o
}

func (o *capability) Start(ctx context.Context) error {
	return o.StartOnce("OCR3Capability", func() error {
		err := o.reqHandler.Start(ctx)
		if err != nil {
			return fmt.Errorf("failed to start request handler: %w", err)
		}

		return nil
	})
}

func (o *capability) Close() error {
	return o.StopOnce("OCR3Capability", func() error {
		close(o.stopCh)
		o.wg.Wait()
		err := o.reqHandler.Close()
		if err != nil {
			return fmt.Errorf("failed to close request handler: %w", err)
		}

		return nil
	})
}

func (o *capability) Name() string { return o.lggr.Name() }

func (o *capability) HealthReport() map[string]error {
	return map[string]error{o.Name(): o.Healthy()}
}

func (o *capability) RegisterToWorkflow(ctx context.Context, request capabilities.RegisterToWorkflowRequest) error {
	c, err := o.ValidateConfig(request.Config)
	if err != nil {
		return err
	}

	o.mu.Lock()
	defer o.mu.Unlock()
	agg, err := o.aggregatorFactory(c.AggregationMethod, *c.AggregationConfig, o.lggr)
	if err != nil {
		return err
	}
	o.aggregators[request.Metadata.WorkflowID] = agg

	encoder, err := o.encoderFactory(c.EncoderConfig)
	if err != nil {
		return err
	}
	o.encoders[request.Metadata.WorkflowID] = encoder
	o.registeredWorkflowsIDs[request.Metadata.WorkflowID] = true
	return nil
}

func (o *capability) getAggregator(workflowID string) (types.Aggregator, error) {
	agg, ok := o.aggregators[workflowID]
	if !ok {
		return nil, fmt.Errorf("no aggregator found for workflowID %s", workflowID)
	}

	return agg, nil
}

func (o *capability) getEncoder(workflowID string) (types.Encoder, error) {
	enc, ok := o.encoders[workflowID]
	if !ok {
		return nil, fmt.Errorf("no aggregator found for workflowID %s", workflowID)
	}

	return enc, nil
}

func (o *capability) getRegisteredWorkflowsIDs() []string {
	o.mu.RLock()
	defer o.mu.RUnlock()

	workflows := make([]string, 0, len(o.registeredWorkflowsIDs))
	for wf := range o.registeredWorkflowsIDs {
		workflows = append(workflows, wf)
	}
	return workflows
}

func (o *capability) unregisterWorkflowID(workflowID string) {
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

// Execute enqueues a new consensus request, passing it to the reporting plugin as needed.
// IMPORTANT: OCR3 only exposes signatures via the contractTransmitter, which is located
// in a separate process to the reporting plugin LOOPP. However, only the reporting plugin
// LOOPP is able to transmit responses back to the workflow engine. As a workaround to this, we've implemented a custom contract transmitter which fetches this capability from the
// registry and calls Execute with the response, setting "method = `methodSendResponse`".
func (o *capability) Execute(ctx context.Context, r capabilities.CapabilityRequest) (capabilities.CapabilityResponse, error) {
	payload := &pb.Event{
		Component: "OCR3 capability",
		Event:     "Execute",
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
	}
	payloadBytes, err := proto.Marshal(payload)
	if err != nil {
		return capabilities.CapabilityResponse{}, err
	}

	beholder.GetEmitter().Emit(context.Background(), payloadBytes,
		"beholder_data_schema", "/custom-message/versions/1", // required
		"beholder_data_type", "custom_message",
		"package_name", "capabilities_test",
	)

	m := struct {
		Method       string
		Transmission map[string]any
		Terminate    bool
	}{
		Method: methodStartRequest,
	}
	err = r.Inputs.UnwrapTo(&m)
	if err != nil {
		o.lggr.Warnf("could not unwrap method from CapabilityRequest, using default: %v", err)
	}

	switch m.Method {
	case methodSendResponse:
		inputs, err := values.NewMap(m.Transmission)
		if err != nil {
			return capabilities.CapabilityResponse{}, fmt.Errorf("failed to create map for response inputs: %w", err)
		}
		o.lggr.Debugw("Execute - sending response", "workflowExecutionID", r.Metadata.WorkflowExecutionID, "inputs", inputs, "terminate", m.Terminate)
		var responseErr error
		if m.Terminate {
			o.lggr.Debugw("Execute - terminating execution", "workflowExecutionID", r.Metadata.WorkflowExecutionID)
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

		config, err := o.ValidateConfig(r.Config)
		if err != nil {
			return capabilities.CapabilityResponse{}, err
		}

		ch, err := o.queueRequestForProcessing(ctx, r.Metadata, inputs, config)
		if err != nil {
			return capabilities.CapabilityResponse{}, err
		}

		response := <-ch
		return capabilities.CapabilityResponse{
			Value: response.Value,
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
	requestTimeout := o.requestTimeout
	if c.RequestTimeoutMS != 0 {
		requestTimeout = time.Duration(c.RequestTimeoutMS) * time.Millisecond
	}

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
		KeyID:                    c.KeyID,
		ExpiresAt:                o.clock.Now().Add(requestTimeout),
	}

	o.lggr.Debugw("Execute - adding to store", "workflowID", r.WorkflowID, "workflowExecutionID", r.WorkflowExecutionID, "observations", r.Observations)

	o.reqHandler.SendRequest(ctx, r)
	return callbackCh, nil
}
