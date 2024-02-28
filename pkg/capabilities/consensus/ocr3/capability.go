package ocr3

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/mitchellh/mapstructure"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/datafeeds"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/types"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/mercury"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

const (
	ocrCapabilityID = "offchain_reporting"
)

var info = capabilities.MustNewCapabilityInfo(
	ocrCapabilityID,
	capabilities.CapabilityTypeConsensus,
	"OCR3 consensus exposed as a capability.",
	"v1.0.0",
)

type capability struct {
	services.StateMachine
	capabilities.CapabilityInfo
	store  *store
	stopCh services.StopChan
	wg     sync.WaitGroup
	lggr   logger.Logger

	requestTimeout time.Duration
	clock          clockwork.Clock

	newCleanupWorkerCh chan *request

	aggregators map[string]types.Aggregator

	encoderFactory EncoderFactory
	encoders       map[string]types.Encoder
}

var _ capabilityIface = (*capability)(nil)

func newCapability(s *store, clock clockwork.Clock, requestTimeout time.Duration, encoderFactory EncoderFactory, lggr logger.Logger) *capability {
	o := &capability{
		CapabilityInfo:     info,
		store:              s,
		newCleanupWorkerCh: make(chan *request),
		clock:              clock,
		requestTimeout:     requestTimeout,
		stopCh:             make(chan struct{}),
		lggr:               logger.Named(lggr, "OCR3CapabilityClient"),
		encoderFactory:     encoderFactory,
		aggregators:        map[string]types.Aggregator{},
		encoders:           map[string]types.Encoder{},
	}
	return o
}

func (o *capability) Start(ctx context.Context) error {
	return o.StartOnce("OCR3Capability", func() error {
		o.wg.Add(1)
		go o.loop()
		return nil
	})
}

func (o *capability) Close() error {
	return o.StopOnce("OCR3Capability", func() error {
		close(o.stopCh)
		o.wg.Wait()
		return nil
	})
}

func (o *capability) Name() string { return o.lggr.Name() }

func (o *capability) HealthReport() map[string]error {
	return map[string]error{o.Name(): o.Healthy()}
}

type workflowConfig struct {
	AggregationMethod string         `mapstructure:"aggregation_method"`
	AggregationConfig map[string]any `mapstructure:"aggregation_config"`
	Encoder           string         `mapstructure:"encoder"`
	EncoderConfig     map[string]any `mapstructure:"encoder_config"`
}

func (o *capability) RegisterToWorkflow(ctx context.Context, request capabilities.RegisterToWorkflowRequest) error {
	confMap, err := request.Config.Unwrap()
	if err != nil {
		return err
	}

	// TODO: values lib should a wrapped version of decode
	// which can handle passthrough translations of maps to values.Map.
	// This will avoid the need to translate/untranslate
	c := &workflowConfig{}
	err = mapstructure.Decode(confMap, c)
	if err != nil {
		return err
	}

	if c.AggregationConfig == nil {
		o.lggr.Warn("aggregation_config is empty")
		c.AggregationConfig = map[string]any{}
	}
	if c.EncoderConfig == nil {
		o.lggr.Warn("encoder_config is empty")
		c.EncoderConfig = map[string]any{}
	}

	switch c.AggregationMethod {
	case "data_feeds_2_0":
		cm, err := values.NewMap(c.AggregationConfig)
		if err != nil {
			return err
		}

		mc := mercury.NewCodec()
		agg, err := datafeeds.NewDataFeedsAggregator(*cm, mc, o.lggr)
		if err != nil {
			return err
		}

		o.aggregators[request.Metadata.WorkflowID] = agg

		em, err := values.NewMap(c.EncoderConfig)
		if err != nil {
			return err
		}

		encoder, err := o.encoderFactory(em)
		if err != nil {
			return err
		}
		o.encoders[request.Metadata.WorkflowID] = encoder
	default:
		return fmt.Errorf("aggregator %s not supported", c.AggregationMethod)
	}

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

func (o *capability) UnregisterFromWorkflow(ctx context.Context, request capabilities.UnregisterFromWorkflowRequest) error {
	delete(o.aggregators, request.Metadata.WorkflowID)
	delete(o.encoders, request.Metadata.WorkflowID)
	return nil
}

func (o *capability) Execute(ctx context.Context, callback chan<- capabilities.CapabilityResponse, request capabilities.CapabilityRequest) error {
	// Receives and stores an observation to do consensus on
	// Receives an aggregation method; at this point the method has been validated
	// Returns the consensus result over a channel
	r, err := o.unmarshalRequest(ctx, request, callback)
	if err != nil {
		return err
	}

	err = o.store.add(ctx, r)
	if err != nil {
		return err
	}

	o.newCleanupWorkerCh <- r
	return nil
}

func (o *capability) loop() {
	ctx, cancel := o.stopCh.NewCtx()
	defer cancel()
	defer o.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case r := <-o.newCleanupWorkerCh:
			o.wg.Add(1)
			go o.cleanupWorker(ctx, r)
		}
	}
}

// cleanupWorker is responsible for closing the callback channel.
// It does this either a) when the request has expired, or b)
// when transmitResponse has been called, signalling that we expect no further responses
// to be issued.
// Both of these cases are handled in a single goroutine to ensure that closing the
// channel can only happen once.
func (o *capability) cleanupWorker(ctx context.Context, r *request) {
	defer o.wg.Done()

	d := r.ExpiresAt.Sub(o.clock.Now())
	tr := o.clock.NewTimer(d)
	defer tr.Stop()

	select {
	case <-ctx.Done():
		return
	case <-tr.Chan():
		wasPresent := o.store.evict(ctx, r.WorkflowExecutionID)
		if !wasPresent {
			o.lggr.Errorf("cleanupWorker timed out but could not find store entry for %s", r.WorkflowExecutionID)
			return
		}

		timeoutResp := capabilities.CapabilityResponse{
			Err: fmt.Errorf("timeout exceeded: could not process request before expiry %+v", r.WorkflowExecutionID),
		}

		select {
		case <-r.RequestCtx.Done():
		case r.CallbackCh <- timeoutResp:
			close(r.CallbackCh)
		}
	case <-r.CleanupCh:
		wasPresent := o.store.evict(ctx, r.WorkflowExecutionID)
		if !wasPresent {
			o.lggr.Errorf("cleanupWorker triggered by transmitResponse but could not find a store entry for %s", r.WorkflowExecutionID)
			return
		}

		select {
		case <-r.RequestCtx.Done():
		default:
			close(r.CallbackCh)
		}
	}
}

func (o *capability) transmitResponse(ctx context.Context, resp response) error {
	req, err := o.store.get(ctx, resp.WorkflowExecutionID)
	if err != nil {
		return err
	}

	r := capabilities.CapabilityResponse{
		Value: resp.Value,
		Err:   resp.Err,
	}

	select {
	case <-req.RequestCtx.Done():
		return fmt.Errorf("request canceled: not propagating response %+v to caller", resp)
	case req.CallbackCh <- r:
		req.CleanupCh <- struct{}{}
		return nil
	}
}

func (o *capability) unmarshalRequest(ctx context.Context, r capabilities.CapabilityRequest, callback chan<- capabilities.CapabilityResponse) (*request, error) {
	valuesMap, err := r.Inputs.Unwrap()
	if err != nil {
		return nil, err
	}

	expiresAt := o.clock.Now().Add(o.requestTimeout)
	req := &request{
		RequestCtx:          context.Background(), // TODO: set correct context
		CleanupCh:           make(chan struct{}),
		CallbackCh:          callback,
		WorkflowExecutionID: r.Metadata.WorkflowExecutionID,
		WorkflowID:          r.Metadata.WorkflowID,
		Observations:        r.Inputs.Underlying["observations"],
		ExpiresAt:           expiresAt,
	}
	err = mapstructure.Decode(valuesMap, req)
	if err != nil {
		return nil, err
	}

	configMap, err := r.Config.Unwrap()
	if err != nil {
		return nil, err
	}

	err = mapstructure.Decode(configMap, req)
	if err != nil {
		return nil, err
	}

	return req, err
}
