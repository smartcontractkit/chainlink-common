package ocr3

import (
	"fmt"
	"google.golang.org/protobuf/proto"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/requests"
	pbtypes "github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/types"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/values/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Serializable interface {
	Serialize(lggr logger.Logger) ([]string, []byte, error)
	Len() int
	Mid(mid int) Serializable
}

type QueriesSerializable struct {
	batch []*requests.Request
}

func (q QueriesSerializable) Mid(mid int) Serializable {
	if mid < 0 {
		return QueriesSerializable{batch: []*requests.Request{}}
	}
	if mid >= len(q.batch) {
		return QueriesSerializable{q.batch}
	}
	return QueriesSerializable{q.batch[:mid]}
}

func (q QueriesSerializable) Len() int {
	if q.batch == nil {
		return 0
	}
	return len(q.batch)
}

func (q QueriesSerializable) Serialize(_ logger.Logger) ([]string, []byte, error) {
	ids := make([]*pbtypes.Id, 0)
	allExecutionIDs := make([]string, 0)
	for _, rq := range q.batch {
		ids = append(ids, &pbtypes.Id{
			WorkflowExecutionId:      rq.WorkflowExecutionID,
			WorkflowId:               rq.WorkflowID,
			WorkflowOwner:            rq.WorkflowOwner,
			WorkflowName:             rq.WorkflowName,
			WorkflowDonId:            rq.WorkflowDonID,
			WorkflowDonConfigVersion: rq.WorkflowDonConfigVersion,
			ReportId:                 rq.ReportID,
			KeyId:                    rq.KeyID,
		})
		allExecutionIDs = append(allExecutionIDs, rq.WorkflowExecutionID)
	}

	serialized, err := proto.MarshalOptions{Deterministic: true}.Marshal(&pbtypes.Query{
		Ids: ids,
	})
	return allExecutionIDs, serialized, err
}

type ObservationSerializable struct {
	reqMap                map[string]*requests.Request
	registeredWorkflowIDs []string
}

func (o ObservationSerializable) Serialize(lggr logger.Logger) ([]string, []byte, error) {
	obs := &pbtypes.Observations{}
	allExecutionIDs := make([]string, 0)

	weids := make([]string, 0, len(o.reqMap))
	for k := range o.reqMap {
		weids = append(weids, k)
	}

	for _, weid := range weids {
		rq, ok := o.reqMap[weid]
		if !ok {
			lggr.Debugw("could not find local observations for weid requested in the query", "executionID", weid)
			continue
		}

		lggr := logger.With(
			lggr,
			"executionID", rq.WorkflowExecutionID,
			"workflowID", rq.WorkflowID,
		)

		listProto := values.Proto(rq.Observations).GetListValue()
		if listProto == nil {
			lggr.Errorw("observations are not a list")
			continue
		}

		var cfgProto *pb.Map
		if rq.OverriddenEncoderConfig != nil {
			cp := values.Proto(rq.OverriddenEncoderConfig).GetMapValue()
			cfgProto = cp
		}

		newOb := &pbtypes.Observation{
			Observations: listProto,
			Id: &pbtypes.Id{
				WorkflowExecutionId:      rq.WorkflowExecutionID,
				WorkflowId:               rq.WorkflowID,
				WorkflowOwner:            rq.WorkflowOwner,
				WorkflowName:             rq.WorkflowName,
				WorkflowDonId:            rq.WorkflowDonID,
				WorkflowDonConfigVersion: rq.WorkflowDonConfigVersion,
				ReportId:                 rq.ReportID,
				KeyId:                    rq.KeyID,
			},
			OverriddenEncoderName:   rq.OverriddenEncoderName,
			OverriddenEncoderConfig: cfgProto,
		}

		obs.Observations = append(obs.Observations, newOb)
		allExecutionIDs = append(allExecutionIDs, rq.WorkflowExecutionID)
	}

	obs.RegisteredWorkflowIds = o.registeredWorkflowIDs
	obs.Timestamp = timestamppb.New(time.Now())
	serialized, err := proto.MarshalOptions{Deterministic: true}.Marshal(obs)
	if err != nil {
		lggr.Errorw("failed to serialize observations", "error", err)
		return allExecutionIDs, nil, err
	}

	return allExecutionIDs, serialized, nil
}

func (o ObservationSerializable) Len() int {
	if o.reqMap == nil {
		return 0
	}
	return len(o.reqMap)
}

func (o ObservationSerializable) Mid(mid int) Serializable {
	if mid < 0 {
		return ObservationSerializable{map[string]*requests.Request{}, o.registeredWorkflowIDs}
	}
	if mid >= len(o.reqMap) {
		return o
	}

	weids := make([]string, 0, len(o.reqMap))
	for k := range o.reqMap {
		weids = append(weids, k)
	}

	weids = weids[:mid]

	newMap := make(map[string]*requests.Request, len(weids))
	for _, weid := range weids {
		if req, ok := o.reqMap[weid]; ok {
			newMap[weid] = req
		}
	}

	return ObservationSerializable{
		reqMap: newMap,
	}
}

type OutcomeSerializable struct {
	previousOutcome *pbtypes.Outcome
	weids           []*pbtypes.Id
	outcomes        map[string]*pbtypes.AggregationOutcome
}

func (o OutcomeSerializable) Serialize(_ logger.Logger) ([]string, []byte, error) {
	currentReports := make([]*pbtypes.Report, 0)
	var allExecutionIDs []string

	for _, weid := range o.weids {
		outcome := o.outcomes[weid.WorkflowExecutionId]
		report := &pbtypes.Report{
			Outcome: outcome,
			Id:      weid,
		}
		currentReports = append(currentReports, report)
		allExecutionIDs = append(allExecutionIDs, weid.WorkflowExecutionId)

		o.previousOutcome.Outcomes[weid.WorkflowId] = outcome
	}

	o.previousOutcome.CurrentReports = currentReports
	rawOutcome, err := proto.MarshalOptions{Deterministic: true}.Marshal(o.previousOutcome)

	return allExecutionIDs, rawOutcome, err
}

func (o OutcomeSerializable) Len() int {
	return len(o.weids)
}

func (o OutcomeSerializable) Mid(mid int) Serializable {
	if mid < 0 {
		return OutcomeSerializable{
			o.previousOutcome,
			[]*pbtypes.Id{},
			make(map[string]*pbtypes.AggregationOutcome),
		}
	}
	if mid >= len(o.outcomes) {
		return o
	}

	weids := make([]*pbtypes.Id, 0, len(o.weids))
	for _, k := range o.weids {
		weids = append(weids, k)
	}

	weids = weids[:mid]

	newOutcomes := make(map[string]*pbtypes.AggregationOutcome, len(weids))
	for _, weid := range weids {
		if outcome, ok := o.outcomes[weid.WorkflowExecutionId]; ok {
			newOutcomes[weid.WorkflowExecutionId] = outcome
		}
	}

	return OutcomeSerializable{
		o.previousOutcome,
		weids,
		newOutcomes,
	}

}

// packToSizeLimit function maximizes the number of requests being included in a batch.
// It finds the best utilization of space with the protobuf-marshalled structures using logarithmic (binary search)
// approach to identify the optimal number of Requests that can be serialized without exceeding
// the limit (defaultBatchSizeMiB).
func packToSizeLimit(lggr logger.Logger, all Serializable) ([]string, []byte, error) {
	if all.Len() == 0 {
		return nil, nil, fmt.Errorf("no requests to pack")
	}

	var best []byte
	var bestRequests Serializable
	var bestExecutionIDs []string

	low, high := 0, all.Len()

	optRound := 0

	for low < high {
		mid := (low + high) / 2
		if mid == 0 {
			mid = 1 // poor man's ceil
		}
		candidate := all.Mid(mid)
		executionIDs, serialized, err := candidate.Serialize(lggr)
		if err != nil {
			return nil, nil, err
		}
		size := len(serialized)

		if size <= defaultBatchSizeMiB {
			best = serialized
			bestRequests = candidate
			bestExecutionIDs = executionIDs
			low = mid + 1 // try more Requests
		} else {
			high = mid - 1 // try fewer Requests
		}

		optRound++
	}

	if bestRequests == nil {
		lggr.Warnw("packToSizeLimit: no suitable batch size found, returning empty result")
		return nil, nil, fmt.Errorf("no suitable batch size found")
	}

	lggr.Debugw(
		"packToSizeLimit: best batch size",
		"len",
		bestRequests.Len(),
		"size",
		len(best),
		"maxSize",
		defaultBatchSizeMiB,
		"optRound",
		optRound,
	)

	return bestExecutionIDs, best, nil
}
