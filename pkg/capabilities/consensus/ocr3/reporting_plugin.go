package ocr3

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"slices"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/smartcontractkit/libocr/quorumhelper"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/requests"
	ocrcommon "github.com/smartcontractkit/libocr/commontypes"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/smartcontractkit/chainlink-protos/cre/go/values"
	"github.com/smartcontractkit/chainlink-protos/cre/go/values/pb"

	pbtypes "github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/types"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

var _ ocr3types.ReportingPlugin[[]byte] = (*reportingPlugin)(nil)

type CapabilityIface interface {
	GetAggregator(workflowID string) (pbtypes.Aggregator, error)
	GetEncoderByWorkflowID(workflowID string) (pbtypes.Encoder, error)
	GetEncoderByName(encoderName string, config *values.Map) (pbtypes.Encoder, error)
	GetRegisteredWorkflowsIDs() []string
	UnregisterWorkflowID(workflowID string)
}

type reportingPlugin struct {
	batchSize int
	s         *requests.Store[*ReportRequest]
	r         CapabilityIface
	config    ocr3types.ReportingPluginConfig
	limits    *pbtypes.ReportingPluginConfig
	lggr      logger.Logger
}

func NewReportingPlugin(s *requests.Store[*ReportRequest], r CapabilityIface, batchSize int, config ocr3types.ReportingPluginConfig,
	limits *pbtypes.ReportingPluginConfig, lggr logger.Logger) (*reportingPlugin, error) {
	return &reportingPlugin{
		s:         s,
		r:         r,
		batchSize: batchSize,
		config:    config,
		limits:    limits,
		lggr:      logger.Named(lggr, "OCR3ConsensusReportingPlugin"),
	}, nil
}

func (r *reportingPlugin) Query(ctx context.Context, outctx ocr3types.OutcomeContext) (types.Query, error) {
	batch, err := r.s.FirstN(r.batchSize)

	if err != nil {
		r.lggr.Errorw("could not retrieve batch", "error", err)
		return nil, err
	}

	ids := []*pbtypes.Id{}
	allExecutionIDs := []string{}
	seenIds := make(map[idKey]bool)
	cachedQuerySize := 0

	for _, rq := range batch {
		key := GetIDKey(rq)
		newId := &pbtypes.Id{
			WorkflowExecutionId:      rq.WorkflowExecutionID,
			WorkflowId:               rq.WorkflowID,
			WorkflowOwner:            rq.WorkflowOwner,
			WorkflowName:             rq.WorkflowName,
			WorkflowDonId:            rq.WorkflowDonID,
			WorkflowDonConfigVersion: rq.WorkflowDonConfigVersion,
			ReportId:                 rq.ReportID,
			KeyId:                    rq.KeyID,
		}

		// Simple duplicate elimination using a map
		if seenIds[key] {
			continue
		}

		// If the new id would exceed the max query size, stop adding more ids
		ok, newSize := QueryBatchHasCapacity(cachedQuerySize, newId, int(r.limits.MaxQueryLengthBytes))
		if !ok {
			break
		}

		seenIds[key] = true
		ids = append(ids, newId)
		allExecutionIDs = append(allExecutionIDs, rq.WorkflowExecutionID)
		cachedQuerySize = newSize
	}

	r.lggr.Debugw("Query complete", "len", len(ids), "allExecutionIDs", allExecutionIDs)
	return proto.MarshalOptions{Deterministic: true}.Marshal(&pbtypes.Query{
		Ids: ids,
	})
}

func (r *reportingPlugin) Observation(ctx context.Context, outctx ocr3types.OutcomeContext, query types.Query) (types.Observation, error) {
	queryReq := &pbtypes.Query{}
	err := proto.Unmarshal(query, queryReq)
	if err != nil {
		return nil, err
	}

	weids := []string{}
	for _, q := range queryReq.Ids {
		if q == nil {
			r.lggr.Debugw("skipping nil id for query", "query", queryReq)
			continue
		}
		weids = append(weids, q.WorkflowExecutionId)
	}

	reqs := r.s.GetByIDs(weids)
	reqMap := map[string]*ReportRequest{}
	for _, req := range reqs {
		reqMap[req.WorkflowExecutionID] = req
	}

	obs := &pbtypes.Observations{
		RegisteredWorkflowIds: r.r.GetRegisteredWorkflowsIDs(),
		Timestamp:             timestamppb.New(time.Now()),
	}
	allExecutionIDs := []string{}

	// Initialize cached size with the base message size (RegisteredWorkflowIds and Timestamp)
	cachedObsSize := CalculateObservationsMessageSize(obs)

	for _, weid := range weids {
		rq, ok := reqMap[weid]
		if !ok {
			r.lggr.Debugw("could not find local observations for weid requested in the query", "executionID", weid)
			continue
		}

		lggr := logger.With(
			r.lggr,
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

		ok, newSize := ObservationsBatchHasCapacity(cachedObsSize, newOb, int(r.limits.MaxObservationLengthBytes))
		if !ok {
			break
		}

		obs.Observations = append(obs.Observations, newOb)
		allExecutionIDs = append(allExecutionIDs, rq.WorkflowExecutionID)
		cachedObsSize = newSize
	}

	r.lggr.Debugw("Observation complete", "len", len(obs.Observations), "queryLen", len(queryReq.Ids), "allExecutionIDs", allExecutionIDs)
	return proto.MarshalOptions{Deterministic: true}.Marshal(obs)
}

func (r *reportingPlugin) ValidateObservation(ctx context.Context, outctx ocr3types.OutcomeContext, query types.Query, ao types.AttributedObservation) error {
	return nil
}

func (r *reportingPlugin) ObservationQuorum(ctx context.Context, outctx ocr3types.OutcomeContext, query types.Query, aos []types.AttributedObservation) (bool, error) {
	return quorumhelper.ObservationCountReachesObservationQuorum(quorumhelper.QuorumTwoFPlusOne, r.config.N, r.config.F, aos), nil
}

func shaForOverriddenEncoder(obs *pbtypes.Observation) (string, error) {
	hash := sha256.New()
	_, err := hash.Write([]byte(obs.OverriddenEncoderName))
	if err != nil {
		return "", fmt.Errorf("could not write encoder name to hash: %w", err)
	}

	marshalled, err := proto.MarshalOptions{Deterministic: true}.Marshal(obs.OverriddenEncoderConfig)
	if err != nil {
		return "", fmt.Errorf("could not marshal overridden encoder: %w", err)
	}

	_, err = hash.Write(marshalled)
	if err != nil {
		return "", fmt.Errorf("could not write encoder config to hash: %w", err)
	}

	return string(hash.Sum([]byte{})), nil
}

type encoderConfig struct {
	name   string
	config *pb.Map
}

func (r *reportingPlugin) Outcome(ctx context.Context, outctx ocr3types.OutcomeContext, query types.Query, attributedObservations []types.AttributedObservation) (ocr3types.Outcome, error) {
	// execution ID -> oracle ID -> list of observations
	execIDToOracleObservations := map[string]map[ocrcommon.OracleID][]values.Value{}
	seenWorkflowIDs := map[string]int{}
	var sortedTimestamps []*timestamppb.Timestamp
	var finalTimestamp *timestamppb.Timestamp
	execIDToEncoderShaToCount := map[string]map[string]int{}
	shaToEncoder := map[string]encoderConfig{}
	for _, attributedObservation := range attributedObservations {
		obs := &pbtypes.Observations{}
		err := proto.Unmarshal(attributedObservation.Observation, obs)
		if err != nil {
			r.lggr.Errorw("could not unmarshal observation", "error", err, "observation", obs)
			continue
		}

		countedWorkflowIDs := map[string]bool{}
		for _, id := range obs.RegisteredWorkflowIds {
			// Skip if we've already counted this workflow ID. we want to avoid duplicates in the seen workflow IDs.
			if _, ok := countedWorkflowIDs[id]; ok {
				continue
			}

			// Count how many times a workflow ID is seen from Observations, no need for initial value since it's 0 by default.
			seenWorkflowIDs[id]++

			countedWorkflowIDs[id] = true
		}

		sortedTimestamps = append(sortedTimestamps, obs.Timestamp)

		for _, request := range obs.Observations {
			if request == nil {
				r.lggr.Debugw("skipping nil request in observations", "observations", obs.Observations)
				continue
			}

			if request.Id == nil {
				r.lggr.Debugw("skipping nil id in request", "request", request)
				continue
			}

			weid := request.Id.WorkflowExecutionId

			obsList, innerErr := values.FromListValueProto(request.Observations)
			if obsList == nil || innerErr != nil {
				r.lggr.Errorw("observations are not a list", "weID", weid, "oracleID", attributedObservation.Observer, "err", innerErr)
				continue
			}

			if _, ok := execIDToOracleObservations[weid]; !ok {
				execIDToOracleObservations[weid] = make(map[ocrcommon.OracleID][]values.Value)
			}
			execIDToOracleObservations[weid][attributedObservation.Observer] = obsList.Underlying

			sha, err := shaForOverriddenEncoder(request)
			if err != nil {
				r.lggr.Errorw("could not calculate sha for overridden encoder", "error", err, "observation", obs)
				continue
			}

			shaToEncoder[sha] = encoderConfig{
				name:   request.OverriddenEncoderName,
				config: request.OverriddenEncoderConfig,
			}
			if _, ok := execIDToEncoderShaToCount[weid]; !ok {
				execIDToEncoderShaToCount[weid] = map[string]int{}
			}
			execIDToEncoderShaToCount[weid][sha]++
		}
	}

	// Since we will most likely get N different timestamps, each with frequency=1, we get the median instead of the mode.
	slices.SortFunc(sortedTimestamps, func(a, b *timestamppb.Timestamp) int {
		if a.AsTime().Before(b.AsTime()) {
			return -1
		}
		if a.AsTime().After(b.AsTime()) {
			return 1
		}
		return 0
	})
	timestampCount := len(sortedTimestamps)
	mid := timestampCount / 2
	if timestampCount%2 == 1 {
		finalTimestamp = sortedTimestamps[mid]
	} else {
		a := sortedTimestamps[mid-1].AsTime().Unix()
		b := sortedTimestamps[mid].AsTime().Unix()
		// a + (b-a) / 2 to avoid overflows
		finalTimestamp = timestamppb.New(time.Unix(a+(b-a)/2, 0))
	}

	q := &pbtypes.Query{}
	err := proto.Unmarshal(query, q)
	if err != nil {
		return nil, err
	}

	previousOutcome := &pbtypes.Outcome{}
	err = proto.Unmarshal(outctx.PreviousOutcome, previousOutcome)
	if err != nil {
		return nil, err
	}
	if previousOutcome.Outcomes == nil {
		previousOutcome.Outcomes = map[string]*pbtypes.AggregationOutcome{}
	}

	// Wipe out the CurrentReports. This gets regenerated
	// every time since we only want to transmit reports that
	// are part of the current Query.
	previousOutcome.CurrentReports = []*pbtypes.Report{}
	var allExecutionIDs []string
	cachedReportSize := 0

	for _, weid := range q.Ids {
		if weid == nil {
			r.lggr.Debugw("skipping nil id in query", "query", q)
			continue
		}
		lggr := logger.With(r.lggr, "executionID", weid.WorkflowExecutionId, "workflowID", weid.WorkflowId)
		obs, ok := execIDToOracleObservations[weid.WorkflowExecutionId]
		if !ok {
			lggr.Debugw("could not find any observations matching weid requested in the query")
			continue
		}

		workflowOutcome, ok := previousOutcome.Outcomes[weid.WorkflowId]
		if !ok {
			lggr.Debugw("could not find existing outcome for workflow, aggregator will create a new one")
		}

		if len(obs) < (2*r.config.F + 1) {
			lggr.Warnw("Insufficient observations for workflow execution: consensus skipped because fewer than 2f+1 nodes reported observations",
				"observationCount", len(obs), "requiredCount", 2*r.config.F+1)
			continue
		}

		agg, err2 := r.r.GetAggregator(weid.WorkflowId)
		if err2 != nil {
			lggr.Errorw("could not retrieve aggregator for workflow", "error", err2)
			continue
		}

		outcome, err2 := agg.Aggregate(lggr, workflowOutcome, obs, r.config.F)
		if err2 != nil {
			lggr.Errorw("error aggregating outcome", "error", err2)
			continue
		}

		// Only if the previous outcome exists:
		// We carry the last seen round from the previous outcome, since the aggregation does carry it.
		// So each `Aggregate()` call will return an outcome with a zero value for LastSeenAt.
		if workflowOutcome != nil {
			outcome.LastSeenAt = workflowOutcome.LastSeenAt
		}

		outcome.Timestamp = finalTimestamp

		shaToCount, ok := execIDToEncoderShaToCount[weid.WorkflowExecutionId]
		if !ok {
			lggr.Debugw("could not find any encoder shas matching weid requested in the query")
			continue
		}

		// Note: no need to check the observation count here,
		// we've checked this above when we checked the observations count.
		var encCfg *encoderConfig
		for sha, count := range shaToCount {
			if count >= 2*r.config.F+1 {
				encoderCfg, ok := shaToEncoder[sha]
				if !ok {
					lggr.Debugw("could not find encoder matching sha")
					continue
				}

				lggr.Debugw("consensus reached on overridden encoder", "encoderName", encoderCfg.name)
				encCfg = &encoderCfg
				break
			}
		}

		if encCfg != nil {
			lggr.Debugw("overridden encoder set", "name", encCfg.name, "cfg", encCfg.config)
			outcome.EncoderName = encCfg.name
			outcome.EncoderConfig = encCfg.config
		}

		report := &pbtypes.Report{
			Outcome: outcome,
			Id:      weid,
		}

		ok, newSize := ReportBatchHasCapacity(cachedReportSize, report, int(r.limits.MaxOutcomeLengthBytes))
		if !ok {
			break
		}

		previousOutcome.CurrentReports = append(previousOutcome.CurrentReports, report)
		allExecutionIDs = append(allExecutionIDs, weid.WorkflowExecutionId)
		cachedReportSize = newSize

		previousOutcome.Outcomes[weid.WorkflowId] = outcome
	}

	// We need to prune outcomes from previous workflows that are no longer relevant.
	for workflowID, outcome := range previousOutcome.Outcomes {
		// Update the last seen round for this outcome. But this should only happen if the workflow is seen by F+1 nodes.
		if seenWorkflowIDs[workflowID] >= (r.config.F + 1) {
			r.lggr.Debugw("updating last seen round of outcome for workflow", "workflowID", workflowID)
			outcome.LastSeenAt = outctx.SeqNr
		} else if outctx.SeqNr-outcome.LastSeenAt > r.limits.OutcomePruningThreshold {
			r.lggr.Debugw("pruning outcome for workflow", "workflowID", workflowID, "SeqNr", outctx.SeqNr, "lastSeenAt", outcome.LastSeenAt)
			delete(previousOutcome.Outcomes, workflowID)
			r.r.UnregisterWorkflowID(workflowID)
		}
	}

	rawOutcome, err := proto.MarshalOptions{Deterministic: true}.Marshal(previousOutcome)
	h := sha256.New()
	h.Write(rawOutcome)
	outcomeHash := h.Sum(nil)
	r.lggr.Debugw("Outcome complete", "len", len(previousOutcome.Outcomes), "nAggregatedWorkflowExecutions", len(previousOutcome.CurrentReports), "allExecutionIDs", allExecutionIDs, "outcomeHash", hex.EncodeToString(outcomeHash), "err", err)
	return rawOutcome, err
}

func marshalReportInfo(info *pbtypes.ReportInfo, keyID string) ([]byte, error) {
	p, err := proto.MarshalOptions{Deterministic: true}.Marshal(info)
	if err != nil {
		return nil, err
	}

	infos, err := structpb.NewStruct(map[string]any{
		"keyBundleName": keyID,
		"reportInfo":    p,
	})
	if err != nil {
		return nil, err
	}

	ip, err := proto.MarshalOptions{Deterministic: true}.Marshal(infos)
	if err != nil {
		return nil, err
	}

	return ip, nil
}

func (r *reportingPlugin) Reports(ctx context.Context, seqNr uint64, outcome ocr3types.Outcome) ([]ocr3types.ReportPlus[[]byte], error) {
	o := &pbtypes.Outcome{}
	err := proto.Unmarshal(outcome, o)
	if err != nil {
		return nil, err
	}

	reports := []ocr3types.ReportPlus[[]byte]{}

	for _, report := range o.CurrentReports {
		if report == nil {
			r.lggr.Debugw("skipping nil report in outcome", "outcome", o)
			continue
		}

		if report.Id == nil {
			r.lggr.Debugw("skipping report with nil id in outcome", "report", report)
			continue
		}

		if report.Outcome == nil {
			r.lggr.Debugw("skipping report with nil outcome", "report", report)
			continue
		}

		lggr := logger.With(
			r.lggr,
			"workflowID", report.Id.WorkflowId,
			"executionID", report.Id.WorkflowExecutionId,
			"shouldReport", report.Outcome.ShouldReport,
		)
		lggr.Debugw("generating reports", "len", len(o.CurrentReports))

		outcome, id := report.Outcome, report.Id

		info := &pbtypes.ReportInfo{
			Id:           id,
			ShouldReport: outcome.ShouldReport,
		}

		var rawReport []byte
		if info.ShouldReport {
			meta := &pbtypes.Metadata{
				Version:          1,
				ExecutionID:      id.WorkflowExecutionId,
				Timestamp:        uint32(outcome.Timestamp.AsTime().Unix()),
				DONID:            id.WorkflowDonId,
				DONConfigVersion: id.WorkflowDonConfigVersion,
				WorkflowID:       id.WorkflowId,
				WorkflowName:     id.WorkflowName,
				WorkflowOwner:    id.WorkflowOwner,
				ReportID:         id.ReportId,
			}
			newOutcome, err := pbtypes.AppendMetadata(outcome, meta)
			if err != nil {
				lggr.Errorw("could not append IDs")
				continue
			}

			var encoder pbtypes.Encoder
			if newOutcome.EncoderName != "" {
				lggr.Debugw("using encoder from outcome", "encoderName", newOutcome.EncoderName, "executionID", report.Id.WorkflowExecutionId)
				encoderConfig, err2 := values.FromMapValueProto(newOutcome.EncoderConfig)
				if err2 != nil {
					lggr.Errorw("could not convert desired encoder config to values.Map", "error", err2, "executionID", report.Id.WorkflowExecutionId)
				} else {
					encoder, err2 = r.r.GetEncoderByName(newOutcome.EncoderName, encoderConfig)
					if err2 != nil {
						lggr.Errorw("could not retrieve desired encoder, will use per-workflow default", "error", err2, "executionID", report.Id.WorkflowExecutionId)
					}
				}
			}

			if encoder == nil {
				encoder, err = r.r.GetEncoderByWorkflowID(id.WorkflowId)
				if err != nil {
					lggr.Errorw("could not retrieve encoder for workflow", "error", err)
					continue
				}
			}

			mv, err := values.FromMapValueProto(newOutcome.EncodableOutcome)
			if err != nil {
				lggr.Errorw("could not decode map from map value proto", "error", err)
				continue
			}

			rawReport, err = encoder.Encode(ctx, *mv)
			if err != nil {
				if cerr := ctx.Err(); cerr != nil {
					lggr.Errorw("report encoding cancelled", "err", cerr)
					return nil, cerr
				}
				lggr.Errorw("could not encode report for workflow", "error", err)
				continue
			}
		}

		infob, err := marshalReportInfo(info, id.KeyId)
		if err != nil {
			lggr.Errorw("could not marshal id into ReportWithInfo", "error", err)
			continue
		}

		// Append every report, even if shouldReport = false, to let the transmitter mark the step as complete.
		reports = append(reports, ocr3types.ReportPlus[[]byte]{
			ReportWithInfo: ocr3types.ReportWithInfo[[]byte]{
				Report: rawReport,
				Info:   infob,
			},
		})
	}

	r.lggr.Debugw("Reports complete", "len", len(reports))
	return reports, nil
}

func (r *reportingPlugin) ShouldAcceptAttestedReport(ctx context.Context, seqNr uint64, rwi ocr3types.ReportWithInfo[[]byte]) (bool, error) {
	// True because we always want to transmit a report, even if shouldReport = false.
	return true, nil
}

func (r *reportingPlugin) ShouldTransmitAcceptedReport(ctx context.Context, seqNr uint64, rwi ocr3types.ReportWithInfo[[]byte]) (bool, error) {
	// True because we always want to transmit a report, even if shouldReport = false.
	return true, nil
}

func (r *reportingPlugin) Close() error {
	return nil
}
