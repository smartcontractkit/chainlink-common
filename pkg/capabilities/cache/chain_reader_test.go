package cache_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/cache"
	"github.com/smartcontractkit/chainlink-common/pkg/types/mocks"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

type testParams struct {
	A int
	B string
}

var (
	testContract = "contract"
	testMethod   = "method"
)

func TestContractReader_DefaultFallthrough(t *testing.T) {
	t.Parallel()

	mcr := new(mocks.ContractReader)
	crCache := cache.NewChainReader(mcr)
	ctx := tests.Context(t)
	expectedError := errors.New("test error")

	mcr.On(
		"GetLatestValue",
		mock.Anything,
		testContract,
		testMethod,
		mock.AnythingOfType("testParams"),
		mock.AnythingOfType("*int64")).Return(expectedError)

	var output int64

	err := crCache.GetLatestValue(ctx, testContract, testMethod, testParams{}, &output)

	require.ErrorIs(t, err, expectedError)
}

func TestContractReader_FallthroughOnStale(t *testing.T) {
	t.Parallel()

	mcr := new(mocks.ContractReader)
	crCache := cache.NewChainReader(mcr)
	ctx := tests.Context(t)

	// use FallthroughOnStale cache strategy
	strategy := cache.FallthroughOnStale{
		Staleness: 5 * time.Second,
	}

	// setting the cache strategy should not return an error
	require.NoError(t, crCache.SetCacheStrategy(testContract, testMethod, strategy))

	callFunc := func(t *testing.T, val int64) {
		t.Helper()

		var output int64

		mcr.On(
			"GetLatestValue",
			mock.Anything,
			testContract,
			testMethod,
			mock.AnythingOfType("cache_test.testParams"),
			mock.AnythingOfType("*int64")).
			Run(func(args mock.Arguments) {
				reflect.Indirect(reflect.ValueOf(args.Get(4))).Set(reflect.ValueOf(val))
			}).
			Return(nil).Once()

		require.NoError(t, crCache.GetLatestValue(ctx, testContract, testMethod, testParams{}, &output))
		require.NoError(t, crCache.GetLatestValue(ctx, testContract, testMethod, testParams{}, &output))

		assert.Equal(t, val, output)
	}

	callFunc(t, 8)
	time.Sleep(strategy.Staleness)
	callFunc(t, 9)
}

func TestContractReader_PollAndCache(t *testing.T) {
	t.Parallel()

	mcr := new(mocks.ContractReader)
	crCache := cache.NewChainReader(mcr)
	ctx := tests.Context(t)

	strategy := cache.PollAndCache{
		TriggerAndCache: cache.TriggerAndCache{
			Params: testParams{},
			Value:  reflect.New(reflect.TypeOf(int64(0))).Interface(),
		},
		Interval: 5 * time.Second,
	}

	require.NoError(t, crCache.SetCacheStrategy(testContract, testMethod, strategy))

	t.Cleanup(func() {
		require.NoError(t, crCache.Close())
	})

	// polling will call GetLatestValue so mock the call before the service starts
	mockOnly(t, mcr, 8, strategy.Params)
	mockOnly(t, mcr, 8, strategy.Params)

	require.NoError(t, crCache.Start(ctx))

	// wait for a single polling interval
	time.Sleep(time.Second)
	callAndAssert(ctx, t, crCache, 8, strategy.Params)

	// request with different params falls through
	mockAndCall(ctx, t, crCache, mcr, true, 9, testParams{A: 2})
	time.Sleep(strategy.Interval)

	// after polling interval
	callAndAssert(ctx, t, crCache, 8, strategy.Params)
}

func mockAndCall(
	ctx context.Context,
	t *testing.T,
	crCache *cache.ChainReader,
	mcr *mocks.ContractReader,
	mockCall bool,
	returnVal int64,
	params any,
) {
	t.Helper()

	if mockCall {
		mockOnly(t, mcr, returnVal, params)
	}

	callAndAssert(ctx, t, crCache, returnVal, params)
}

func mockOnly(
	t *testing.T,
	mcr *mocks.ContractReader,
	returnVal int64,
	params any,
) {
	t.Helper()

	mcr.On(
		"GetLatestValue",
		mock.Anything,
		testContract,
		testMethod,
		mock.MatchedBy(func(match testParams) bool {
			return reflect.DeepEqual(match, params)
		}),
		mock.AnythingOfType("*int64")).
		Run(func(args mock.Arguments) {
			reflect.Indirect(reflect.ValueOf(args.Get(4))).Set(reflect.ValueOf(returnVal))
		}).
		Return(nil).Once()
}

func callAndAssert(
	ctx context.Context,
	t *testing.T,
	crCache *cache.ChainReader,
	expected int64,
	params any,
) {
	t.Helper()

	var output int64

	require.NoError(t, crCache.GetLatestValue(ctx, testContract, testMethod, params, &output))
	assert.Equal(t, expected, output)
}
