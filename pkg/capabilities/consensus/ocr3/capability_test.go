package ocr3

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jonboulle/clockwork"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/requests"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/types"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/utils"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

const workflowTestID = "consensus-workflow-test-id-1"
const workflowTestID2 = "consensus-workflow-test-id-2"
const workflowTestID3 = "consensus-workflow-test-id-3"
const workflowExecutionTestID = "consensus-workflow-execution-test-id-1"
const workflowTestName = "consensus-workflow-test-name-1"
const reportTestID = "rep-id-1"

type mockAggregator struct {
	types.Aggregator
}

func mockAggregatorFactory(_ string, _ values.Map, _ logger.Logger) (types.Aggregator, error) {
	return &mockAggregator{}, nil
}

type encoder struct {
	types.Encoder
}

func mockEncoderFactory(_ string, _ *values.Map, _ logger.Logger) (types.Encoder, error) {
	return &encoder{}, nil
}

func TestOCR3Capability_Schema(t *testing.T) {
	n := time.Now()
	fc := clockwork.NewFakeClockAt(n)
	lggr := logger.Nop()

	s := requests.NewStore()

	cp := newCapability(s, fc, 1*time.Second, mockAggregatorFactory, mockEncoderFactory, lggr, 10)
	schema, err := cp.Schema()
	require.NoError(t, err)

	var shouldUpdate = false
	if shouldUpdate {
		err = os.WriteFile("./testdata/fixtures/capability/schema.json", []byte(schema), 0600)
		require.NoError(t, err)
	}

	fixture, err := os.ReadFile("./testdata/fixtures/capability/schema.json")
	require.NoError(t, err)

	utils.AssertJSONEqual(t, fixture, []byte(schema))
}

func TestOCR3Capability(t *testing.T) {
	cases := []struct {
		name              string
		aggregationMethod string
	}{
		{
			name:              "success - aggregation_method data_feeds",
			aggregationMethod: "data_feeds",
		},
		{
			name:              "success - aggregation_method reduce",
			aggregationMethod: "reduce",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			n := time.Now()
			fc := clockwork.NewFakeClockAt(n)
			lggr := logger.Test(t)

			ctx := tests.Context(t)

			s := requests.NewStore()

			cp := newCapability(s, fc, 1*time.Second, mockAggregatorFactory, mockEncoderFactory, lggr, 10)
			require.NoError(t, cp.Start(ctx))

			config, err := values.NewMap(
				map[string]any{
					"aggregation_method": tt.aggregationMethod,
					"aggregation_config": map[string]any{},
					"encoder_config":     map[string]any{},
					"encoder":            "evm",
					"report_id":          "ffff",
					"key_id":             "evm",
				},
			)
			require.NoError(t, err)

			ethUsdValStr := "1.123456"
			ethUsdValue, err := decimal.NewFromString(ethUsdValStr)
			require.NoError(t, err)
			observationKey := "ETH_USD"
			obs := []any{map[string]any{observationKey: ethUsdValue}}
			inputs, err := values.NewMap(map[string]any{"observations": obs})
			require.NoError(t, err)

			executeReq := capabilities.CapabilityRequest{
				Metadata: capabilities.RequestMetadata{
					WorkflowID:          workflowTestID,
					WorkflowExecutionID: workflowExecutionTestID,
				},
				Config: config,
				Inputs: inputs,
			}

			respCh := executeAsync(ctx, executeReq, cp.Execute)

			obsv, err := values.NewList(obs)
			require.NoError(t, err)

			// Mock the oracle returning a response
			mresp, err := values.NewMap(map[string]any{"observations": obsv})
			cp.reqHandler.SendResponse(ctx, requests.Response{
				Value:               mresp,
				WorkflowExecutionID: workflowExecutionTestID,
			})
			require.NoError(t, err)

			resp := <-respCh
			assert.NoError(t, resp.Err)

			assert.Equal(t, mresp, resp.Value)
		})
	}
}

func TestOCR3Capability_Eviction(t *testing.T) {
	n := time.Now()
	fc := clockwork.NewFakeClockAt(n)
	lggr := logger.Test(t)

	ctx := tests.Context(t)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	rea := time.Second
	s := requests.NewStore()
	cp := newCapability(s, fc, rea, mockAggregatorFactory, mockEncoderFactory, lggr, 10)
	require.NoError(t, cp.Start(ctx))

	config, err := values.NewMap(
		map[string]any{
			"aggregation_method": "data_feeds",
			"aggregation_config": map[string]any{},
			"encoder_config":     map[string]any{},
			"encoder":            "evm",
			"report_id":          "aaaa",
			"key_id":             "evm",
		},
	)
	require.NoError(t, err)

	ethUsdValue, err := decimal.NewFromString("1.123456")
	require.NoError(t, err)
	inputs, err := values.NewMap(map[string]any{"observations": []any{map[string]any{"ETH_USD": ethUsdValue}}})
	require.NoError(t, err)

	rid := uuid.New().String()
	executeReq := capabilities.CapabilityRequest{
		Metadata: capabilities.RequestMetadata{
			WorkflowID:          workflowTestID,
			WorkflowExecutionID: rid,
		},
		Config: config,
		Inputs: inputs,
	}

	done := make(chan struct{})
	t.Cleanup(func() { <-done })
	go func() {
		defer close(done)
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				fc.Advance(1 * time.Hour)
			}
		}
	}()

	respCh := executeAsync(ctx, executeReq, cp.Execute)

	resp := <-respCh
	assert.ErrorContains(t, resp.Err, "timeout exceeded: could not process request before expiry")

	request := s.Get(rid)
	assert.Nil(t, request)

	assert.Nil(t, err)
}

func TestOCR3Capability_EvictionUsingConfig(t *testing.T) {
	n := time.Now()
	fc := clockwork.NewFakeClockAt(n)
	lggr := logger.Test(t)

	ctx := tests.Context(t)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	// This is the default expired at
	rea := time.Hour
	s := requests.NewStore()
	cp := newCapability(s, fc, rea, mockAggregatorFactory, mockEncoderFactory, lggr, 10)
	require.NoError(t, cp.Start(ctx))

	config, err := values.NewMap(
		map[string]any{
			"aggregation_method": "data_feeds",
			"aggregation_config": map[string]any{},
			"encoder_config":     map[string]any{},
			"encoder":            "evm",
			"report_id":          "aaaa",
			"key_id":             "evm",
			"request_timeout_ms": 10000,
		},
	)
	require.NoError(t, err)

	ethUsdValue, err := decimal.NewFromString("1.123456")
	require.NoError(t, err)
	inputs, err := values.NewMap(map[string]any{"observations": []any{map[string]any{"ETH_USD": ethUsdValue}}})
	require.NoError(t, err)

	rid := uuid.New().String()
	executeReq := capabilities.CapabilityRequest{
		Metadata: capabilities.RequestMetadata{
			WorkflowID:          workflowTestID,
			WorkflowExecutionID: rid,
		},
		Config: config,
		Inputs: inputs,
	}

	// 1 minute is more than the config timeout we provided, but less than
	// the hardcoded timeout.
	done := make(chan struct{})
	t.Cleanup(func() { <-done })
	go func() {
		defer close(done)
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// 1 minute is more than the config timeout we provided, but less than
				// the hardcoded timeout.
				fc.Advance(1 * time.Minute)
			}
		}
	}()

	_, err = cp.Execute(ctx, executeReq)

	assert.ErrorContains(t, err, "timeout exceeded: could not process request before expiry")

	reqs := s.GetByIDs([]string{rid})

	assert.Equal(t, 0, len(reqs))
}

func TestOCR3Capability_Registration(t *testing.T) {
	n := time.Now()
	fc := clockwork.NewFakeClockAt(n)
	lggr := logger.Test(t)

	ctx := tests.Context(t)
	s := requests.NewStore()
	cp := newCapability(s, fc, 1*time.Second, mockAggregatorFactory, mockEncoderFactory, lggr, 10)
	require.NoError(t, cp.Start(ctx))

	config, err := values.NewMap(map[string]any{
		"aggregation_method": "data_feeds",
		"aggregation_config": map[string]any{},
		"encoder":            "",
		"encoder_config":     map[string]any{},
		"report_id":          "000f",
		"key_id":             "evm",
	})
	require.NoError(t, err)

	registerReq := capabilities.RegisterToWorkflowRequest{
		Metadata: capabilities.RegistrationMetadata{
			WorkflowID: workflowTestID,
		},
		Config: config,
	}

	err = cp.RegisterToWorkflow(ctx, registerReq)
	require.NoError(t, err)

	agg, err := cp.getAggregator(workflowTestID)
	require.NoError(t, err)
	assert.NotNil(t, agg)

	unregisterReq := capabilities.UnregisterFromWorkflowRequest{
		Metadata: capabilities.RegistrationMetadata{
			WorkflowID: workflowTestID,
		},
	}

	err = cp.UnregisterFromWorkflow(ctx, unregisterReq)
	require.NoError(t, err)

	_, err = cp.getAggregator(workflowTestID)
	assert.ErrorContains(t, err, "no aggregator found for")
}

func TestOCR3Capability_ValidateConfig(t *testing.T) {
	n := time.Now()
	fc := clockwork.NewFakeClockAt(n)
	lggr := logger.Test(t)

	s := requests.NewStore()

	o := newCapability(s, fc, 1*time.Second, mockAggregatorFactory, mockEncoderFactory, lggr, 10)

	t.Run("ValidConfig", func(t *testing.T) {
		config, err := values.NewMap(map[string]any{
			"aggregation_method": "data_feeds",
			"aggregation_config": map[string]any{},
			"encoder":            "",
			"encoder_config":     map[string]any{},
			"report_id":          "aaaa",
			"key_id":             "evm",
		})
		require.NoError(t, err)

		c, err := o.ValidateConfig(config)
		require.NoError(t, err)
		require.NotNil(t, c)
	})

	t.Run("InvalidConfig null", func(t *testing.T) {
		config, err := values.NewMap(map[string]any{
			"aggregation_method": "data_feeds",
			"report_id":          "aaaa",
			"key_id":             "evm",
		})
		require.NoError(t, err)

		c, err := o.ValidateConfig(config)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected object, but got null") // taken from the error json schema error message
		require.Nil(t, c)
	})

	t.Run("InvalidConfig illegal report_id", func(t *testing.T) {
		config, err := values.NewMap(map[string]any{
			"aggregation_method": "data_feeds",
			"aggregation_config": map[string]any{},
			"encoder":            "",
			"encoder_config":     map[string]any{},
			"report_id":          "aa",
			"key_id":             "evm",
		})
		require.NoError(t, err)

		c, err := o.ValidateConfig(config)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "does not match pattern") // taken from the error json schema error message
		require.Nil(t, c)
	})

	t.Run("InvalidConfig no key_id", func(t *testing.T) {
		config, err := values.NewMap(map[string]any{
			"aggregation_method": "data_feeds",
			"aggregation_config": map[string]any{},
			"encoder":            "",
			"encoder_config":     map[string]any{},
			"report_id":          "aaaa",
		})
		require.NoError(t, err)

		c, err := o.ValidateConfig(config)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "missing properties: 'key_id'") // taken from the error json schema error message
		require.Nil(t, c)
	})
}

func TestOCR3Capability_RespondsToLateRequest(t *testing.T) {
	n := time.Now()
	fc := clockwork.NewFakeClockAt(n)
	lggr := logger.Test(t)

	ctx := tests.Context(t)

	s := requests.NewStore()

	cp := newCapability(s, fc, 1*time.Second, mockAggregatorFactory, mockEncoderFactory, lggr, 10)
	require.NoError(t, cp.Start(ctx))

	config, err := values.NewMap(
		map[string]any{
			"aggregation_method": "data_feeds",
			"aggregation_config": map[string]any{},
			"encoder_config":     map[string]any{},
			"encoder":            "evm",
			"report_id":          "ffff",
			"key_id":             "evm",
		},
	)
	require.NoError(t, err)

	ethUsdValStr := "1.123456"
	ethUsdValue, err := decimal.NewFromString(ethUsdValStr)
	require.NoError(t, err)
	observationKey := "ETH_USD"
	obs := map[string]any{observationKey: ethUsdValue}
	inputs, err := values.NewMap(map[string]any{"observations": []any{obs}})
	require.NoError(t, err)

	obsv, err := values.NewMap(obs)
	require.NoError(t, err)

	// Mock the oracle returning a response prior to the request being sent
	cp.reqHandler.SendResponse(ctx, requests.Response{
		Value:               obsv,
		WorkflowExecutionID: workflowExecutionTestID,
	})
	require.NoError(t, err)

	executeReq := capabilities.CapabilityRequest{
		Metadata: capabilities.RequestMetadata{
			WorkflowID:          workflowTestID,
			WorkflowExecutionID: workflowExecutionTestID,
		},
		Config: config,
		Inputs: inputs,
	}
	response, err := cp.Execute(ctx, executeReq)
	require.NoError(t, err)

	expectedCapabilityResponse := capabilities.CapabilityResponse{
		Value: obsv,
	}

	assert.Equal(t, expectedCapabilityResponse, response)
}

func TestOCR3Capability_RespondingToLateRequestDoesNotBlockOnSlowResponseConsumer(t *testing.T) {
	n := time.Now()
	fc := clockwork.NewFakeClockAt(n)
	lggr := logger.Test(t)

	ctx := tests.Context(t)

	s := requests.NewStore()

	cp := newCapability(s, fc, 1*time.Second, mockAggregatorFactory, mockEncoderFactory, lggr, 0)
	require.NoError(t, cp.Start(ctx))

	config, err := values.NewMap(
		map[string]any{
			"aggregation_method": "data_feeds",
			"aggregation_config": map[string]any{},
			"encoder_config":     map[string]any{},
			"encoder":            "evm",
			"report_id":          "ffff",
			"key_id":             "evm",
		},
	)
	require.NoError(t, err)

	ethUsdValStr := "1.123456"
	ethUsdValue, err := decimal.NewFromString(ethUsdValStr)
	require.NoError(t, err)
	observationKey := "ETH_USD"
	obs := map[string]any{observationKey: ethUsdValue}
	inputs, err := values.NewMap(map[string]any{"observations": []any{obs}})
	require.NoError(t, err)

	obsv, err := values.NewMap(obs)
	require.NoError(t, err)

	// Mock the oracle returning a response prior to the request being sent
	cp.reqHandler.SendResponse(ctx, requests.Response{
		Value:               obsv,
		WorkflowExecutionID: workflowExecutionTestID,
	})
	require.NoError(t, err)

	executeReq := capabilities.CapabilityRequest{
		Metadata: capabilities.RequestMetadata{
			WorkflowID:          workflowTestID,
			WorkflowExecutionID: workflowExecutionTestID,
		},
		Config: config,
		Inputs: inputs,
	}
	resp, err := cp.Execute(ctx, executeReq)
	require.NoError(t, err)

	expectedCapabilityResponse := capabilities.CapabilityResponse{
		Value: obsv,
	}

	assert.Equal(t, expectedCapabilityResponse, resp)
}

type asyncCapabilityResponse struct {
	capabilities.CapabilityResponse
	Err error
}

func executeAsync(ctx context.Context, request capabilities.CapabilityRequest, toExecute func(ctx context.Context, request capabilities.CapabilityRequest) (capabilities.CapabilityResponse, error)) <-chan asyncCapabilityResponse {
	respCh := make(chan asyncCapabilityResponse, 1)
	go func() {
		resp, err := toExecute(ctx, request)
		respCh <- asyncCapabilityResponse{CapabilityResponse: capabilities.CapabilityResponse{Value: resp.Value}, Err: err}
		close(respCh)
	}()

	return respCh
}
