package ocr3

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/smartcontractkit/libocr/commontypes"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/requests"

	pbtypes "github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/types"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/values/pb"
)

func TestReportingPlugin_Query_ErrorInQueueCall(t *testing.T) {
	ctx := tests.Context(t)
	lggr := logger.Test(t)
	s := requests.NewStore()
	batchSize := 0
	rp, err := newReportingPlugin(s, nil, batchSize, ocr3types.ReportingPluginConfig{}, defaultOutcomePruningThreshold, lggr)
	require.NoError(t, err)

	outcomeCtx := ocr3types.OutcomeContext{
		PreviousOutcome: []byte(""),
	}
	_, err = rp.Query(ctx, outcomeCtx)
	assert.Error(t, err)
}

func TestReportingPlugin_Query(t *testing.T) {
	ctx := tests.Context(t)
	lggr := logger.Test(t)
	s := requests.NewStore()
	rp, err := newReportingPlugin(s, nil, defaultBatchSize, ocr3types.ReportingPluginConfig{}, defaultOutcomePruningThreshold, lggr)
	require.NoError(t, err)

	eid := uuid.New().String()
	wowner := uuid.New().String()

	err = s.Add(&requests.Request{
		WorkflowID:          workflowTestID,
		WorkflowExecutionID: eid,
		WorkflowOwner:       wowner,
		WorkflowName:        workflowTestName,
		ReportID:            reportTestID,
	})
	require.NoError(t, err)
	outcomeCtx := ocr3types.OutcomeContext{
		PreviousOutcome: []byte(""),
	}

	q, err := rp.Query(ctx, outcomeCtx)
	require.NoError(t, err)

	qry := &pbtypes.Query{}
	err = proto.Unmarshal(q, qry)
	require.NoError(t, err)

	assert.Len(t, qry.Ids, 1)
	assert.Equal(t, qry.Ids[0].WorkflowId, workflowTestID)
	assert.Equal(t, qry.Ids[0].WorkflowExecutionId, eid)
}

type mockCapability struct {
	t                   *testing.T
	aggregator          *aggregator
	encoder             *enc
	registeredWorkflows map[string]bool
	expectedEncoderName string
}

type aggregator struct {
	gotObs  map[commontypes.OracleID][]values.Value
	outcome *pbtypes.AggregationOutcome
}

func (a *aggregator) Aggregate(pout *pbtypes.AggregationOutcome, observations map[commontypes.OracleID][]values.Value, _ int) (*pbtypes.AggregationOutcome, error) {
	a.gotObs = observations
	nm, err := values.NewMap(
		map[string]any{
			"aggregated": "outcome",
		},
	)
	if err != nil {
		return nil, err
	}
	a.outcome = &pbtypes.AggregationOutcome{
		EncodableOutcome: values.Proto(nm).GetMapValue(),
	}
	return a.outcome, nil
}

type enc struct {
	gotInput values.Map
}

func (e *enc) Encode(ctx context.Context, input values.Map) ([]byte, error) {
	e.gotInput = input
	return proto.Marshal(values.Proto(&input))
}

func (mc *mockCapability) getAggregator(workflowID string) (pbtypes.Aggregator, error) {
	return mc.aggregator, nil
}

func (mc *mockCapability) getEncoderByWorkflowID(workflowID string) (pbtypes.Encoder, error) {
	return mc.encoder, nil
}

func (mc *mockCapability) getEncoderByName(encoderName string, config *values.Map) (pbtypes.Encoder, error) {
	require.Equal(mc.t, mc.expectedEncoderName, encoderName)
	return mc.encoder, nil
}

func (mc *mockCapability) getRegisteredWorkflowsIDs() []string {
	workflows := make([]string, 0, len(mc.registeredWorkflows))
	for wf := range mc.registeredWorkflows {
		workflows = append(workflows, wf)
	}
	return workflows
}

func (mc *mockCapability) unregisterWorkflowID(workflowID string) {
	delete(mc.registeredWorkflows, workflowID)
}

func TestReportingPlugin_Observation(t *testing.T) {
	ctx := tests.Context(t)
	lggr := logger.Test(t)
	s := requests.NewStore()
	mcap := &mockCapability{
		aggregator: &aggregator{},
		encoder:    &enc{},
		registeredWorkflows: map[string]bool{
			workflowTestID:  true,
			workflowTestID2: true,
		},
	}
	rp, err := newReportingPlugin(s, mcap, defaultBatchSize, ocr3types.ReportingPluginConfig{}, defaultOutcomePruningThreshold, lggr)
	require.NoError(t, err)

	o, err := values.NewList([]any{"hello"})
	require.NoError(t, err)

	eid := uuid.New().String()
	wowner := uuid.New().String()
	err = s.Add(&requests.Request{
		WorkflowID:          workflowTestID,
		WorkflowExecutionID: eid,
		WorkflowOwner:       wowner,
		WorkflowName:        workflowTestName,
		ReportID:            reportTestID,
		Observations:        o,
	})
	require.NoError(t, err)
	outcomeCtx := ocr3types.OutcomeContext{
		PreviousOutcome: []byte(""),
	}

	q, err := rp.Query(ctx, outcomeCtx)
	require.NoError(t, err)

	obs, err := rp.Observation(ctx, outcomeCtx, q)
	require.NoError(t, err)

	obspb := &pbtypes.Observations{}
	err = proto.Unmarshal(obs, obspb)
	require.NoError(t, err)

	assert.Len(t, obspb.Observations, 1)
	fo := obspb.Observations[0]
	assert.Equal(t, fo.Id.WorkflowExecutionId, eid)
	assert.Equal(t, fo.Id.WorkflowId, workflowTestID)
	lvp, err := values.FromListValueProto(fo.Observations)
	require.NoError(t, err)
	assert.Equal(t, o, lvp)
	expected := []string{workflowTestID, workflowTestID2}
	actual := obspb.RegisteredWorkflowIds
	sort.Slice(actual, func(i, j int) bool { return actual[i] < actual[j] })
	sort.Slice(expected, func(i, j int) bool { return expected[i] < expected[j] })
	assert.Equal(t, expected, actual)
}

func TestReportingPlugin_Observation_NilIds(t *testing.T) {
	ctx := tests.Context(t)
	lggr := logger.Test(t)
	s := requests.NewStore()
	mcap := &mockCapability{
		aggregator: &aggregator{},
		encoder:    &enc{},
		registeredWorkflows: map[string]bool{
			workflowTestID:  true,
			workflowTestID2: true,
		},
	}
	rp, err := newReportingPlugin(s, mcap, defaultBatchSize, ocr3types.ReportingPluginConfig{}, defaultOutcomePruningThreshold, lggr)
	require.NoError(t, err)

	outcomeCtx := ocr3types.OutcomeContext{
		PreviousOutcome: []byte(""),
	}

	q := &pbtypes.Query{
		Ids: []*pbtypes.Id{
			nil,
			{
				WorkflowExecutionId: uuid.New().String(),
			},
		},
	}
	qb, err := proto.Marshal(q)
	require.NoError(t, err)

	_, err = rp.Observation(ctx, outcomeCtx, qb)
	require.NoError(t, err)
}

func TestReportingPlugin_Observation_NoResults(t *testing.T) {
	ctx := tests.Context(t)
	lggr := logger.Test(t)
	s := requests.NewStore()
	mcap := &mockCapability{
		aggregator: &aggregator{},
		encoder:    &enc{},
	}
	rp, err := newReportingPlugin(s, mcap, defaultBatchSize, ocr3types.ReportingPluginConfig{}, defaultOutcomePruningThreshold, lggr)
	require.NoError(t, err)

	outcomeCtx := ocr3types.OutcomeContext{
		PreviousOutcome: []byte(""),
	}

	q, err := rp.Query(ctx, outcomeCtx)
	require.NoError(t, err)

	obs, err := rp.Observation(ctx, outcomeCtx, q)
	require.NoError(t, err)

	obspb := &pbtypes.Observations{}
	err = proto.Unmarshal(obs, obspb)
	require.NoError(t, err)

	assert.Len(t, obspb.Observations, 0)
}

func TestReportingPlugin_Outcome(t *testing.T) {
	lggr := logger.Test(t)
	s := requests.NewStore()
	mcap := &mockCapability{
		aggregator: &aggregator{},
		encoder:    &enc{},
	}
	rp, err := newReportingPlugin(s, mcap, defaultBatchSize, ocr3types.ReportingPluginConfig{}, defaultOutcomePruningThreshold, lggr)
	require.NoError(t, err)

	weid := uuid.New().String()
	wowner := uuid.New().String()
	id := &pbtypes.Id{
		WorkflowExecutionId: weid,
		WorkflowId:          workflowTestID,
		WorkflowOwner:       wowner,
		WorkflowName:        workflowTestName,
		ReportId:            reportTestID,
	}
	q := &pbtypes.Query{
		Ids: []*pbtypes.Id{id},
	}
	qb, err := proto.Marshal(q)
	require.NoError(t, err)
	o, err := values.NewList([]any{"hello"})
	require.NoError(t, err)
	obs := &pbtypes.Observations{
		Observations: []*pbtypes.Observation{
			{
				Id:           id,
				Observations: values.Proto(o).GetListValue(),
			},
		},
	}

	rawObs, err := proto.Marshal(obs)
	require.NoError(t, err)
	aos := []types.AttributedObservation{
		{
			Observation: rawObs,
			Observer:    commontypes.OracleID(1),
		},
	}

	outcome, err := rp.Outcome(ocr3types.OutcomeContext{}, qb, aos)
	require.NoError(t, err)

	opb := &pbtypes.Outcome{}
	err = proto.Unmarshal(outcome, opb)
	require.NoError(t, err)

	assert.Len(t, opb.CurrentReports, 1)

	cr := opb.CurrentReports[0]
	assert.EqualExportedValues(t, cr.Id, id)
	assert.EqualExportedValues(t, cr.Outcome, mcap.aggregator.outcome)
	assert.EqualExportedValues(t, opb.Outcomes[workflowTestID], mcap.aggregator.outcome)
}

func TestReportingPlugin_Outcome_NilDerefs(t *testing.T) {
	lggr := logger.Test(t)
	s := requests.NewStore()
	mcap := &mockCapability{
		aggregator: &aggregator{},
		encoder:    &enc{},
	}
	rp, err := newReportingPlugin(s, mcap, defaultBatchSize, ocr3types.ReportingPluginConfig{}, defaultOutcomePruningThreshold, lggr)
	require.NoError(t, err)

	weid := uuid.New().String()
	wowner := uuid.New().String()
	id := &pbtypes.Id{
		WorkflowExecutionId: weid,
		WorkflowId:          workflowTestID,
		WorkflowOwner:       wowner,
		WorkflowName:        workflowTestName,
		ReportId:            reportTestID,
	}
	q := &pbtypes.Query{
		Ids: []*pbtypes.Id{
			id,
			nil,
		},
	}
	qb, err := proto.Marshal(q)
	require.NoError(t, err)
	aos := []types.AttributedObservation{
		{
			Observer: commontypes.OracleID(1),
		},
		{},
	}

	_, err = rp.Outcome(ocr3types.OutcomeContext{}, qb, aos)
	require.NoError(t, err)

	obs := &pbtypes.Observations{
		Observations: []*pbtypes.Observation{
			nil,
			{},
		},
		RegisteredWorkflowIds: nil,
	}
	obsb, err := proto.Marshal(obs)
	require.NoError(t, err)

	aos = []types.AttributedObservation{
		{
			Observation: obsb,
			Observer:    commontypes.OracleID(1),
		},
	}
	_, err = rp.Outcome(ocr3types.OutcomeContext{}, qb, aos)
	require.NoError(t, err)
}

func TestReportingPlugin_Reports_ShouldReportFalse(t *testing.T) {
	lggr := logger.Test(t)
	s := requests.NewStore()
	mcap := &mockCapability{
		aggregator: &aggregator{},
		encoder:    &enc{},
	}
	rp, err := newReportingPlugin(s, mcap, defaultBatchSize, ocr3types.ReportingPluginConfig{}, defaultOutcomePruningThreshold, lggr)
	require.NoError(t, err)

	var sqNr uint64
	weid := uuid.New().String()
	wowner := uuid.New().String()
	id := &pbtypes.Id{
		WorkflowExecutionId: weid,
		WorkflowId:          workflowTestID,
		WorkflowOwner:       wowner,
	}
	nm, err := values.NewMap(
		map[string]any{
			"our": "aggregation",
		},
	)
	require.NoError(t, err)
	outcome := &pbtypes.Outcome{
		CurrentReports: []*pbtypes.Report{
			{
				Id: id,
				Outcome: &pbtypes.AggregationOutcome{
					EncodableOutcome: values.Proto(nm).GetMapValue(),
				},
			},
		},
	}
	pl, err := proto.Marshal(outcome)
	require.NoError(t, err)
	reports, err := rp.Reports(sqNr, pl)
	require.NoError(t, err)

	assert.Len(t, reports, 1)
	gotRep := reports[0]
	assert.Len(t, gotRep.Report, 0)

	ib := gotRep.Info
	info, err := extractReportInfo(ib)
	require.NoError(t, err)

	assert.EqualExportedValues(t, info.Id, id)
	assert.False(t, info.ShouldReport)
}

func TestReportingPlugin_Reports_NilDerefs(t *testing.T) {
	lggr := logger.Test(t)
	s := requests.NewStore()
	mcap := &mockCapability{
		aggregator: &aggregator{},
		encoder:    &enc{},
	}
	rp, err := newReportingPlugin(s, mcap, defaultBatchSize, ocr3types.ReportingPluginConfig{}, defaultOutcomePruningThreshold, lggr)
	require.NoError(t, err)

	var sqNr uint64
	weid := uuid.New().String()
	wowner := uuid.New().String()
	id := &pbtypes.Id{
		WorkflowExecutionId: weid,
		WorkflowId:          workflowTestID,
		WorkflowOwner:       wowner,
	}
	require.NoError(t, err)
	outcome := &pbtypes.Outcome{
		CurrentReports: []*pbtypes.Report{
			{
				Id: id,
				Outcome: &pbtypes.AggregationOutcome{
					EncodableOutcome: nil,
				},
			},
			{},
			{
				Outcome: &pbtypes.AggregationOutcome{},
			},
			{
				Id: id,
			},
		},
	}
	pl, err := proto.Marshal(outcome)
	require.NoError(t, err)
	_, err = rp.Reports(sqNr, pl)
	require.NoError(t, err)
}

func TestReportingPlugin_Reports_ShouldReportTrue(t *testing.T) {
	lggr := logger.Test(t)
	dynamicEncoderName := "special_encoder"
	s := requests.NewStore()
	mcap := &mockCapability{
		t:                   t,
		aggregator:          &aggregator{},
		encoder:             &enc{},
		expectedEncoderName: dynamicEncoderName,
	}
	rp, err := newReportingPlugin(s, mcap, defaultBatchSize, ocr3types.ReportingPluginConfig{}, defaultOutcomePruningThreshold, lggr)
	require.NoError(t, err)

	var sqNr uint64
	weid := uuid.New().String()
	wowner := uuid.New().String()
	donID := uint32(1)
	id := &pbtypes.Id{
		WorkflowExecutionId: weid,
		WorkflowId:          workflowTestID,
		WorkflowOwner:       wowner,
		WorkflowName:        workflowTestName,
		ReportId:            reportTestID,
		WorkflowDonId:       donID,
	}
	nm, err := values.NewMap(
		map[string]any{
			"our": "aggregation",
		},
	)
	nmp := values.Proto(nm).GetMapValue()
	require.NoError(t, err)
	outcome := &pbtypes.Outcome{
		CurrentReports: []*pbtypes.Report{
			{
				Id: id,
				Outcome: &pbtypes.AggregationOutcome{
					EncodableOutcome: nmp,
					ShouldReport:     true,
					EncoderName:      dynamicEncoderName,
				},
			},
		},
	}
	pl, err := proto.Marshal(outcome)
	require.NoError(t, err)
	reports, err := rp.Reports(sqNr, pl)
	require.NoError(t, err)

	assert.Len(t, reports, 1)
	gotRep := reports[0]

	rep := &pb.Value{}
	err = proto.Unmarshal(gotRep.Report, rep)
	require.NoError(t, err)

	// The workflow ID and execution ID get added to the report.
	nm.Underlying[pbtypes.MetadataFieldName], err = values.NewMap(map[string]any{
		"Version":          1,
		"ExecutionID":      weid,
		"Timestamp":        0,
		"DONID":            donID,
		"DONConfigVersion": 0,
		"WorkflowID":       workflowTestID,
		"WorkflowName":     workflowTestName,
		"WorkflowOwner":    wowner,
		"ReportID":         reportTestID,
	})
	require.NoError(t, err)
	fp, err := values.FromProto(rep)
	require.NoError(t, err)
	require.Equal(t, nm, fp)

	ib := gotRep.Info
	info, err := extractReportInfo(ib)
	require.NoError(t, err)

	assert.EqualExportedValues(t, info.Id, id)
	assert.True(t, info.ShouldReport)
}

func TestReportingPlugin_Outcome_ShouldPruneOldOutcomes(t *testing.T) {
	lggr := logger.Test(t)
	s := requests.NewStore()
	mcap := &mockCapability{
		aggregator: &aggregator{},
		encoder:    &enc{},
		registeredWorkflows: map[string]bool{
			workflowTestID:  true,
			workflowTestID2: true,
		},
	}
	rp, err := newReportingPlugin(s, mcap, defaultBatchSize, ocr3types.ReportingPluginConfig{}, defaultOutcomePruningThreshold, lggr)
	require.NoError(t, err)

	weid := uuid.New().String()
	wowner := uuid.New().String()
	id := &pbtypes.Id{
		WorkflowExecutionId: weid,
		WorkflowId:          workflowTestID,
		WorkflowOwner:       wowner,
		WorkflowName:        workflowTestName,
		ReportId:            reportTestID,
	}
	id2 := &pbtypes.Id{
		WorkflowExecutionId: weid,
		WorkflowId:          workflowTestID2,
		WorkflowOwner:       wowner,
		WorkflowName:        workflowTestName,
		ReportId:            reportTestID,
	}
	id3 := &pbtypes.Id{
		WorkflowExecutionId: weid,
		WorkflowId:          workflowTestID3,
		WorkflowOwner:       wowner,
		WorkflowName:        workflowTestName,
		ReportId:            reportTestID,
	}
	q := &pbtypes.Query{
		Ids: []*pbtypes.Id{id, id2, id3},
	}
	qb, err := proto.Marshal(q)
	require.NoError(t, err)
	o, err := values.NewList([]any{"hello"})
	require.NoError(t, err)
	obs := &pbtypes.Observations{
		Observations: []*pbtypes.Observation{
			{
				Id:           id,
				Observations: values.Proto(o).GetListValue(),
			},
			{
				Id:           id2,
				Observations: values.Proto(o).GetListValue(),
			},
			{
				Id:           id3,
				Observations: values.Proto(o).GetListValue(),
			},
		},
		RegisteredWorkflowIds: []string{workflowTestID, workflowTestID2},
	}
	obs2 := &pbtypes.Observations{
		Observations: []*pbtypes.Observation{
			{
				Id:           id,
				Observations: values.Proto(o).GetListValue(),
			},
			{
				Id:           id2,
				Observations: values.Proto(o).GetListValue(),
			},
			{
				Id:           id3,
				Observations: values.Proto(o).GetListValue(),
			},
		},
		RegisteredWorkflowIds: []string{workflowTestID},
	}

	rawObs, err := proto.Marshal(obs)
	require.NoError(t, err)
	rawObs2, err := proto.Marshal(obs2)
	require.NoError(t, err)
	aos := []types.AttributedObservation{
		{
			Observation: rawObs,
			Observer:    commontypes.OracleID(1),
		},
	}
	aos2 := []types.AttributedObservation{
		{
			Observation: rawObs2,
			Observer:    commontypes.OracleID(1),
		},
	}

	outcome1, err := rp.Outcome(ocr3types.OutcomeContext{SeqNr: 100}, qb, aos)
	require.NoError(t, err)
	opb1 := &pbtypes.Outcome{}
	err = proto.Unmarshal(outcome1, opb1)
	require.NoError(t, err)

	outcome2, err := rp.Outcome(ocr3types.OutcomeContext{SeqNr: defaultOutcomePruningThreshold + 100, PreviousOutcome: outcome1}, qb, aos2)
	require.NoError(t, err)
	opb2 := &pbtypes.Outcome{}
	err = proto.Unmarshal(outcome2, opb2)
	require.NoError(t, err)

	assert.Equal(t, uint64(100), opb1.Outcomes[workflowTestID].LastSeenAt)
	assert.Equal(t, uint64(100), opb1.Outcomes[workflowTestID2].LastSeenAt)
	assert.Equal(t, uint64(0), opb1.Outcomes[workflowTestID3].LastSeenAt)
	assert.Equal(t, uint64(defaultOutcomePruningThreshold+100), opb2.Outcomes[workflowTestID].LastSeenAt)
	assert.Equal(t, uint64(100), opb2.Outcomes[workflowTestID2].LastSeenAt)
	assert.Zero(t, opb2.Outcomes[workflowTestID3]) // This outcome was pruned
}

func TestReportPlugin_Outcome_ShouldReturnMedianTimestamp(t *testing.T) {
	lggr := logger.Test(t)
	s := requests.NewStore()
	mcap := &mockCapability{
		aggregator: &aggregator{},
		encoder:    &enc{},
		registeredWorkflows: map[string]bool{
			workflowTestID:  true,
			workflowTestID2: true,
		},
	}
	rp, err := newReportingPlugin(s, mcap, defaultBatchSize, ocr3types.ReportingPluginConfig{}, defaultOutcomePruningThreshold, lggr)
	require.NoError(t, err)

	weid := uuid.New().String()
	wowner := uuid.New().String()
	id := &pbtypes.Id{
		WorkflowExecutionId: weid,
		WorkflowId:          workflowTestID,
		WorkflowOwner:       wowner,
		WorkflowName:        workflowTestName,
		ReportId:            reportTestID,
	}
	id2 := &pbtypes.Id{
		WorkflowExecutionId: weid,
		WorkflowId:          workflowTestID2,
		WorkflowOwner:       wowner,
		WorkflowName:        workflowTestName,
		ReportId:            reportTestID,
	}
	id3 := &pbtypes.Id{
		WorkflowExecutionId: weid,
		WorkflowId:          workflowTestID3,
		WorkflowOwner:       wowner,
		WorkflowName:        workflowTestName,
		ReportId:            reportTestID,
	}
	q := &pbtypes.Query{
		Ids: []*pbtypes.Id{id, id2, id3},
	}
	qb, err := proto.Marshal(q)
	require.NoError(t, err)
	o, err := values.NewList([]any{"hello"})
	require.NoError(t, err)
	time1 := time.Now().Add(time.Second * 1)
	time2 := time.Now().Add(time.Second * 2)
	time3 := time.Now().Add(time.Second * 3)
	obs := &pbtypes.Observations{
		Observations: []*pbtypes.Observation{
			{
				Id:           id,
				Observations: values.Proto(o).GetListValue(),
			},
			{
				Id:           id2,
				Observations: values.Proto(o).GetListValue(),
			},
			{
				Id:           id3,
				Observations: values.Proto(o).GetListValue(),
			},
		},
		RegisteredWorkflowIds: []string{workflowTestID, workflowTestID2},
		Timestamp:             timestamppb.New(time1),
	}
	obs2 := &pbtypes.Observations{
		Observations: []*pbtypes.Observation{
			{
				Id:           id,
				Observations: values.Proto(o).GetListValue(),
			},
			{
				Id:           id2,
				Observations: values.Proto(o).GetListValue(),
			},
			{
				Id:           id3,
				Observations: values.Proto(o).GetListValue(),
			},
		},
		RegisteredWorkflowIds: []string{workflowTestID},
		Timestamp:             timestamppb.New(time2),
	}
	obs3 := &pbtypes.Observations{
		Observations: []*pbtypes.Observation{
			{
				Id:           id,
				Observations: values.Proto(o).GetListValue(),
			},
			{
				Id:           id2,
				Observations: values.Proto(o).GetListValue(),
			},
			{
				Id:           id3,
				Observations: values.Proto(o).GetListValue(),
			},
		},
		RegisteredWorkflowIds: []string{workflowTestID},
		Timestamp:             timestamppb.New(time3),
	}

	rawObs, err := proto.Marshal(obs)
	require.NoError(t, err)
	rawObs2, err := proto.Marshal(obs2)
	require.NoError(t, err)
	rawObs3, err := proto.Marshal(obs3)
	require.NoError(t, err)
	aos := []types.AttributedObservation{
		{
			Observation: rawObs,
			Observer:    commontypes.OracleID(1),
		},
		{
			Observation: rawObs2,
			Observer:    commontypes.OracleID(2),
		},
		{
			Observation: rawObs3,
			Observer:    commontypes.OracleID(3),
		},
	}

	outcome, err := rp.Outcome(ocr3types.OutcomeContext{SeqNr: 100}, qb, aos)
	require.NoError(t, err)
	opb1 := &pbtypes.Outcome{}
	err = proto.Unmarshal(outcome, opb1)
	require.NoError(t, err)

	assert.Equal(t, timestamppb.New(time2), opb1.Outcomes[workflowTestID].Timestamp)
}

func TestReportPlugin_Outcome_ShouldReturnOverriddenEncoder(t *testing.T) {
	lggr := logger.Test(t)
	s := requests.NewStore()
	mcap := &mockCapability{
		aggregator: &aggregator{},
		encoder:    &enc{},
		registeredWorkflows: map[string]bool{
			workflowTestID:  true,
			workflowTestID2: true,
		},
	}
	rp, err := newReportingPlugin(s, mcap, defaultBatchSize, ocr3types.ReportingPluginConfig{F: 1}, defaultOutcomePruningThreshold, lggr)
	require.NoError(t, err)

	wowner := uuid.New().String()
	id := &pbtypes.Id{
		WorkflowExecutionId: uuid.New().String(),
		WorkflowId:          workflowTestID,
		WorkflowOwner:       wowner,
		WorkflowName:        workflowTestName,
		ReportId:            reportTestID,
	}
	id2 := &pbtypes.Id{
		WorkflowExecutionId: uuid.New().String(),
		WorkflowId:          workflowTestID2,
		WorkflowOwner:       wowner,
		WorkflowName:        workflowTestName,
		ReportId:            reportTestID,
	}
	id3 := &pbtypes.Id{
		WorkflowExecutionId: uuid.New().String(),
		WorkflowId:          workflowTestID3,
		WorkflowOwner:       wowner,
		WorkflowName:        workflowTestName,
		ReportId:            reportTestID,
	}
	q := &pbtypes.Query{
		Ids: []*pbtypes.Id{id, id2, id3},
	}
	qb, err := proto.Marshal(q)
	require.NoError(t, err)
	o, err := values.NewList([]any{"hello"})
	require.NoError(t, err)
	time1 := time.Now().Add(time.Second * 1)
	time2 := time.Now().Add(time.Second * 2)
	time3 := time.Now().Add(time.Second * 3)
	m, err := values.NewMap(map[string]any{"foo": "bar"})
	require.NoError(t, err)
	mc := values.ProtoMap(m)
	obs := &pbtypes.Observations{
		Observations: []*pbtypes.Observation{
			{
				Id:                      id,
				Observations:            values.Proto(o).GetListValue(),
				OverriddenEncoderName:   "evm",
				OverriddenEncoderConfig: mc,
			},
			{
				Id:                      id2,
				Observations:            values.Proto(o).GetListValue(),
				OverriddenEncoderName:   "evm",
				OverriddenEncoderConfig: mc,
			},
			{
				Id:           id3,
				Observations: values.Proto(o).GetListValue(),
			},
		},
		RegisteredWorkflowIds: []string{workflowTestID, workflowTestID2},
		Timestamp:             timestamppb.New(time1),
	}
	obs2 := &pbtypes.Observations{
		Observations: []*pbtypes.Observation{
			{
				Id:                      id,
				Observations:            values.Proto(o).GetListValue(),
				OverriddenEncoderName:   "evm",
				OverriddenEncoderConfig: mc,
			},
			{
				Id:                      id2,
				Observations:            values.Proto(o).GetListValue(),
				OverriddenEncoderName:   "evm",
				OverriddenEncoderConfig: mc,
			},
			{
				Id:           id3,
				Observations: values.Proto(o).GetListValue(),
			},
		},
		RegisteredWorkflowIds: []string{workflowTestID},
		Timestamp:             timestamppb.New(time2),
	}
	obs3 := &pbtypes.Observations{
		Observations: []*pbtypes.Observation{
			{
				Id:                      id,
				Observations:            values.Proto(o).GetListValue(),
				OverriddenEncoderName:   "evm",
				OverriddenEncoderConfig: mc,
			},
			{
				Id:                      id2,
				Observations:            values.Proto(o).GetListValue(),
				OverriddenEncoderName:   "solana",
				OverriddenEncoderConfig: mc,
			},
			{
				Id:           id3,
				Observations: values.Proto(o).GetListValue(),
			},
		},
		RegisteredWorkflowIds: []string{workflowTestID},
		Timestamp:             timestamppb.New(time3),
	}

	rawObs, err := proto.Marshal(obs)
	require.NoError(t, err)
	rawObs2, err := proto.Marshal(obs2)
	require.NoError(t, err)
	rawObs3, err := proto.Marshal(obs3)
	require.NoError(t, err)
	aos := []types.AttributedObservation{
		{
			Observation: rawObs,
			Observer:    commontypes.OracleID(1),
		},
		{
			Observation: rawObs2,
			Observer:    commontypes.OracleID(2),
		},
		{
			Observation: rawObs3,
			Observer:    commontypes.OracleID(3),
		},
	}

	outcome, err := rp.Outcome(ocr3types.OutcomeContext{SeqNr: 100}, qb, aos)
	require.NoError(t, err)
	opb1 := &pbtypes.Outcome{}
	err = proto.Unmarshal(outcome, opb1)
	require.NoError(t, err)

	assert.Equal(t, opb1.Outcomes[workflowTestID].EncoderName, "evm")
	ec, err := values.FromMapValueProto(opb1.Outcomes[workflowTestID].EncoderConfig)
	assert.Equal(t, ec, m)

	// No consensus on outcome 2
	assert.Equal(t, opb1.Outcomes[workflowTestID2].EncoderName, "")
	assert.Nil(t, opb1.Outcomes[workflowTestID2].EncoderConfig)

	// Outcome 3 doesn't set the encoder
	assert.Equal(t, opb1.Outcomes[workflowTestID3].EncoderName, "")
	assert.Nil(t, opb1.Outcomes[workflowTestID3].EncoderConfig)
}
