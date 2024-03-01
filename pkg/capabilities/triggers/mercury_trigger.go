package triggers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
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
	chans                 map[string]chan<- capabilities.CapabilityResponse
	feedIDsForTriggerID   map[string][]int64 // TODO: switch this to uint64 when value.go supports it
	mu                    sync.Mutex
	triggerIDToWorkflowID map[string]string
}

var _ capabilities.TriggerCapability = (*MercuryTriggerService)(nil)

func NewMercuryTriggerService() *MercuryTriggerService {
	return &MercuryTriggerService{
		CapabilityInfo:        mercuryInfo,
		chans:                 map[string]chan<- capabilities.CapabilityResponse{},
		feedIDsForTriggerID:   make(map[string][]int64),
		triggerIDToWorkflowID: make(map[string]string),
	}
}

// maybe use mercury-pipline reports.go report struct instead of MercuryReport struct, and then we can use the pipeline report generators
func (o *MercuryTriggerService) ProcessReport(reports []mercury.FeedReport) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	currentTime := time.Now()
	unixTimestampMillis := currentTime.UnixNano() / int64(time.Millisecond)
	triggerIDsForReports := make(map[string][]int)
	reportIndex := 0

	for _, report := range reports {
		// for each feed id, we need to find the triggerId associated with it.
		reportFeedID := report.FeedID
		for triggerID, feedIDs := range o.feedIDsForTriggerID {
			for _, feedID := range feedIDs {
				if reportFeedID == feedID {
					// if its not initialized, initialize it
					if _, ok := triggerIDsForReports[triggerID]; !ok {
						triggerIDsForReports[triggerID] = make([]int, 0)
					}
					triggerIDsForReports[triggerID] = append(triggerIDsForReports[triggerID], reportIndex)
				}
			}
		}
		reportIndex++
	}

	// Then for each trigger id, find which reports correspond to that trigger and create an event bundling the reports
	// and send it to the channel associated with the trigger id.
	for triggerID, reportIDs := range triggerIDsForReports {
		reportPayload := make([]mercury.FeedReport, 0)
		for _, reportID := range reportIDs {
			reportPayload = append(reportPayload, reports[reportID])
		}

		triggerEvent := mercury.TriggerEvent{
			TriggerType: "mercury",
			ID:          GenerateTriggerEventID(reportPayload),
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

		wfID := o.triggerIDToWorkflowID[triggerID]
		ch, ok := o.chans[wfID+triggerID]
		if ok {
			ch <- capabilityResponse
		} else {
			return fmt.Errorf("no registration for %s", triggerID)
		}
	}
	return nil
}

func (o *MercuryTriggerService) RegisterTrigger(ctx context.Context, callback chan<- capabilities.CapabilityResponse, req capabilities.CapabilityRequest) error {
	wid := req.Metadata.WorkflowID

	o.mu.Lock()
	defer o.mu.Unlock()

	// set feedIdsForTriggerId
	triggerID := o.GetTriggerID(req) // TODO: This is manadatory, so throw an error if its not present or already exists
	feedIDs := o.GetFeedIDs(req)     // TODO: what if feedIds is empty? should we throw an error or allow it?
	// concat wid and triggerID to get the key
	o.chans[wid+triggerID] = callback
	o.feedIDsForTriggerID[triggerID] = feedIDs

	o.triggerIDToWorkflowID[triggerID] = wid
	return nil
}

func (o *MercuryTriggerService) UnregisterTrigger(ctx context.Context, req capabilities.CapabilityRequest) error {
	wid := req.Metadata.WorkflowID

	o.mu.Lock()
	defer o.mu.Unlock()

	triggerID := o.GetTriggerID(req)

	ch, ok := o.chans[wid+triggerID]
	if ok {
		close(ch)
	}

	delete(o.chans, wid+triggerID)
	delete(o.feedIDsForTriggerID, triggerID)
	delete(o.triggerIDToWorkflowID, triggerID)

	return nil
}

// Get array of feedIds from CapabilityRequest req
func (o *MercuryTriggerService) GetFeedIDs(req capabilities.CapabilityRequest) []int64 {
	feedIDs := make([]int64, 0)
	// Unwrap the inputs which should return pair (map, nil) and then get the feedIds from the map
	if inputs, err := req.Inputs.Unwrap(); err == nil {
		if feeds, ok := inputs.(map[string]interface{})["feedIds"].([]any); ok {
			// Copy to feedIds
			for _, feed := range feeds {
				if id, ok := feed.(int64); ok {
					feedIDs = append(feedIDs, id)
				}
			}
		}
	}
	return feedIDs
}

// Get the triggerId from the CapabilityRequest req map
func (o *MercuryTriggerService) GetTriggerID(req capabilities.CapabilityRequest) string {
	var triggerID string
	// Unwrap the inputs which should return pair (map, nil) and then get the triggerId from the map
	if inputs, err := req.Inputs.Unwrap(); err == nil {
		if id, ok := inputs.(map[string]interface{})["triggerId"].(string); ok {
			triggerID = id
		}
	}
	return triggerID
}

func sha256Hash(s string) string {
	hash := sha256.New()
	hash.Write([]byte(s))
	return hex.EncodeToString(hash.Sum(nil))
}

func GenerateTriggerEventID(reports []mercury.FeedReport) string {
	// Let's hash all the feedIds and timestamps together
	sort.Slice(reports, func(i, j int) bool {
		if reports[i].FeedID == reports[j].FeedID {
			return reports[i].ObservationTimestamp < reports[j].ObservationTimestamp
		}
		return reports[i].FeedID < reports[j].FeedID
	})
	s := ""
	for _, report := range reports {
		s += strconv.FormatInt(report.FeedID, 10) + strconv.FormatInt(report.ObservationTimestamp, 10) + ","
	}
	return sha256Hash(s)
}

func ValidateInput(mercuryTriggerEvent values.Value) error {
	// TODO: Fill this in
	return nil
}

func ExampleOutput() (values.Value, error) {
	event := mercury.TriggerEvent{
		TriggerType: "mercury",
		ID:          "123",
		Timestamp:   "2024-01-17T04:00:10Z",
		Payload: []mercury.FeedReport{
			{
				FeedID:               2,
				FullReport:           []byte("hello"),
				BenchmarkPrice:       100,
				ObservationTimestamp: 123,
			},
			{
				FeedID:               3,
				FullReport:           []byte("world"),
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
