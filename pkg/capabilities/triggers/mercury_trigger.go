package triggers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

// TODO: Register the capiability in the NewDelegateFunction in chalink/core/services/workflows.delegate.go

var mercuryInfo = capabilities.MustNewCapabilityInfo(
	"mercury-trigger",
	capabilities.CapabilityTypeTrigger,
	"An example mercury trigger.",
	"v1.0.0",
)

// This Trigger Service allows for the registration and deregistration of triggers. You can also send reports to the service.
type MercuryTriggerService struct {
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

type MercuryTriggerEvent struct {
	triggerType string // "mercury"
	id          string // "sha256 of payload.feedId + payload.timestamp"
	timestamp   string // "current time"
	payload     MercuryReport
}

var _ capabilities.TriggerCapability = (*MercuryTriggerService)(nil)

func NewMercuryTriggerService() *MercuryTriggerService {
	return &MercuryTriggerService{
		CapabilityInfo:      mercuryInfo,
		chans:               map[workflowID]chan<- capabilities.CapabilityResponse{},
		feedIdsForTriggerId: make(map[string][]uint64),
		feedIdToWorkflowId:  make(map[uint64]string),
	}
}

// maybe use mercury-pipline reports.go report struct instead of MercuryReport struct, and then we can use the pipeline report generators
func (o *MercuryTriggerService) ProcessReport(report MercuryReport) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	currentTime := time.Now()
	unixTimestampMillis := currentTime.UnixNano() / int64(time.Millisecond)
	feedId := report.feedId

	// Create a new CapabilityResponse with the MercuryTriggerEvent
	triggerEvent := MercuryTriggerEvent{
		triggerType: "mercury",
		id:          sha256Hash(strconv.FormatUint(feedId, 10) + strconv.FormatUint(uint64(report.observationTimestamp), 10)),
		timestamp:   strconv.FormatInt(unixTimestampMillis, 10),
		payload:     report,
	}

	val, err := values.Wrap(triggerEvent)
	if err != nil {
		return err
	}

	// Create a new CapabilityResponse with the MercuryTriggerEvent
	event := capabilities.CapabilityResponse{
		Value: val,
	}

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

func (o *MercuryTriggerService) RegisterTrigger(ctx context.Context, callback chan<- capabilities.CapabilityResponse, req capabilities.CapabilityRequest) error {
	wid := req.Metadata.WorkflowID

	o.mu.Lock()
	defer o.mu.Unlock()

	o.chans[workflowID(wid)] = callback // TODO: if this workflowID is already in the map, throw an error?
	// set feedIdsForTriggerId
	triggerId := o.GetTriggerId(req) // TODO: This is manadatory, so throw an error if its not present or already exists
	feedIds := o.GetFeedIds(req)     // TODO: what if feedIds is empty? should we throw an error or allow it?
	o.feedIdsForTriggerId[triggerId] = feedIds
	// set feedIdToWorkflowId for each feedId
	for _, feedId := range feedIds {
		o.feedIdToWorkflowId[feedId] = wid // TODO: should this be an array?
	}
	return nil
}

func (o *MercuryTriggerService) UnregisterTrigger(ctx context.Context, req capabilities.CapabilityRequest) error {
	wid := req.Metadata.WorkflowID

	o.mu.Lock()
	defer o.mu.Unlock()

	ch, ok := o.chans[workflowID(wid)]
	if ok {
		close(ch)
	}

	triggerId := o.GetTriggerId(req)
	// delete feedIdToWorkflowId for each feedId
	feedIds := o.feedIdsForTriggerId[triggerId]
	for _, feedId := range feedIds {
		delete(o.feedIdToWorkflowId, feedId)
	}

	delete(o.chans, workflowID(wid))
	delete(o.feedIdsForTriggerId, triggerId)

	return nil
}

// Get array of feedIds from CapabilityRequest req
func (o *MercuryTriggerService) GetFeedIds(req capabilities.CapabilityRequest) []uint64 {
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
func (o *MercuryTriggerService) GetTriggerId(req capabilities.CapabilityRequest) string {
	var triggerId string
	// Unwrap the inputs which should return pair (map, nil) and then get the triggerId from the map
	if inputs, err := req.Inputs.Unwrap(); err == nil {
		if id, ok := inputs.(map[string]interface{})["triggerId"].(string); ok {
			triggerId = id
		}
	}
	return triggerId
}

func sha256Hash(s string) string {
	hash := sha256.New()
	hash.Write([]byte(s))
	return hex.EncodeToString(hash.Sum(nil))
}

// TODO: Capability Validation API stub out here
