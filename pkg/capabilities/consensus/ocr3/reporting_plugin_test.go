package ocr3

import (
	"context"
	"sort"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

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
	rp, err := newReportingPlugin(s, nil, batchSize, ocr3types.ReportingPluginConfig{}, lggr)
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
	rp, err := newReportingPlugin(s, nil, defaultBatchSize, ocr3types.ReportingPluginConfig{}, lggr)
	require.NoError(t, err)

	eid := uuid.New().String()
	wowner := uuid.New().String()

	err = s.Add(&requests.Request{
		WorkflowID:          workflowTestID,
		WorkflowExecutionID: eid,
		WorkflowOwner:       wowner,
		WorkflowName:        workflowTestName,
		ReportID:            reportTestId,
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
	gotResponse         *requests.Response
	aggregator          *aggregator
	encoder             *enc
	registeredWorkflows map[string]bool
}

func (mc *mockCapability) transmitResponse(ctx context.Context, resp *requests.Response) error {
	mc.gotResponse = resp
	return nil
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

func (mc *mockCapability) getEncoder(workflowID string) (pbtypes.Encoder, error) {
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
	rp, err := newReportingPlugin(s, mcap, defaultBatchSize, ocr3types.ReportingPluginConfig{}, lggr)
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
		ReportID:            reportTestId,
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
	assert.Equal(t, o, values.FromListValueProto(fo.Observations))
	expected := []string{workflowTestID, workflowTestID2}
	actual := obspb.RegisteredWorkflowIds
	sort.Slice(actual, func(i, j int) bool { return actual[i] < actual[j] })
	sort.Slice(expected, func(i, j int) bool { return expected[i] < expected[j] })
	assert.Equal(t, expected, actual)
}

func TestReportingPlugin_Observation_NoResults(t *testing.T) {
	ctx := tests.Context(t)
	lggr := logger.Test(t)
	s := requests.NewStore()
	mcap := &mockCapability{
		aggregator: &aggregator{},
		encoder:    &enc{},
	}
	rp, err := newReportingPlugin(s, mcap, defaultBatchSize, ocr3types.ReportingPluginConfig{}, lggr)
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
	cap := &mockCapability{
		aggregator: &aggregator{},
		encoder:    &enc{},
	}
	rp, err := newReportingPlugin(s, cap, defaultBatchSize, ocr3types.ReportingPluginConfig{}, lggr)
	require.NoError(t, err)

	weid := uuid.New().String()
	wowner := uuid.New().String()
	id := &pbtypes.Id{
		WorkflowExecutionId: weid,
		WorkflowId:          workflowTestID,
		WorkflowOwner:       wowner,
		WorkflowName:        workflowTestName,
		ReportId:            reportTestId,
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
	assert.EqualExportedValues(t, cr.Outcome, cap.aggregator.outcome)
	assert.EqualExportedValues(t, opb.Outcomes[workflowTestID], cap.aggregator.outcome)
}

func TestReportingPlugin_Reports_ShouldReportFalse(t *testing.T) {
	lggr := logger.Test(t)
	s := requests.NewStore()
	cap := &mockCapability{
		aggregator: &aggregator{},
		encoder:    &enc{},
	}
	rp, err := newReportingPlugin(s, cap, defaultBatchSize, ocr3types.ReportingPluginConfig{}, lggr)
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

func TestReportingPlugin_Reports_ShouldReportTrue(t *testing.T) {
	lggr := logger.Test(t)
	s := requests.NewStore()
	cap := &mockCapability{
		aggregator: &aggregator{},
		encoder:    &enc{},
	}
	rp, err := newReportingPlugin(s, cap, defaultBatchSize, ocr3types.ReportingPluginConfig{}, lggr)
	require.NoError(t, err)

	var sqNr uint64
	weid := uuid.New().String()
	wowner := uuid.New().String()
	donId := uint32(1)
	id := &pbtypes.Id{
		WorkflowExecutionId: weid,
		WorkflowId:          workflowTestID,
		WorkflowOwner:       wowner,
		WorkflowName:        workflowTestName,
		ReportId:            reportTestId,
		WorkflowDonId:       donId,
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
		"DONID":            donId,
		"DONConfigVersion": 0,
		"WorkflowID":       workflowTestID,
		"WorkflowName":     workflowTestName,
		"WorkflowOwner":    wowner,
		"ReportID":         reportTestId,
	})
	require.NoError(t, err)
	fp := values.FromProto(rep)
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
	cap := &mockCapability{
		aggregator: &aggregator{},
		encoder:    &enc{},
		registeredWorkflows: map[string]bool{
			workflowTestID:  true,
			workflowTestID2: true,
		},
	}
	rp, err := newReportingPlugin(s, cap, defaultBatchSize, ocr3types.ReportingPluginConfig{}, lggr)
	require.NoError(t, err)

	weid := uuid.New().String()
	wowner := uuid.New().String()
	id := &pbtypes.Id{
		WorkflowExecutionId: weid,
		WorkflowId:          workflowTestID,
		WorkflowOwner:       wowner,
		WorkflowName:        workflowTestName,
		ReportId:            reportTestId,
	}
	id2 := &pbtypes.Id{
		WorkflowExecutionId: weid,
		WorkflowId:          workflowTestID2,
		WorkflowOwner:       wowner,
		WorkflowName:        workflowTestName,
		ReportId:            reportTestId,
	}
	id3 := &pbtypes.Id{
		WorkflowExecutionId: weid,
		WorkflowId:          workflowTestID3,
		WorkflowOwner:       wowner,
		WorkflowName:        workflowTestName,
		ReportId:            reportTestId,
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

	outcome2, err := rp.Outcome(ocr3types.OutcomeContext{SeqNr: outcomePruningThreshold + 100, PreviousOutcome: outcome1}, qb, aos2)
	require.NoError(t, err)
	opb2 := &pbtypes.Outcome{}
	err = proto.Unmarshal(outcome2, opb2)
	require.NoError(t, err)

	assert.Equal(t, uint64(100), opb1.Outcomes[workflowTestID].LastSeenAt)
	assert.Equal(t, uint64(100), opb1.Outcomes[workflowTestID2].LastSeenAt)
	assert.Equal(t, uint64(0), opb1.Outcomes[workflowTestID3].LastSeenAt)
	assert.Equal(t, uint64(outcomePruningThreshold+100), opb2.Outcomes[workflowTestID].LastSeenAt)
	assert.Equal(t, uint64(100), opb2.Outcomes[workflowTestID2].LastSeenAt)
	assert.Zero(t, opb2.Outcomes[workflowTestID3]) // This outcome was pruned
}
