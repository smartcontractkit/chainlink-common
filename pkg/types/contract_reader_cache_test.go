package types_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

var (
	testAddress  = "0x42"
	testContract = "anyContract"
	testReadName = "readName"
	testBinding  = types.BoundContract{
		Name:    testContract,
		Address: testAddress,
	}
)

func TestCachedContractReader_DefaultFallthrough(t *testing.T) {
	t.Parallel()

	mcr := newTestReader()
	crCache := types.NewCachedContractReader(mcr)
	ctx := tests.Context(t)

	var output int64

	err := crCache.GetLatestValue(
		ctx,
		testBinding.ReadIdentifier(testReadName),
		primitives.Finalized,
		map[string]any{},
		&output,
	)

	require.ErrorIs(t, err, types.ErrInvalidEncoding)
}

func TestCachedContractReader_FallthroughOnStale(t *testing.T) {
	t.Parallel()

	mcr := newTestReader()
	crCache := types.NewCachedContractReader(mcr)
	ctx := tests.Context(t)

	strategy := types.FallthroughOnStale{
		Staleness: 5 * time.Second,
	}

	readIdentifier := testBinding.ReadIdentifier(testReadName)

	// setting the cache strategy should not return an error
	require.NoError(t, crCache.SetCacheStrategy(readIdentifier, primitives.Finalized, strategy))

	callFunc := func(t *testing.T, val map[string]any, callCount int) {
		t.Helper()

		output := make(map[string]any)
		params := map[string]any{"param1": "value1"}

		require.NoError(t, crCache.GetLatestValue(ctx, readIdentifier, primitives.Finalized, &params, &output))
		require.NoError(t, crCache.GetLatestValue(ctx, readIdentifier, primitives.Finalized, &params, &output))
		require.Equal(t, callCount, mcr.CallCount())

		assert.Equal(t, val, output)
	}

	callFunc(t, map[string]any{"ret1": "latestValue1"}, 1)
	time.Sleep(strategy.Staleness)
	callFunc(t, map[string]any{"ret1": "latestValue1"}, 2)
}

func TestCachedContractReader_PollAndCache(t *testing.T) {
	t.Parallel()

	mcr := newTestReader()
	crCache := types.NewCachedContractReader(mcr)
	ctx := tests.Context(t)

	strategy := types.PollAndCache{
		TriggerAndCache: types.TriggerAndCache{
			Params: &map[string]any{"param1": "value1"},
			Value:  reflect.New(reflect.TypeOf(map[string]any{})).Interface(),
		},
		Interval: 5 * time.Second,
	}

	require.NoError(t, crCache.SetCacheStrategy(testBinding.ReadIdentifier(testReadName), primitives.Finalized, strategy))

	t.Cleanup(func() {
		require.NoError(t, crCache.Close())
	})

	require.NoError(t, crCache.Start(ctx))

	// wait for a single polling interval
	time.Sleep(time.Second)
	callAndAssert(ctx, t, crCache, mcr, 1, map[string]any{"ret1": "latestValue1"}, strategy.Params)

	// request with different params falls through
	callAndAssert(ctx, t, crCache, mcr, 2, map[string]any{"ret2": "latestValue2"}, &map[string]any{"param2": "value2"})
	time.Sleep(strategy.Interval)

	// after polling interval
	callAndAssert(ctx, t, crCache, mcr, 2, map[string]any{"ret1": "latestValue1"}, strategy.Params)
}

func callAndAssert(
	ctx context.Context,
	t *testing.T,
	crCache *types.CachedContractReader,
	mcr *staticContractReader,
	expectedCount int,
	expected any,
	params any,
) {
	t.Helper()

	output := make(map[string]any)

	require.NoError(t, crCache.GetLatestValue(ctx, testBinding.ReadIdentifier(testReadName), primitives.Finalized, params, &output))
	assert.Equal(t, expected, output)
	assert.Equal(t, expectedCount, mcr.CallCount())
}

func newTestReader() *staticContractReader {
	// ContractReader is a static implementation of [types.ContractReader], [testtypes.Evaluator] and [types.PluginProvider
	// it is used for testing the [types.PluginProvider] interface
	return &staticContractReader{
		address:        testAddress,
		contractName:   testContract,
		contractMethod: testReadName,
		latestValue:    map[string]any{"ret1": "latestValue1", "ret2": "latestValue2"},
		params:         map[string]any{"param1": "value1", "param2": "value2"},
	}
}

// staticContractReader is a static implementation of ContractReaderTester
type staticContractReader struct {
	types.UnimplementedContractReader

	address        string
	contractName   string
	contractMethod string
	latestValue    map[string]any
	params         map[string]any

	callCount int
}

var _ types.ContractReader = &staticContractReader{}

func (c staticContractReader) Start(_ context.Context) error { return nil }

func (c staticContractReader) Close() error { return nil }

func (c staticContractReader) Ready() error { panic("unimplemented") }

func (c staticContractReader) Name() string { panic("unimplemented") }

func (c staticContractReader) HealthReport() map[string]error { panic("unimplemented") }

func (c staticContractReader) Bind(_ context.Context, _ []types.BoundContract) error {
	return nil
}

func (c staticContractReader) Unbind(_ context.Context, _ []types.BoundContract) error {
	return nil
}

func (c *staticContractReader) GetLatestValue(_ context.Context, readName string, _ primitives.ConfidenceLevel, params, returnVal any) error {
	c.callCount++
	comp := types.BoundContract{
		Address: c.address,
		Name:    c.contractName,
	}.ReadIdentifier(c.contractMethod)

	if !assert.ObjectsAreEqual(readName, comp) {
		return fmt.Errorf("%w: expected report context %v but got %v", types.ErrInvalidType, comp, readName)
	}

	//gotParams, ok := params.(*map[string]string)
	gotParams, ok := params.(*map[string]any)
	if !ok {
		return fmt.Errorf("%w: Invalid parameter type received in GetLatestValue. Expected %T but received %T", types.ErrInvalidEncoding, c.params, params)
	}

	ret, ok := returnVal.(*map[string]any)
	if !ok {
		return fmt.Errorf("%w: Wrong type passed for retVal param to GetLatestValue impl (expected %T instead of %T", types.ErrInvalidType, c.latestValue, returnVal)
	}

	if val, ok := (*gotParams)["param1"]; ok {
		if val != c.params["param1"] {
			return fmt.Errorf("%w: Wrong params value received in GetLatestValue. Expected %v but received %v", types.ErrInvalidEncoding, c.params, *gotParams)
		}

		*ret = map[string]any{"ret1": c.latestValue["ret1"]}
	}

	if val, ok := (*gotParams)["param2"]; ok {
		if val != c.params["param2"] {
			return fmt.Errorf("%w: Wrong params value received in GetLatestValue. Expected %v but received %v", types.ErrInvalidEncoding, c.params, *gotParams)
		}

		*ret = map[string]any{"ret2": c.latestValue["ret2"]}
	}

	return nil
}

func (c *staticContractReader) BatchGetLatestValues(_ context.Context, _ types.BatchGetLatestValuesRequest) (types.BatchGetLatestValuesResult, error) {
	return nil, nil
}

func (c staticContractReader) QueryKey(_ context.Context, _ types.BoundContract, _ query.KeyFilter, _ query.LimitAndSort, _ any) ([]types.Sequence, error) {
	return nil, nil
}

func (c *staticContractReader) CallCount() int {
	return c.callCount
}

func (c staticContractReader) Evaluate(ctx context.Context, cr types.ContractReader) error {
	gotLatestValue := make(map[string]any)

	if err := cr.GetLatestValue(
		ctx,
		types.BoundContract{
			Address: c.address,
			Name:    c.contractName,
		}.ReadIdentifier(c.contractMethod),
		primitives.Unconfirmed,
		&c.params,
		&gotLatestValue,
	); err != nil {
		return fmt.Errorf("failed to call GetLatestValue(): %w", err)
	}

	if !assert.ObjectsAreEqual(gotLatestValue, c.latestValue) {
		return fmt.Errorf("GetLatestValue: expected %v but got %v", c.latestValue, gotLatestValue)
	}

	return nil
}
