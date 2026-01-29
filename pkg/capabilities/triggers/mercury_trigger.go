package triggers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/datastreams"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/triggers/streams"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-protos/cre/go/values"
)

const (
	defaultCapabilityName     = "streams-trigger"
	defaultCapabilityVersion  = "1.1.0"
	defaultTickerResolutionMs = 1000
	// TODO pending capabilities configuration implementation - this should be configurable with a sensible default
	defaultSendChannelBufferSize = 1000
)

// This Trigger Service allows for the registration and deregistration of triggers. You can also send reports to the service.
type MercuryTriggerService struct {
	capabilities.CapabilityInfo
	tickerResolutionMs int64
	subscribers        map[string]*subscriber
	latestReports      map[datastreams.FeedID]datastreams.FeedReport
	mu                 sync.Mutex
	stopCh             services.StopChan
	wg                 sync.WaitGroup
	lggr               logger.Logger
	metaOverride       datastreams.Metadata // usually empty, but set to a value in mock trigger
}

var _ capabilities.TriggerCapability = (*MercuryTriggerService)(nil)
var _ services.Service = &MercuryTriggerService{}

type subscriber struct {
	ch         chan<- capabilities.TriggerResponse
	workflowID string
	config     streams.TriggerConfig
}

// Mercury Trigger will send events to each subscriber every MaxFrequencyMs (configurable per subscriber).
// Event generation happens whenever local unix time is a multiple of tickerResolutionMs. Therefore,
// all subscribers' MaxFrequencyMs values need to be a multiple of tickerResolutionMs.
func NewMercuryTriggerService(tickerResolutionMs int64, capName string, capVersion string, lggr logger.Logger) (*MercuryTriggerService, error) {
	if tickerResolutionMs == 0 {
		tickerResolutionMs = defaultTickerResolutionMs
	}
	if capName == "" {
		capName = defaultCapabilityName
	}
	if capVersion == "" {
		capVersion = defaultCapabilityVersion
	}
	capInfo, err := capabilities.NewCapabilityInfo(
		capName+"@"+capVersion,
		capabilities.CapabilityTypeTrigger,
		"Streams Trigger",
	)
	if err != nil {
		return nil, err
	}
	return &MercuryTriggerService{
		CapabilityInfo:     capInfo,
		tickerResolutionMs: tickerResolutionMs,
		subscribers:        make(map[string]*subscriber),
		latestReports:      make(map[datastreams.FeedID]datastreams.FeedReport),
		stopCh:             make(services.StopChan),
		lggr:               logger.Named(lggr, "MercuryTriggerService")}, nil
}

func (o *MercuryTriggerService) SetMetaOverride(meta datastreams.Metadata) {
	o.metaOverride = meta
}

func (o *MercuryTriggerService) ProcessReport(reports []datastreams.FeedReport) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.lggr.Debugw("ProcessReport", "nReports", len(reports))
	for _, report := range reports {
		feedID := datastreams.FeedID(report.FeedID)
		o.latestReports[feedID] = report
	}
	return nil
}

func (o *MercuryTriggerService) AckEvent(ctx context.Context, triggerId string, eventId string) error {
	return nil
}

func (o *MercuryTriggerService) RegisterTrigger(ctx context.Context, req capabilities.TriggerRegistrationRequest) (<-chan capabilities.TriggerResponse, error) {
	wid := req.Metadata.WorkflowID

	o.mu.Lock()
	defer o.mu.Unlock()

	config, err := o.ValidateConfig(req.Config)
	if err != nil {
		return nil, err
	}

	// If triggerId is already registered, return an error
	if _, ok := o.subscribers[req.TriggerID]; ok {
		return nil, fmt.Errorf("triggerId %s already registered", o.ID)
	}

	if int64(config.MaxFrequencyMs)%o.tickerResolutionMs != 0 {
		return nil, fmt.Errorf("MaxFrequencyMs must be a multiple of %d", o.tickerResolutionMs)
	}

	ch := make(chan capabilities.TriggerResponse, defaultSendChannelBufferSize)
	o.subscribers[req.TriggerID] =
		&subscriber{
			ch:         ch,
			workflowID: wid,
			config:     *config,
		}
	return ch, nil
}

func (o *MercuryTriggerService) ValidateConfig(config *values.Map) (*streams.TriggerConfig, error) {
	cfg := &streams.TriggerConfig{}
	if err := config.UnwrapTo(cfg); err != nil {
		return nil, err
	}

	// TODO QOL improvement, the generator for the builders can add a validate function that just copies code after unmarshalling to Plain
	b, _ := json.Marshal(cfg)
	if err := json.Unmarshal(b, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (o *MercuryTriggerService) UnregisterTrigger(ctx context.Context, req capabilities.TriggerRegistrationRequest) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	subscriber, ok := o.subscribers[req.TriggerID]
	if !ok {
		return fmt.Errorf("triggerId %s not registered", o.ID)
	}
	close(subscriber.ch)
	delete(o.subscribers, req.TriggerID)
	return nil
}

func (o *MercuryTriggerService) loop() {
	defer o.wg.Done()
	now := time.Now().UnixMilli()
	nextWait := o.tickerResolutionMs - now%o.tickerResolutionMs

	for {
		select {
		case <-o.stopCh:
			return
		case <-time.After(time.Duration(nextWait) * time.Millisecond):
			startTs := time.Now().UnixMilli()
			// find closest timestamp that is a multiple of o.tickerResolutionMs
			aligned := (startTs + o.tickerResolutionMs/2) / o.tickerResolutionMs * o.tickerResolutionMs
			o.process(aligned)
			endTs := time.Now().UnixMilli()
			if endTs-startTs > o.tickerResolutionMs {
				o.lggr.Errorw("processing took longer than ticker resolution", "duration", endTs-startTs, "tickerResolutionMs", o.tickerResolutionMs)
			}
			nextWait = getNextWaitIntervalMs(aligned, o.tickerResolutionMs, endTs)
		}
	}
}

func getNextWaitIntervalMs(lastTs, tickerResolutionMs, currentTs int64) int64 {
	desiredNext := lastTs + tickerResolutionMs
	nextWait := max(desiredNext-currentTs, 0)
	return nextWait
}

func (o *MercuryTriggerService) process(timestamp int64) {
	o.mu.Lock()
	defer o.mu.Unlock()
	for _, sub := range o.subscribers {
		if timestamp%int64(sub.config.MaxFrequencyMs) == 0 {
			reportList := make([]datastreams.FeedReport, 0)
			for _, feedID := range sub.config.FeedIds {
				if latest, ok := o.latestReports[datastreams.FeedID(feedID)]; ok {
					reportList = append(reportList, latest)
				}
			}

			// use 32-byte-padded timestamp as EventID (human-readable)
			eventID := fmt.Sprintf("streams_%024s", strconv.FormatInt(timestamp, 10))
			capabilityResponse, err := WrapReports(reportList, eventID, timestamp, o.metaOverride, o.ID)
			if err != nil {
				o.lggr.Errorw("error wrapping reports", "err", err)
				continue
			}

			o.lggr.Debugw("ProcessReport pushing event", "nReports", len(reportList), "eventID", eventID)
			select {
			case sub.ch <- capabilityResponse:
			default:
				o.lggr.Errorw("subscriber channel full, dropping event", "eventID", eventID, "workflowID", sub.workflowID)
			}
		}
	}
}

func WrapReports(reportList []datastreams.FeedReport, eventID string, timestamp int64, meta datastreams.Metadata, capID string) (capabilities.TriggerResponse, error) {
	out := datastreams.StreamsTriggerEvent{
		Payload:   reportList,
		Metadata:  meta,
		Timestamp: timestamp,
	}
	outputsv, err := values.WrapMap(out)
	if err != nil {
		return capabilities.TriggerResponse{}, err
	}

	// Create a new TriggerRegistrationResponse with the MercuryTriggerEvent
	return capabilities.TriggerResponse{
		Event: capabilities.TriggerEvent{
			TriggerType: capID,
			ID:          eventID,
			Outputs:     outputsv,
		},
	}, nil
}

func (o *MercuryTriggerService) Start(ctx context.Context) error {
	o.wg.Add(1)
	go o.loop()
	o.lggr.Info("MercuryTriggerService started")
	return nil
}

func (o *MercuryTriggerService) Close() error {
	close(o.stopCh)
	o.wg.Wait()
	o.lggr.Info("MercuryTriggerService closed")
	return nil
}

func (o *MercuryTriggerService) Ready() error {
	return nil
}

func (o *MercuryTriggerService) HealthReport() map[string]error {
	return nil
}

func (o *MercuryTriggerService) Name() string {
	return o.lggr.Name()
}
