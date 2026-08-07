package ocr3

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/libocr/commontypes"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"

	pbtypes "github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/types"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/requests"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-protos/cre/go/values"
)

type forgeAggregator struct{}

func (forgeAggregator) Aggregate(_ logger.Logger, _ *pbtypes.AggregationOutcome, obs map[commontypes.OracleID][]values.Value, _ int) (*pbtypes.AggregationOutcome, error) {
	var price string
	for _, vals := range obs {
		if len(vals) > 0 {
			s, _ := vals[0].Unwrap()
			price, _ = s.(string)
			break
		}
	}
	nm, err := values.NewMap(map[string]any{"price": price})
	if err != nil {
		return nil, err
	}
	return &pbtypes.AggregationOutcome{
		EncodableOutcome: values.Proto(nm).GetMapValue(),
		ShouldReport:     true,
	}, nil
}

type forgeCapability struct {
	aggWfID string
	encWfID string
}

func (c *forgeCapability) GetAggregator(workflowID string) (pbtypes.Aggregator, error) {
	c.aggWfID = workflowID
	return forgeAggregator{}, nil
}

func (c *forgeCapability) GetEncoderByWorkflowID(workflowID string) (pbtypes.Encoder, error) {
	c.encWfID = workflowID
	return &enc{}, nil
}

func (c *forgeCapability) GetEncoderByName(string, *values.Map) (pbtypes.Encoder, error) {
	return &enc{}, nil
}

func (c *forgeCapability) GetRegisteredWorkflowsIDs() []string { return nil }
func (c *forgeCapability) UnregisterWorkflowID(string)         {}

func TestReportingPlugin_LeaderQueryIdentityForgery(t *testing.T) {
	const (
		sharedExecID = "exec-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

		attackerWfID    = "attacker-workflow-id"
		attackerWfOwner = "attacker-owner"
		attackerWfName  = "attacker-name"
		attackerReport  = "0002"
		attackerKey     = "attacker-key"

		victimWfID    = "victim-workflow-id"
		victimWfOwner = "victim-owner"
		victimWfName  = "victim-name"
		victimReport  = "0001"
		victimKey     = "victim-key"
	)

	obsVals, err := values.NewList([]any{"attacker-price"})
	require.NoError(t, err)

	mkObs := func(id *pbtypes.Id) []types.AttributedObservation {
		obs := &pbtypes.Observations{
			Observations: []*pbtypes.Observation{{
				Id:           id,
				Observations: values.Proto(obsVals).GetListValue(),
			}},
			RegisteredWorkflowIds: []string{attackerWfID, victimWfID},
		}
		raw, err := proto.Marshal(obs)
		require.NoError(t, err)
		return []types.AttributedObservation{
			{Observation: raw, Observer: commontypes.OracleID(0)},
			{Observation: raw, Observer: commontypes.OracleID(1)},
			{Observation: raw, Observer: commontypes.OracleID(2)},
		}
	}

	runRound := func(t *testing.T, queryID, obsID *pbtypes.Id) (*pbtypes.Outcome, []ocr3types.ReportPlus[[]byte], *forgeCapability) {
		t.Helper()
		cap := &forgeCapability{}
		rp, err := NewReportingPlugin(
			requests.NewStore[*ReportRequest](),
			cap,
			defaultBatchSize,
			ocr3types.ReportingPluginConfig{F: 1},
			defaultLimits(),
			logger.Test(t),
		)
		require.NoError(t, err)

		queryBytes, err := proto.Marshal(&pbtypes.Query{Ids: []*pbtypes.Id{queryID}})
		require.NoError(t, err)

		outcomeBytes, err := rp.Outcome(t.Context(), ocr3types.OutcomeContext{}, queryBytes, mkObs(obsID))
		require.NoError(t, err)

		outcome := &pbtypes.Outcome{}
		require.NoError(t, proto.Unmarshal(outcomeBytes, outcome))

		reports, err := rp.Reports(t.Context(), 1, outcomeBytes)
		require.NoError(t, err)

		return outcome, reports, cap
	}

	attackerID := &pbtypes.Id{
		WorkflowExecutionId: sharedExecID,
		WorkflowId:          attackerWfID,
		WorkflowOwner:       attackerWfOwner,
		WorkflowName:        attackerWfName,
		ReportId:            attackerReport,
		KeyId:               attackerKey,
	}
	victimID := &pbtypes.Id{
		WorkflowExecutionId: sharedExecID,
		WorkflowId:          victimWfID,
		WorkflowOwner:       victimWfOwner,
		WorkflowName:        victimWfName,
		ReportId:            victimReport,
		KeyId:               victimKey,
	}

	t.Run("honest leader: query identity matches observation identity", func(t *testing.T) {
		outcome, _, cap := runRound(t, attackerID, attackerID)

		require.Len(t, outcome.CurrentReports, 1)
		rpt := outcome.CurrentReports[0]
		require.Equal(t, attackerWfID, rpt.Id.WorkflowId)
		require.Equal(t, attackerWfOwner, rpt.Id.WorkflowOwner)
		require.Equal(t, attackerKey, rpt.Id.KeyId)
		require.Equal(t, attackerWfID, cap.aggWfID, "aggregator selected for attacker's workflow")
		require.Equal(t, attackerWfID, cap.encWfID, "encoder selected for attacker's workflow")
	})

	t.Run("byzantine leader: query carries victim identity, observations carry attacker identity", func(t *testing.T) {
		outcome, _, _ := runRound(t, victimID, attackerID)

		require.Empty(t, outcome.CurrentReports,
			"forged report must not be produced when observation identity does not match query identity")
	})
}
