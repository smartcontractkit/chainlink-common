package chainreader

import (
	"fmt"
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/mocks"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

func TestContractReaderByIDsUnbind(t *testing.T) {
	ctx := tests.Context(t)
	mockReader := new(mocks.ContractReader)
	crByIDs := &ContractReaderByIDs{
		bindings: sync.Map{},
		cr:       mockReader,
	}

	bc1, bc2 := types.BoundContract{Address: "0x123", Name: "testContract1"}, types.BoundContract{Address: "0x321", Name: "testContract2"}
	bc1CustomID, bc2CustomID := "customID1", "customID2"
	// order can vary, so we need to match the arguments in a way that ignores order
	bindMatcher := mock.MatchedBy(func(bcsArg []types.BoundContract) bool {
		expectedBcs := []types.BoundContract{bc1, bc2}
		return assert.ElementsMatchf(t, bcsArg, expectedBcs, fmt.Sprintf("expected %v, got %v", expectedBcs, bcsArg))
	})

	mockReader.On("Bind", ctx, bindMatcher).Return(nil)
	require.NoError(t, crByIDs.Bind(ctx, map[string]types.BoundContract{bc1CustomID: bc1, bc2CustomID: bc2}))

	// bindings shouldn't be removed from map if underlying CR impl. returned error
	mockReader.On("Unbind", ctx, []types.BoundContract{bc1}).Return(fmt.Errorf("some error")).Once()
	require.Error(t, crByIDs.Unbind(ctx, map[string]types.BoundContract{bc1CustomID: bc1}))
	val1, ok1 := crByIDs.bindings.Load(bc1CustomID)
	val2, ok2 := crByIDs.bindings.Load(bc2CustomID)
	require.True(t, ok1)
	require.Equal(t, val1, bc1)
	require.True(t, ok2)
	require.Equal(t, val2, bc2)

	// on happy path bindings are properly removed from map
	mockReader.On("Unbind", ctx, []types.BoundContract{bc1}).Return(nil)
	require.NoError(t, crByIDs.Unbind(ctx, map[string]types.BoundContract{bc1CustomID: bc1}))
	_, ok1 = crByIDs.bindings.Load(bc1CustomID)
	val2, ok2 = crByIDs.bindings.Load(bc2CustomID)
	require.False(t, ok1)
	require.True(t, ok2)
	require.Equal(t, val2, bc2)
}

func TestContractReaderByIDsGetLatestValue(t *testing.T) {
	ctx := tests.Context(t)
	mockReader := new(mocks.ContractReader)
	crByIDs := &ContractReaderByIDs{
		bindings: sync.Map{},
		cr:       mockReader,
	}

	bc1, bc2 := types.BoundContract{Address: "0x123", Name: "testContract1"}, types.BoundContract{Address: "0x321", Name: "testContract2"}
	bc1CustomID, bc2CustomID := "customID1", "customID2"
	// order can vary, so we need to match the arguments in a way that ignores order
	bindMatcher := mock.MatchedBy(func(bcsArg []types.BoundContract) bool {
		expectedBcs := []types.BoundContract{bc1, bc2}
		return assert.ElementsMatchf(t, bcsArg, expectedBcs, fmt.Sprintf("expected %v, got %v", expectedBcs, bcsArg))
	})

	mockReader.On("Bind", ctx, bindMatcher).Return(nil)
	require.NoError(t, crByIDs.Bind(ctx, map[string]types.BoundContract{bc1CustomID: bc1, bc2CustomID: bc2}))

	readName1 := "readName1"
	mockReader.On("GetLatestValue", ctx, bc1.ReadIdentifier(readName1), primitives.ConfidenceLevel(""), nil, nil).Return(nil)
	assert.NoError(t, crByIDs.GetLatestValue(ctx, bc1CustomID, readName1, "", nil, nil))

	readName2 := "readName2"
	mockReader.On("GetLatestValue", ctx, bc2.ReadIdentifier(readName2), primitives.ConfidenceLevel(""), nil, nil).Return(nil)
	assert.NoError(t, crByIDs.GetLatestValue(ctx, bc2CustomID, readName2, "", nil, nil))

	// After unbinding the contract shouldn't be registered and should return error
	mockReader.On("Unbind", ctx, []types.BoundContract{bc1}).Return(nil)
	require.NoError(t, crByIDs.Unbind(ctx, map[string]types.BoundContract{bc1CustomID: bc1}))
	mockReader.On("GetLatestValue", ctx, bc1.ReadIdentifier(readName1), primitives.ConfidenceLevel(""), nil, nil).Return(nil)
	assert.Error(t, crByIDs.GetLatestValue(ctx, bc1CustomID, readName1, "", nil, nil))

	// contract 2 should still be registered
	mockReader.On("GetLatestValue", ctx, bc2.ReadIdentifier(readName2), primitives.ConfidenceLevel(""), nil, nil).Return(nil)
	assert.NoError(t, crByIDs.GetLatestValue(ctx, bc2CustomID, readName2, "", nil, nil))

	// After unbinding the contract2 shouldn't be registered and should return error
	mockReader.On("Unbind", ctx, []types.BoundContract{bc2}).Return(nil)
	require.NoError(t, crByIDs.Unbind(ctx, map[string]types.BoundContract{bc2CustomID: bc2}))
	mockReader.On("GetLatestValue", ctx, bc2.ReadIdentifier(readName2), primitives.ConfidenceLevel(""), nil, nil).Return(nil)
	assert.Error(t, crByIDs.GetLatestValue(ctx, bc2CustomID, readName2, "", nil, nil))

	mockReader.AssertExpectations(t)
}

func TestContractReaderByIDsQueryKey(t *testing.T) {
	ctx := tests.Context(t)
	mockReader := new(mocks.ContractReader)
	crByIDs := &ContractReaderByIDs{
		bindings: sync.Map{},
		cr:       mockReader,
	}

	bc1, bc2 := types.BoundContract{Address: "0x123", Name: "testContract1"}, types.BoundContract{Address: "0x321", Name: "testContract2"}
	bc1CustomID, bc2CustomID := "customID1", "customID2"
	// order can vary, so we need to match the arguments in a way that ignores order
	bindMatcher := mock.MatchedBy(func(bcsArg []types.BoundContract) bool {
		expectedBcs := []types.BoundContract{bc1, bc2}
		return assert.ElementsMatchf(t, bcsArg, expectedBcs, fmt.Sprintf("expected %v, got %v", expectedBcs, bcsArg))
	})
	mockReader.On("Bind", ctx, bindMatcher).Return(nil)
	require.NoError(t, crByIDs.Bind(ctx, map[string]types.BoundContract{bc1CustomID: bc1, bc2CustomID: bc2}))

	filter := query.KeyFilter{}
	limitAndSort := query.LimitAndSort{}
	var sequenceDataType any

	// query contract1
	expectedSequences := []types.Sequence{{Data: "sequenceData1"}}
	mockReader.On("QueryKey", ctx, bc1, filter, limitAndSort, sequenceDataType).Return(expectedSequences, nil)
	sequences, err := crByIDs.QueryKey(ctx, bc1CustomID, filter, limitAndSort, sequenceDataType)
	assert.NoError(t, err)
	assert.Equal(t, expectedSequences, sequences)

	// query contract2
	expectedSequences2 := []types.Sequence{{Data: "sequenceData2"}}
	mockReader.On("QueryKey", ctx, bc2, filter, limitAndSort, sequenceDataType).Return(expectedSequences2, nil)
	sequences2, err := crByIDs.QueryKey(ctx, bc2CustomID, filter, limitAndSort, sequenceDataType)
	assert.NoError(t, err)
	assert.Equal(t, expectedSequences2, sequences2)

	// After unbinding contract1 shouldn't be registered and should return error
	mockReader.On("Unbind", ctx, []types.BoundContract{bc1}).Return(nil)
	require.NoError(t, crByIDs.Unbind(ctx, map[string]types.BoundContract{bc1CustomID: bc1}))
	_, err = crByIDs.QueryKey(ctx, bc1CustomID, filter, limitAndSort, sequenceDataType)
	assert.Error(t, err)

	// contract2 should still be registered
	mockReader.On("QueryKey", ctx, bc2, filter, limitAndSort, sequenceDataType)
	_, err = crByIDs.QueryKey(ctx, bc2CustomID, filter, limitAndSort, sequenceDataType)
	assert.NoError(t, err)

	// After unbinding contract2 shouldn't be registered and should return error
	mockReader.On("Unbind", ctx, []types.BoundContract{bc2}).Return(nil)
	require.NoError(t, crByIDs.Unbind(ctx, map[string]types.BoundContract{bc2CustomID: bc2}))
	_, err = crByIDs.QueryKey(ctx, bc2CustomID, filter, limitAndSort, sequenceDataType)
	assert.Error(t, err)

	mockReader.AssertExpectations(t)
}

func TestContractReaderByIDsBatchGetLatestValues(t *testing.T) {
	ctx := tests.Context(t)
	mockReader := new(mocks.ContractReader)
	crByIDs := &ContractReaderByIDs{
		bindings: sync.Map{},
		cr:       mockReader,
	}

	bc1, bc2 := types.BoundContract{Address: "0x123", Name: "testContract1"}, types.BoundContract{Address: "0x321", Name: "testContract2"}
	bc1CustomID, bc2CustomID := "customID1", "customID2"

	// order can vary, so we need to match the arguments in a way that ignores order
	bindMatcher := mock.MatchedBy(func(bcsArg []types.BoundContract) bool {
		expectedBcs := []types.BoundContract{bc1, bc2}
		return assert.ElementsMatchf(t, bcsArg, expectedBcs, fmt.Sprintf("expected %v, got %v", expectedBcs, bcsArg))
	})

	mockReader.On("Bind", ctx, bindMatcher).Return(nil)
	require.NoError(t, crByIDs.Bind(ctx, map[string]types.BoundContract{bc1CustomID: bc1, bc2CustomID: bc2}))

	// Requests
	bc1Batch := []types.BatchRead{
		{ReadName: "read1"},
		{ReadName: "read2"},
	}
	bc2Batch := []types.BatchRead{
		{ReadName: "read3"},
	}

	request := types.BatchGetLatestValuesRequest{
		bc1: bc1Batch,
		bc2: bc2Batch,
	}

	requestByCustomIDs := BatchGetLatestValuesRequestByCustomID{
		bc1CustomID: bc1Batch,
		bc2CustomID: bc2Batch,
	}

	// order can vary, so we need to match the arguments in a way that ignores order
	mapArgMatcher := mock.MatchedBy(func(arg types.BatchGetLatestValuesRequest) bool {
		return reflect.DeepEqual(arg, request)
	})

	// Results
	bc1BatchResult1 := types.BatchReadResult{ReadName: bc1Batch[0].ReadName}
	bc1BatchResult1.SetResult("res-"+bc1Batch[0].ReadName, nil)
	bc1BatchResult2 := types.BatchReadResult{ReadName: bc1Batch[1].ReadName}
	bc1BatchResult2.SetResult(nil, fmt.Errorf("err"))

	bc2BatchResult1 := types.BatchReadResult{ReadName: bc2Batch[0].ReadName}
	bc2BatchResult1.SetResult("res-"+bc1Batch[0].ReadName, nil)

	result := types.BatchGetLatestValuesResult{
		bc1: {bc1BatchResult1, bc1BatchResult2},
		bc2: {bc2BatchResult1}}

	resultByCustomIDs := BatchGetLatestValuesResultByCustomID{
		bc1CustomID: {bc1BatchResult1, bc1BatchResult2},
		bc2CustomID: {bc2BatchResult1}}

	// Batch read both contracts
	mockReader.On("BatchGetLatestValues", ctx, mapArgMatcher).Return(result, nil)
	results, err := crByIDs.BatchGetLatestValues(ctx, requestByCustomIDs)
	assert.NoError(t, err)
	assert.Equal(t, results, resultByCustomIDs)

	// After unbinding bc1, it shouldn't be registered, and an error should occur
	mockReader.On("Unbind", ctx, []types.BoundContract{bc1}).Return(nil)
	require.NoError(t, crByIDs.Unbind(ctx, map[string]types.BoundContract{bc1CustomID: bc1}))
	_, err = crByIDs.BatchGetLatestValues(ctx, requestByCustomIDs)
	assert.Error(t, err)
	_, err = crByIDs.BatchGetLatestValues(ctx, BatchGetLatestValuesRequestByCustomID{bc1CustomID: bc1Batch})
	assert.Error(t, err)

	// Ensure contract 2 is still bound and works
	mockReader.On("BatchGetLatestValues", ctx, types.BatchGetLatestValuesRequest{bc2: bc2Batch}).Return(types.BatchGetLatestValuesResult{bc2: types.ContractBatchResults{bc2BatchResult1}}, nil)
	results, err = crByIDs.BatchGetLatestValues(ctx, BatchGetLatestValuesRequestByCustomID{bc2CustomID: bc2Batch})
	assert.NoError(t, err)
	assert.Equal(t, results, BatchGetLatestValuesResultByCustomID{bc2CustomID: {bc2BatchResult1}})

	// After unbinding contract 2, it should also raise an error
	mockReader.On("Unbind", ctx, []types.BoundContract{bc2}).Return(nil)
	require.NoError(t, crByIDs.Unbind(ctx, map[string]types.BoundContract{bc2CustomID: bc2}))
	_, err = crByIDs.BatchGetLatestValues(ctx, BatchGetLatestValuesRequestByCustomID{bc2CustomID: bc2Batch})
	assert.Error(t, err)

	mockReader.AssertExpectations(t)
}
