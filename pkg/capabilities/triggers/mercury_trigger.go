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
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/mercury"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

// TODO: Register the capability in the NewDelegateFunction in chainlink/core/services/workflows.delegate.go

var mercuryInfo = capabilities.MustNewCapabilityInfo(
	"mercury-trigger",
	capabilities.CapabilityTypeTrigger,
	"An example mercury trigger.",
	"v1.0.0",
)

// This Trigger Service allows for the registration and deregistration of triggers. You can also send reports to the service.
type MercuryTriggerService struct {
	capabilities.CapabilityInfo
	chans                 map[workflowID]chan<- capabilities.CapabilityResponse
	feedIdsForTriggerId   map[string][]int64 // TODO: switch this to uint64 when value.go supports it
	mu                    sync.Mutex
	feedIdToWorkflowId    map[int64]string
	triggerIdToWorkflowId map[string]string
}

var _ capabilities.TriggerCapability = (*MercuryTriggerService)(nil)

func NewMercuryTriggerService() *MercuryTriggerService {
	return &MercuryTriggerService{
		CapabilityInfo:        mercuryInfo,
		chans:                 map[workflowID]chan<- capabilities.CapabilityResponse{},
		feedIdsForTriggerId:   make(map[string][]int64),
		triggerIdToWorkflowId: make(map[string]string),
	}
}

// maybe use mercury-pipline reports.go report struct instead of MercuryReport struct, and then we can use the pipeline report generators
func (o *MercuryTriggerService) ProcessReport(reports []mercury.MercuryReport) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	currentTime := time.Now()
	unixTimestampMillis := currentTime.UnixNano() / int64(time.Millisecond)
	triggerIdsForReports := make(map[string][]int)
	reportIndex := 0

	for _, report := range reports {
		// for each feed id, we need to find the triggerId associated with it.
		reportFeedId := report.FeedId
		for triggerId, feedIds := range o.feedIdsForTriggerId {
			for _, feedId := range feedIds {
				if reportFeedId == feedId {
					// if its not initalized, initialize it
					if _, ok := triggerIdsForReports[triggerId]; !ok {
						triggerIdsForReports[triggerId] = make([]int, 0)
					}
					triggerIdsForReports[triggerId] = append(triggerIdsForReports[triggerId], reportIndex)
				}
			}
		}
		reportIndex += 1
	}

	// Then for each trigger id, find which reports correspond to that trigger and create an event bundling the reports
	// and send it to the channel associated with the trigger id.
	for triggerId, reportIds := range triggerIdsForReports {
		reportPayload := make([]mercury.MercuryReport, 0)
		for _, reportId := range reportIds {
			reportPayload = append(reportPayload, reports[reportId])
		}

		triggerEvent := mercury.MercuryTriggerEvent{
			TriggerType: "mercury",
			Id:          generateTriggerEventId(reportPayload),
			Timestamp:   strconv.FormatInt(unixTimestampMillis, 10),
			Payload:     reportPayload,
		}

		// TODO: Modify values.Wrap to handle MercuryTriggerEvent and MercuryReport structs
		val, err := mercury.Codec{}.WrapMercuryTriggerEvent(triggerEvent)
		if err != nil {
			return err
		}

		// Create a new CapabilityResponse with the MercuryTriggerEvent
		capabilityResponse := capabilities.CapabilityResponse{
			Value: val,
		}

		// If the FeedId is in the feedIdsForTriggerId map, then send the event to the channel.

		wfID := o.triggerIdToWorkflowId[triggerId]
		ch, ok := o.chans[workflowID(wfID)]
		if ok {
			ch <- capabilityResponse
		} else {
			return fmt.Errorf("no registration for %s", triggerId)
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

	o.triggerIdToWorkflowId[triggerId] = wid
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

	delete(o.chans, workflowID(wid))
	delete(o.feedIdsForTriggerId, triggerId)
	delete(o.triggerIdToWorkflowId, triggerId)

	return nil
}

// Get array of feedIds from CapabilityRequest req
func (o *MercuryTriggerService) GetFeedIds(req capabilities.CapabilityRequest) []int64 {
	feedIds := make([]int64, 0)
	// Unwrap the inputs which should return pair (map, nil) and then get the feedIds from the map
	if inputs, err := req.Inputs.Unwrap(); err == nil {
		if feeds, ok := inputs.(map[string]interface{})["feedIds"].([]int64); ok {
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

// TODO: The generated id should probably be the sha256 of all the report feed ids and timestamps associated with this trigger event
func generateTriggerEventId(reports []mercury.MercuryReport) string {
	// Let's hash all the feedIds and timestamps together
	s := ""
	for _, report := range reports {
		s += strconv.FormatInt(report.FeedId, 10) + strconv.FormatInt(report.ObservationTimestamp, 10) + ","
	}
	return sha256Hash(s)
}

func ValidateInput(mercuryTriggerEvent values.Value) error {
	// TODO: Fill this in
	return nil
}

func ExampleOutput() (values.Value, error) {
	event := mercury.MercuryTriggerEvent{
		TriggerType: "mercury",
		Id:          "123",
		Timestamp:   "2024-01-17T04:00:10Z",
		Payload: []mercury.MercuryReport{
			{
				FeedId:               2,
				Fullreport:           []byte("hello"),
				BenchmarkPrice:       100,
				ObservationTimestamp: 123,
			},
			{
				FeedId:               3,
				Fullreport:           []byte("world"),
				BenchmarkPrice:       100,
				ObservationTimestamp: 123,
			},
		},
	}
	return mercury.Codec{}.WrapMercuryTriggerEvent(event)
}

func ValidateConfig(config values.Value) error {
	// TODO: Fill this in
	return nil
}
