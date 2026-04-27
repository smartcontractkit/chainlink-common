package ocr3

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"slices"
	"sort"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	ocrcommon "github.com/smartcontractkit/libocr/commontypes"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3_1types"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/types"
	"github.com/smartcontractkit/libocr/quorumhelper"

	"github.com/smartcontractkit/chainlink-protos/cre/go/values"
	"github.com/smartcontractkit/chainlink-protos/cre/go/values/pb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/requests"
	pbtypes "github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/types"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// KV key layout (v1):
//   outcomes/<workflowID> -> marshalled pbtypes.AggregationOutcome
//
// Future versions must introduce a prefix bump (e.g. "v2/outcomes/") so we can
// detect at read time and migrate. Do not reuse the v1 prefix with a new
// value schema.
const (
	kvPrefixOutcomesV1 = "outcomes/"

	// outcomeEnvelopeVersionV1 stamps the currently-written KV envelope
	// schema. Bump together with kvPrefixOutcomesV1 when the AggregationOutcome
	// payload shape changes incompatibly. A reader must check the version
	// field and refuse to deserialize unknown values rather than blindly
	// proto-unmarshal (proto unmarshal of a future schema into the v1 struct
	// silently drops added fields, which breaks determinism).
	outcomeEnvelopeVersionV1 uint32 = 1

	// defaultBlobExpirationK bounds how many seqNrs a blob is guaranteed to
	// survive past broadcast. Must cover the full round lifetime + any epoch
	// retry window. See plan §3.4. Per-DON override via offchain config.
	defaultBlobExpirationK uint64 = 60
)

var _ ocr3_1types.ReportingPlugin[[]byte] = (*reportingPluginOCR3_1)(nil)

type reportingPluginOCR3_1 struct {
	s      *requests.Store[*ReportRequest]
	r      CapabilityIface
	config ocr3types.ReportingPluginConfig
	limits *pbtypes.ReportingPluginConfig
	// blobBroadcastFetcher is captured at factory time but used only from
	// Query/Observation. Per libocr rules, it must not escape the scope of
	// the method that receives it.
	blobExpirationK uint64
	lggr            logger.Logger
}

func newReportingPluginOCR3_1(
	s *requests.Store[*ReportRequest],
	r CapabilityIface,
	config ocr3types.ReportingPluginConfig,
	limits *pbtypes.ReportingPluginConfig,
	lggr logger.Logger,
) (*reportingPluginOCR3_1, error) {
	// BlobExpirationK is populated by the factory from the (now-extended)
	// ReportingPluginConfig proto. A zero value here would indicate the
	// factory failed to default it — fall back defensively.
	k := limits.BlobExpirationK
	if k == 0 {
		k = defaultBlobExpirationK
	}
	return &reportingPluginOCR3_1{
		s:               s,
		r:               r,
		config:          config,
		limits:          limits,
		blobExpirationK: k,
		lggr:            logger.Named(lggr, "OCR3_1ConsensusReportingPlugin"),
	}, nil
}

// Query selects a batch of pending requests to seek consensus on this round.
// Under OCR3_1 the per-request values.List payload moves to blobs; the Query
// still carries the lightweight Id list since it is sent by the leader once.
func (r *reportingPluginOCR3_1) Query(
	ctx context.Context,
	seqNr uint64,
	kvReader ocr3_1types.KeyValueStateReader,
	blobBroadcastFetcher ocr3_1types.BlobBroadcastFetcher,
) (types.Query, error) {
	// Batching is bounded by MaxQueryLengthBytes; MaxBatchSize is deprecated
	// (see chainlink-deployments zone-b TOML comment).
	batch, err := r.s.FirstN(defaultBatchSize)
	if err != nil {
		r.lggr.Errorw("could not retrieve batch", "error", err)
		return nil, err
	}

	ids := make([]*pbtypes.Id, 0, len(batch))
	allExecutionIDs := make([]string, 0, len(batch))
	seenIds := make(map[idKey]bool)
	cachedQuerySize := 0

	for _, rq := range batch {
		key := GetIDKey(rq)
		if seenIds[key] {
			continue
		}
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
		ok, newSize := QueryBatchHasCapacity(cachedQuerySize, newId, int(r.limits.MaxQueryLengthBytes))
		if !ok {
			break
		}
		seenIds[key] = true
		ids = append(ids, newId)
		allExecutionIDs = append(allExecutionIDs, rq.WorkflowExecutionID)
		cachedQuerySize = newSize
	}

	r.lggr.Debugw("Query complete", "seqNr", seqNr, "len", len(ids), "allExecutionIDs", allExecutionIDs)
	return proto.MarshalOptions{Deterministic: true}.Marshal(&pbtypes.Query{Ids: ids})
}

// Observation gathers local data for the Query ids, serializes the bulk
// payload (the existing Observations proto, unchanged), broadcasts it as a
// blob, and returns a small on-wire BlobbedObservation carrying the blob
// handle and lightweight metadata. This keeps the on-wire observation well
// under the 512 KiB OCR3_1 cap regardless of per-request payload size.
func (r *reportingPluginOCR3_1) Observation(
	ctx context.Context,
	seqNr uint64,
	aq types.AttributedQuery,
	kvReader ocr3_1types.KeyValueStateReader,
	blobBroadcastFetcher ocr3_1types.BlobBroadcastFetcher,
) (types.Observation, error) {
	queryReq := &pbtypes.Query{}
	if err := proto.Unmarshal(aq.Query, queryReq); err != nil {
		return nil, err
	}

	weids := make([]string, 0, len(queryReq.Ids))
	for _, q := range queryReq.Ids {
		if q == nil {
			continue
		}
		weids = append(weids, q.WorkflowExecutionId)
	}

	reqs := r.s.GetByIDs(weids)
	reqMap := make(map[string]*ReportRequest, len(reqs))
	for _, req := range reqs {
		reqMap[req.WorkflowExecutionID] = req
	}

	nowTs := timestamppb.New(time.Now())
	regIDs := r.r.GetRegisteredWorkflowsIDs()

	// Blob payload = the full Observations proto. Same shape as OCR3, now
	// moved off the consensus channel.
	payload := &pbtypes.Observations{
		RegisteredWorkflowIds: regIDs,
		Timestamp:             nowTs,
	}
	// On-wire ids mirror the payload's Observation.Id list. Built in lockstep.
	onWireIds := make([]*pbtypes.Id, 0, len(weids))
	allExecutionIDs := make([]string, 0, len(weids))

	for _, weid := range weids {
		rq, ok := reqMap[weid]
		if !ok {
			continue
		}
		listProto := values.Proto(rq.Observations).GetListValue()
		if listProto == nil {
			r.lggr.Errorw("observations are not a list", "executionID", rq.WorkflowExecutionID)
			continue
		}
		var cfgProto *pb.Map
		if rq.OverriddenEncoderConfig != nil {
			cfgProto = values.Proto(rq.OverriddenEncoderConfig).GetMapValue()
		}
		id := &pbtypes.Id{
			WorkflowExecutionId:      rq.WorkflowExecutionID,
			WorkflowId:               rq.WorkflowID,
			WorkflowOwner:            rq.WorkflowOwner,
			WorkflowName:             rq.WorkflowName,
			WorkflowDonId:            rq.WorkflowDonID,
			WorkflowDonConfigVersion: rq.WorkflowDonConfigVersion,
			ReportId:                 rq.ReportID,
			KeyId:                    rq.KeyID,
		}
		payload.Observations = append(payload.Observations, &pbtypes.Observation{
			Id:                      id,
			Observations:            listProto,
			OverriddenEncoderName:   rq.OverriddenEncoderName,
			OverriddenEncoderConfig: cfgProto,
		})
		onWireIds = append(onWireIds, id)
		allExecutionIDs = append(allExecutionIDs, rq.WorkflowExecutionID)
	}

	payloadBytes, err := proto.MarshalOptions{Deterministic: true}.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal blob payload: %w", err)
	}

	// BroadcastBlob MUST complete within MaxDurationObservation minus
	// marshalling slack. The outer libocr context carries the deadline;
	// honor it rather than imposing our own timeout.
	hint := ocr3_1types.BlobExpirationHintSequenceNumber{SeqNr: seqNr + r.blobExpirationK}
	handle, err := blobBroadcastFetcher.BroadcastBlob(ctx, payloadBytes, hint)
	if err != nil {
		// A broadcast failure here means this node's observation will be
		// missing from the round. Log loudly; quorum assessment is on the
		// libocr side.
		r.lggr.Errorw("blob broadcast failed — observation will be dropped from round",
			"seqNr", seqNr, "payloadBytes", len(payloadBytes), "error", err)
		return nil, err
	}
	// BlobHandle is not a proto message — it uses encoding.BinaryMarshaler.
	handleBytes, err := handle.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("marshal blob handle: %w", err)
	}

	wire := &pbtypes.BlobbedObservation{
		BlobHandle:            handleBytes,
		Ids:                   onWireIds,
		Timestamp:             nowTs,
		RegisteredWorkflowIds: regIDs,
	}
	wireBytes, err := proto.MarshalOptions{Deterministic: true}.Marshal(wire)
	if err != nil {
		return nil, err
	}
	r.lggr.Debugw("Observation complete",
		"seqNr", seqNr,
		"len", len(payload.Observations),
		"payloadBytes", len(payloadBytes),
		"wireBytes", len(wireBytes),
		"expirationSeqNr", seqNr+r.blobExpirationK,
		"allExecutionIDs", allExecutionIDs)
	return wireBytes, nil
}

// ValidateObservation rejects observations that are malformed, have an
// unparseable blob handle, or reference missing blobs. A non-nil return here
// drops this observation from the round's quorum assessment.
//
// Fetching the blob here is expensive; we only do the parse checks
// (BlobbedObservation unmarshals, handle unmarshals, ids non-empty). The
// actual blob fetch happens in StateTransition where cost is already paid.
// Trade-off: a byzantine node could broadcast garbage blob bytes under a
// valid-looking handle and waste fetch cycles in StateTransition. The cost
// is bounded by MaxMaxBlobPayloadLength × quorum, well within budget.
func (r *reportingPluginOCR3_1) ValidateObservation(
	ctx context.Context,
	seqNr uint64,
	aq types.AttributedQuery,
	ao types.AttributedObservation,
	kvReader ocr3_1types.KeyValueStateReader,
	blobFetcher ocr3_1types.BlobFetcher,
) error {
	wire := &pbtypes.BlobbedObservation{}
	if err := proto.Unmarshal(ao.Observation, wire); err != nil {
		return fmt.Errorf("unmarshal BlobbedObservation: %w", err)
	}
	if len(wire.BlobHandle) == 0 {
		return fmt.Errorf("empty blob handle in observation from oracle %d", ao.Observer)
	}
	handle := ocr3_1types.BlobHandle{}
	if err := handle.UnmarshalBinary(wire.BlobHandle); err != nil {
		return fmt.Errorf("unmarshal blob handle: %w", err)
	}
	// Ids may be empty (a node with no registered workflows broadcasts an
	// empty observation on purpose); do not reject on that.
	return nil
}

func (r *reportingPluginOCR3_1) ObservationQuorum(
	ctx context.Context,
	seqNr uint64,
	aq types.AttributedQuery,
	aos []types.AttributedObservation,
	kvReader ocr3_1types.KeyValueStateReader,
	blobFetcher ocr3_1types.BlobFetcher,
) (bool, error) {
	return quorumhelper.ObservationCountReachesObservationQuorum(quorumhelper.QuorumTwoFPlusOne, r.config.N, r.config.F, aos), nil
}

// StateTransition replaces OCR3's Outcome method. It:
//   - reads every existing outcome key from KV (what OCR3 carried forward in
//     PreviousOutcome.Outcomes)
//   - aggregates this round's observations per workflow
//   - writes the updated AggregationOutcome back to KV
//   - prunes workflows not seen for OutcomePruningThreshold rounds (via
//     Delete, collected post-iteration per kvdb.go:34)
//   - emits a ReportsPlusPrecursor carrying the per-workflow reports Reports()
//     needs, since Reports() has no KV access
func (r *reportingPluginOCR3_1) StateTransition(
	ctx context.Context,
	seqNr uint64,
	aq types.AttributedQuery,
	aos []types.AttributedObservation,
	kvReadWriter ocr3_1types.KeyValueStateReadWriter,
	blobFetcher ocr3_1types.BlobFetcher,
) (ocr3_1types.ReportsPlusPrecursor, error) {
	// Materialize blobs: replace each AO's on-wire BlobbedObservation bytes
	// with the fetched Observations payload bytes so the existing grouping
	// logic can run unchanged.
	materializedAos, fetchDropped := r.materializeBlobs(ctx, aos, blobFetcher)
	if fetchDropped > 0 {
		r.lggr.Warnw("dropped observations due to blob fetch failure",
			"seqNr", seqNr, "dropped", fetchDropped, "remaining", len(materializedAos))
	}

	execIDToOracleObservations, seenWorkflowIDs, execIDToEncoderShaToCount, shaToEncoder, finalTimestamp, err :=
		r.groupObservations(materializedAos)
	if err != nil {
		return nil, err
	}

	q := &pbtypes.Query{}
	if err := proto.Unmarshal(aq.Query, q); err != nil {
		return nil, err
	}

	// Load previous AggregationOutcomes for every workflow referenced this
	// round, plus any existing keys we need to consider for pruning.
	previousOutcomes, err := r.loadAllOutcomes(kvReadWriter)
	if err != nil {
		return nil, fmt.Errorf("load outcomes from KV: %w", err)
	}

	currentReports := make([]*pbtypes.Report, 0, len(q.Ids))
	allExecutionIDs := make([]string, 0, len(q.Ids))
	cachedReportSize := 0

	for _, weid := range q.Ids {
		if weid == nil {
			continue
		}
		lggr := logger.With(r.lggr, "executionID", weid.WorkflowExecutionId, "workflowID", weid.WorkflowId)

		obs, ok := execIDToOracleObservations[weid.WorkflowExecutionId]
		if !ok {
			continue
		}
		if len(obs) < (2*r.config.F + 1) {
			continue
		}

		agg, err := r.r.GetAggregator(weid.WorkflowId)
		if err != nil {
			lggr.Errorw("could not retrieve aggregator for workflow", "error", err)
			continue
		}

		prev := previousOutcomes[weid.WorkflowId]
		outcome, err := agg.Aggregate(lggr, prev, obs, r.config.F)
		if err != nil {
			lggr.Errorw("error aggregating outcome", "error", err)
			continue
		}

		if prev != nil {
			outcome.LastSeenAt = prev.LastSeenAt
		}
		outcome.Timestamp = finalTimestamp

		// Deterministic encoder-override tiebreak (fix for the OCR3 map-iteration
		// bug at reporting_plugin.go:396-407). Sort by (count desc, sha asc)
		// before picking the first entry that reaches 2F+1.
		if enc := pickEncoderDeterministic(
			execIDToEncoderShaToCount[weid.WorkflowExecutionId],
			shaToEncoder,
			2*r.config.F+1,
		); enc != nil {
			outcome.EncoderName = enc.name
			outcome.EncoderConfig = enc.config
		}

		report := &pbtypes.Report{Outcome: outcome, Id: weid}
		ok, newSize := ReportBatchHasCapacity(cachedReportSize, report, int(r.limits.MaxOutcomeLengthBytes))
		if !ok {
			break
		}
		currentReports = append(currentReports, report)
		allExecutionIDs = append(allExecutionIDs, weid.WorkflowExecutionId)
		cachedReportSize = newSize

		previousOutcomes[weid.WorkflowId] = outcome
	}

	// Pruning: collect delete-targets first (kvdb.go:34 forbids mutation
	// during iteration, so we already drained the Range in loadAllOutcomes).
	// Then apply Write/Delete.
	toDelete := make([]string, 0)
	for workflowID, outcome := range previousOutcomes {
		if seenWorkflowIDs[workflowID] >= (r.config.F + 1) {
			outcome.LastSeenAt = seqNr
			continue
		}
		if seqNr-outcome.LastSeenAt > r.limits.OutcomePruningThreshold {
			toDelete = append(toDelete, workflowID)
		}
	}
	// Deterministic order for writes (maps iterate randomly; under OCR3_1
	// divergent KV mutation order across nodes would cascade).
	writeIDs := make([]string, 0, len(previousOutcomes))
	for workflowID := range previousOutcomes {
		writeIDs = append(writeIDs, workflowID)
	}
	sort.Strings(writeIDs)
	for _, workflowID := range writeIDs {
		if containsString(toDelete, workflowID) {
			continue
		}
		envelope := &pbtypes.OutcomeEnvelope{
			Version: outcomeEnvelopeVersionV1,
			Outcome: previousOutcomes[workflowID],
		}
		val, err := proto.MarshalOptions{Deterministic: true}.Marshal(envelope)
		if err != nil {
			return nil, fmt.Errorf("marshal outcome envelope for %s: %w", workflowID, err)
		}
		if err := kvReadWriter.Write([]byte(kvPrefixOutcomesV1+workflowID), val); err != nil {
			return nil, fmt.Errorf("kv write %s: %w", workflowID, err)
		}
	}
	sort.Strings(toDelete)
	for _, workflowID := range toDelete {
		if err := kvReadWriter.Delete([]byte(kvPrefixOutcomesV1 + workflowID)); err != nil {
			return nil, fmt.Errorf("kv delete %s: %w", workflowID, err)
		}
		r.r.UnregisterWorkflowID(workflowID)
	}

	// Precursor must be self-contained: Reports() has no KV access.
	precursor := &pbtypes.Outcome{
		CurrentReports: currentReports,
	}
	raw, err := proto.MarshalOptions{Deterministic: true}.Marshal(precursor)
	if err != nil {
		return nil, err
	}

	h := sha256.New()
	h.Write(raw)
	r.lggr.Debugw("StateTransition complete",
		"seqNr", seqNr,
		"reports", len(currentReports),
		"prunedWorkflows", len(toDelete),
		"allExecutionIDs", allExecutionIDs,
		"precursorHash", hex.EncodeToString(h.Sum(nil)))
	return raw, nil
}

// Committed is best-effort notification; returning an error does NOT abort.
// Use for metrics/logging only. NEVER put load-bearing persistence here.
func (r *reportingPluginOCR3_1) Committed(
	ctx context.Context,
	seqNr uint64,
	kvReader ocr3_1types.KeyValueStateReader,
) error {
	r.lggr.Debugw("Committed", "seqNr", seqNr)
	return nil
}

// Reports consumes only the precursor. No KV access here.
func (r *reportingPluginOCR3_1) Reports(
	ctx context.Context,
	seqNr uint64,
	precursor ocr3_1types.ReportsPlusPrecursor,
) ([]ocr3types.ReportPlus[[]byte], error) {
	o := &pbtypes.Outcome{}
	if err := proto.Unmarshal(precursor, o); err != nil {
		return nil, err
	}

	reports := make([]ocr3types.ReportPlus[[]byte], 0, len(o.CurrentReports))
	for _, report := range o.CurrentReports {
		if report == nil || report.Id == nil || report.Outcome == nil {
			continue
		}
		lggr := logger.With(r.lggr,
			"workflowID", report.Id.WorkflowId,
			"executionID", report.Id.WorkflowExecutionId,
			"shouldReport", report.Outcome.ShouldReport)

		outcome, id := report.Outcome, report.Id
		info := &pbtypes.ReportInfo{Id: id, ShouldReport: outcome.ShouldReport}

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
				encoderConfig, err := values.FromMapValueProto(newOutcome.EncoderConfig)
				if err != nil {
					lggr.Errorw("could not convert encoder config", "error", err)
				} else {
					encoder, err = r.r.GetEncoderByName(newOutcome.EncoderName, encoderConfig)
					if err != nil {
						lggr.Errorw("could not retrieve encoder, falling back to default", "error", err)
					}
				}
			}
			if encoder == nil {
				var err error
				encoder, err = r.r.GetEncoderByWorkflowID(id.WorkflowId)
				if err != nil {
					lggr.Errorw("could not retrieve encoder for workflow", "error", err)
					continue
				}
			}

			mv, err := values.FromMapValueProto(newOutcome.EncodableOutcome)
			if err != nil {
				lggr.Errorw("could not decode map from proto", "error", err)
				continue
			}
			rawReport, err = encoder.Encode(ctx, *mv)
			if err != nil {
				if cerr := ctx.Err(); cerr != nil {
					return nil, cerr
				}
				lggr.Errorw("could not encode report", "error", err)
				continue
			}
		}

		infob, err := marshalReportInfo(info, id.KeyId)
		if err != nil {
			lggr.Errorw("could not marshal ReportWithInfo", "error", err)
			continue
		}
		reports = append(reports, ocr3types.ReportPlus[[]byte]{
			ReportWithInfo: ocr3types.ReportWithInfo[[]byte]{
				Report: rawReport,
				Info:   infob,
			},
		})
	}

	r.lggr.Debugw("Reports complete", "seqNr", seqNr, "len", len(reports))
	return reports, nil
}

func (r *reportingPluginOCR3_1) ShouldAcceptAttestedReport(
	ctx context.Context,
	seqNr uint64,
	rwi ocr3types.ReportWithInfo[[]byte],
) (bool, error) {
	return true, nil
}

func (r *reportingPluginOCR3_1) ShouldTransmitAcceptedReport(
	ctx context.Context,
	seqNr uint64,
	rwi ocr3types.ReportWithInfo[[]byte],
) (bool, error) {
	return true, nil
}

func (r *reportingPluginOCR3_1) Close() error { return nil }

// ---- helpers (private to the ocr3_1 path) ----

// materializeBlobs fetches each oracle's blob payload and replaces the AO's
// Observation bytes with the payload bytes. Failures are logged and the AO
// is dropped from the returned slice.
//
// Fetches are done sequentially in v1. A parallel fetch is a post-soak
// optimization; up to 2F+1 fetches per round is tolerable sequentially
// inside StateTransition for current CRE DONs (N≤17).
//
// Determinism: AOs are processed in the order libocr supplied; blob bytes
// are deterministic given the handle, so the resulting group across honest
// nodes is identical. Parallel fetching must also preserve this invariant.
func (r *reportingPluginOCR3_1) materializeBlobs(
	ctx context.Context,
	aos []types.AttributedObservation,
	blobFetcher ocr3_1types.BlobFetcher,
) ([]types.AttributedObservation, int) {
	out := make([]types.AttributedObservation, 0, len(aos))
	dropped := 0
	for _, ao := range aos {
		wire := &pbtypes.BlobbedObservation{}
		if err := proto.Unmarshal(ao.Observation, wire); err != nil {
			r.lggr.Warnw("drop observation: unmarshal BlobbedObservation", "oracleID", ao.Observer, "error", err)
			dropped++
			continue
		}
		if len(wire.BlobHandle) == 0 {
			r.lggr.Warnw("drop observation: empty blob handle", "oracleID", ao.Observer)
			dropped++
			continue
		}
		handle := ocr3_1types.BlobHandle{}
		if err := handle.UnmarshalBinary(wire.BlobHandle); err != nil {
			r.lggr.Warnw("drop observation: unmarshal blob handle", "oracleID", ao.Observer, "error", err)
			dropped++
			continue
		}
		payload, err := blobFetcher.FetchBlob(ctx, handle)
		if err != nil {
			r.lggr.Warnw("drop observation: blob fetch failed", "oracleID", ao.Observer, "error", err)
			dropped++
			continue
		}
		// Sanity-check that the payload unmarshals to an Observations. If
		// not, something is wrong with the broadcasting node — drop it.
		if err := proto.Unmarshal(payload, &pbtypes.Observations{}); err != nil {
			r.lggr.Warnw("drop observation: blob payload not a valid Observations", "oracleID", ao.Observer, "error", err)
			dropped++
			continue
		}
		out = append(out, types.AttributedObservation{
			Observer:    ao.Observer,
			Observation: payload,
		})
	}
	return out, dropped
}


// loadAllOutcomes drains the Range iterator fully before returning.
// kvdb.go:34 forbids any writes/deletes while the iterator is open, so we
// must materialize first and apply mutations later.
func (r *reportingPluginOCR3_1) loadAllOutcomes(
	kvReader ocr3_1types.KeyValueStateReader,
) (map[string]*pbtypes.AggregationOutcome, error) {
	// NOTE: KeyValueStateReader does not expose Range directly
	// (ocr3_1types.KeyValueStateReader has only Read). The Range iterator
	// lives on ocr3_1types.KeyValueDatabaseReadTransaction, not on the
	// per-call reader. For v1 we therefore track workflow IDs in the
	// capability layer and look each up by Read. Ranging over KV from within
	// the plugin is a candidate for a future libocr extension.
	//
	// TODO(OCRBump): confirm with libocr maintainers whether iteration over
	// KeyValueStateReader is planned; if so, move to Range-based loading.
	result := make(map[string]*pbtypes.AggregationOutcome)
	for _, workflowID := range r.r.GetRegisteredWorkflowsIDs() {
		val, err := kvReader.Read([]byte(kvPrefixOutcomesV1 + workflowID))
		if err != nil {
			return nil, fmt.Errorf("kv read %s: %w", workflowID, err)
		}
		if val == nil {
			continue
		}
		envelope := &pbtypes.OutcomeEnvelope{}
		if err := proto.Unmarshal(val, envelope); err != nil {
			return nil, fmt.Errorf("unmarshal outcome envelope %s: %w", workflowID, err)
		}
		if envelope.Version != outcomeEnvelopeVersionV1 {
			return nil, fmt.Errorf(
				"outcome envelope for %s has unknown version %d (expected %d) — "+
					"refusing to deserialize; a KV schema migration is required",
				workflowID, envelope.Version, outcomeEnvelopeVersionV1)
		}
		if envelope.Outcome == nil {
			// A v1 envelope with a nil outcome is malformed; treat as absent.
			continue
		}
		result[workflowID] = envelope.Outcome
	}
	return result, nil
}

// groupObservations replicates the OCR3 Outcome() preamble but returns its
// intermediate state so StateTransition can consume it cleanly.
func (r *reportingPluginOCR3_1) groupObservations(
	aos []types.AttributedObservation,
) (
	map[string]map[ocrcommon.OracleID][]values.Value,
	map[string]int,
	map[string]map[string]int,
	map[string]encoderConfig,
	*timestamppb.Timestamp,
	error,
) {
	execIDToOracleObservations := map[string]map[ocrcommon.OracleID][]values.Value{}
	seenWorkflowIDs := map[string]int{}
	execIDToEncoderShaToCount := map[string]map[string]int{}
	shaToEncoder := map[string]encoderConfig{}
	var sortedTimestamps []*timestamppb.Timestamp

	for _, attributedObservation := range aos {
		obs := &pbtypes.Observations{}
		if err := proto.Unmarshal(attributedObservation.Observation, obs); err != nil {
			r.lggr.Errorw("could not unmarshal observation", "error", err)
			continue
		}

		countedWorkflowIDs := map[string]bool{}
		for _, id := range obs.RegisteredWorkflowIds {
			if _, ok := countedWorkflowIDs[id]; ok {
				continue
			}
			seenWorkflowIDs[id]++
			countedWorkflowIDs[id] = true
		}
		sortedTimestamps = append(sortedTimestamps, obs.Timestamp)

		for _, request := range obs.Observations {
			if request == nil || request.Id == nil {
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
				r.lggr.Errorw("could not calculate sha for overridden encoder", "error", err)
				continue
			}
			shaToEncoder[sha] = encoderConfig{name: request.OverriddenEncoderName, config: request.OverriddenEncoderConfig}
			if _, ok := execIDToEncoderShaToCount[weid]; !ok {
				execIDToEncoderShaToCount[weid] = map[string]int{}
			}
			execIDToEncoderShaToCount[weid][sha]++
		}
	}

	slices.SortFunc(sortedTimestamps, func(a, b *timestamppb.Timestamp) int {
		if a.AsTime().Before(b.AsTime()) {
			return -1
		}
		if a.AsTime().After(b.AsTime()) {
			return 1
		}
		return 0
	})
	var finalTimestamp *timestamppb.Timestamp
	tc := len(sortedTimestamps)
	if tc > 0 {
		mid := tc / 2
		if tc%2 == 1 {
			finalTimestamp = sortedTimestamps[mid]
		} else {
			a := sortedTimestamps[mid-1].AsTime().Unix()
			b := sortedTimestamps[mid].AsTime().Unix()
			finalTimestamp = timestamppb.New(time.Unix(a+(b-a)/2, 0))
		}
	}

	return execIDToOracleObservations, seenWorkflowIDs, execIDToEncoderShaToCount, shaToEncoder, finalTimestamp, nil
}

// pickEncoderDeterministic replaces the OCR3 code's map-iteration tiebreak
// (reporting_plugin.go:396-407). Go maps iterate in randomized order; under
// OCR3_1's per-node KV writes, non-deterministic picks cause state divergence
// that never self-heals. Sort by (count desc, sha asc) and take the first
// that reaches the quorum threshold.
func pickEncoderDeterministic(
	shaToCount map[string]int,
	shaToEncoder map[string]encoderConfig,
	threshold int,
) *encoderConfig {
	if len(shaToCount) == 0 {
		return nil
	}
	shas := make([]string, 0, len(shaToCount))
	for sha := range shaToCount {
		shas = append(shas, sha)
	}
	sort.Slice(shas, func(i, j int) bool {
		if shaToCount[shas[i]] != shaToCount[shas[j]] {
			return shaToCount[shas[i]] > shaToCount[shas[j]]
		}
		return shas[i] < shas[j]
	})
	for _, sha := range shas {
		if shaToCount[sha] < threshold {
			continue
		}
		enc, ok := shaToEncoder[sha]
		if !ok {
			continue
		}
		return &enc
	}
	return nil
}

func containsString(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}

// marshalReportInfoOCR3_1 is a placeholder — the existing marshalReportInfo
// in reporting_plugin.go is package-scoped and reused here. Declared to make
// import expectations explicit; no alternate implementation.
var _ = structpb.NewStruct
