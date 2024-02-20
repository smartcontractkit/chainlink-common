package triggers

import (
	"context"
	"fmt"
	"sync"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
)

var mercuryInfo = capabilities.MustNewCapabilityInfo(
	"mercury-trigger",
	capabilities.CapabilityTypeTrigger,
	"An example mercury trigger.",
	"v1.0.0",
)

type MercuryTrigger struct {
	capabilities.CapabilityInfo
	chans               map[workflowID]chan<- capabilities.CapabilityResponse
	feedIdsForTriggerId map[string][]uint64
	mu                  sync.Mutex
	feedIdToWorkflowId  map[uint64]string
}

type MercuryReport struct {
	feedId               uint64
	fullreport           []byte
	benchmarkPrice       int64
	observationTimestamp uint32
}

type MercuryTriggerReport struct {
	triggerType string
	id          string
	timestamp   string
	payload     MercuryReport
}

var _ capabilities.TriggerCapability = (*MercuryTrigger)(nil)

func NewMercuryTrigger() *MercuryTrigger {
	return &MercuryTrigger{
		CapabilityInfo: mercuryInfo,
		chans:          map[workflowID]chan<- capabilities.CapabilityResponse{},
	}
}

// This function needs to send a report, which will be a part of the CapabilityResponse
func (o *MercuryTrigger) SendReport(ctx context.Context, wid string, event capabilities.CapabilityResponse) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	// Decode the TriggerEvent on the CapabilityResponse.
	// unwrap event.Value
	val, err := event.Value.Unwrap()
	if err != nil {
		return err
	}
	// Convert val to MercuryTriggerReport
	report, ok := val.(MercuryTriggerReport)
	if !ok {
		return fmt.Errorf("could not convert val to MercuryTriggerReport")
	}
	feedId := report.payload.feedId

	// If the FeedId is in the feedIdsForTriggerId map, then send the event to the channel.
	if _, ok := o.feedIdToWorkflowId[feedId]; ok {
		wfID := o.feedIdToWorkflowId[feedId]
		ch, ok := o.chans[workflowID(wfID)]
		if ok {
			ch <- event
		} else {
			return fmt.Errorf("no registration for %s", feedId)
		}
	}
	return nil
}

func (o *MercuryTrigger) RegisterTrigger(ctx context.Context, callback chan<- capabilities.CapabilityResponse, req capabilities.CapabilityRequest) error {
	wid := req.Metadata.WorkflowID

	o.mu.Lock()
	defer o.mu.Unlock()

	o.chans[workflowID(wid)] = callback
	// set feedIdsForTriggerId
	triggerId := o.GetTriggerId(req)
	feedIds := o.GetFeedIds(req)
	o.feedIdsForTriggerId[triggerId] = feedIds
	// set feedIdToWorkflowId for each feedId
	for _, feedId := range feedIds {
		o.feedIdToWorkflowId[feedId] = wid
	}
	return nil
}

func (o *MercuryTrigger) UnregisterTrigger(ctx context.Context, req capabilities.CapabilityRequest) error {
	wid := req.Metadata.WorkflowID

	o.mu.Lock()
	defer o.mu.Unlock()

	ch, ok := o.chans[workflowID(wid)]
	if ok {
		close(ch)
	}
	// delete feedIdToWorkflowId for each feedId
	feedIds := o.GetFeedIds(req)
	for _, feedId := range feedIds {
		delete(o.feedIdToWorkflowId, feedId)
	}

	delete(o.chans, workflowID(wid))
	// delete feedIdsForTriggerId
	triggerId := o.GetTriggerId(req)
	delete(o.feedIdsForTriggerId, triggerId)

	return nil
}

// Get array of feedIds from CapabilityRequest req
func (o *MercuryTrigger) GetFeedIds(req capabilities.CapabilityRequest) []uint64 {
	feedIds := make([]uint64, 0)
	// Unwrap the inputs which should return pair (map, nil) and then get the feedIds from the map
	if inputs, err := req.Inputs.Unwrap(); err == nil {
		if feeds, ok := inputs.(map[string]interface{})["feedIds"].([]uint64); ok {
			// Copy to feedIds
			feedIds = append(feedIds, feeds...)
		}
	}
	return feedIds
}

// Get the triggerId from the CapabilityRequest req map
func (o *MercuryTrigger) GetTriggerId(req capabilities.CapabilityRequest) string {
	var triggerId string
	// Unwrap the inputs which should return pair (map, nil) and then get the triggerId from the map
	if inputs, err := req.Inputs.Unwrap(); err == nil {
		if id, ok := inputs.(map[string]interface{})["triggerId"].(string); ok {
			triggerId = id
		}
	}
	return triggerId
}
