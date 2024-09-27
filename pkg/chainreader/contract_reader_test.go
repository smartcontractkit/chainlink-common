package chainreader

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

func TestContractReaderByIDsUnbind(t *testing.T) {
	ctx := tests.Context(t)
	mockReader := &ContractReaderMock{}
	crByIDs := &ContractReaderByIDs{
		bindings: sync.Map{},
		cr:       mockReader,
	}

	bc1 := types.BoundContract{Address: "0x123", Name: "testContract1"}
	bc2 := types.BoundContract{Address: "0x321", Name: "testContract2"}
	bc1CustomID, bc2CustomID := "customID1", "customID2"

	// Mock Bind function
	mockReader.bindFunc = func(ctx context.Context, contracts []types.BoundContract) error {
		expectedContracts := []types.BoundContract{bc1, bc2}
		assert.ElementsMatch(t, expectedContracts, contracts)
		return nil
	}
	require.NoError(t, crByIDs.Bind(ctx, map[string]types.BoundContract{bc1CustomID: bc1, bc2CustomID: bc2}))

	// Mock Unbind function for error case
	mockReader.unbindFunc = func(ctx context.Context, contracts []types.BoundContract) error {
		return fmt.Errorf("some error")
	}

	// Test Unbind with error shouldn't remove bindings
	require.Error(t, crByIDs.Unbind(ctx, map[string]types.BoundContract{bc1CustomID: bc1}))
	val1, ok1 := crByIDs.bindings.Load(bc1CustomID)
	val2, ok2 := crByIDs.bindings.Load(bc2CustomID)
	require.True(t, ok1)
	require.Equal(t, val1, bc1)
	require.True(t, ok2)
	require.Equal(t, val2, bc2)

	// Mock Unbind function for success case
	mockReader.unbindFunc = func(ctx context.Context, contracts []types.BoundContract) error {
		assert.Equal(t, []types.BoundContract{bc1}, contracts)
		return nil
	}

	// Test successful Unbind
	require.NoError(t, crByIDs.Unbind(ctx, map[string]types.BoundContract{bc1CustomID: bc1}))
	_, ok1 = crByIDs.bindings.Load(bc1CustomID)
	val2, ok2 = crByIDs.bindings.Load(bc2CustomID)
	require.False(t, ok1)
	require.True(t, ok2)
	require.Equal(t, val2, bc2)
}

func TestContractReaderByIDsGetLatestValue(t *testing.T) {
	ctx := tests.Context(t)
	mockReader := &ContractReaderMock{}
	crByIDs := &ContractReaderByIDs{
		bindings: sync.Map{},
		cr:       mockReader,
	}

	bc1 := types.BoundContract{Address: "0x123", Name: "testContract1"}
	bc2 := types.BoundContract{Address: "0x321", Name: "testContract2"}
	bc1CustomID, bc2CustomID := "customID1", "customID2"

	// Mock Bind function
	mockReader.bindFunc = func(ctx context.Context, contracts []types.BoundContract) error {
		expectedContracts := []types.BoundContract{bc1, bc2}
		assert.ElementsMatch(t, expectedContracts, contracts)
		return nil
	}
	require.NoError(t, crByIDs.Bind(ctx, map[string]types.BoundContract{bc1CustomID: bc1, bc2CustomID: bc2}))

	readName1 := "readName1"
	readName2 := "readName2"

	// Mock GetLatestValue function
	mockReader.getLatestValueFunc = func(ctx context.Context, identifier string, confidence primitives.ConfidenceLevel, params, returnVal any) error {
		if identifier == bc1.ReadIdentifier(readName1) {
			return nil
		}
		if identifier == bc2.ReadIdentifier(readName2) {
			return nil
		}
		return fmt.Errorf("not found")
	}

	// Test GetLatestValue for bc1
	assert.NoError(t, crByIDs.GetLatestValue(ctx, bc1CustomID, readName1, "", nil, nil))

	// Test GetLatestValue for bc2
	assert.NoError(t, crByIDs.GetLatestValue(ctx, bc2CustomID, readName2, "", nil, nil))

	// Unbind bc1 and test again
	mockReader.unbindFunc = func(ctx context.Context, contracts []types.BoundContract) error {
		return nil
	}
	require.NoError(t, crByIDs.Unbind(ctx, map[string]types.BoundContract{bc1CustomID: bc1}))

	assert.Error(t, crByIDs.GetLatestValue(ctx, bc1CustomID, readName1, "", nil, nil))

	// Test that bc2 still works
	assert.NoError(t, crByIDs.GetLatestValue(ctx, bc2CustomID, readName2, "", nil, nil))

	// Unbind bc2 and test again
	require.NoError(t, crByIDs.Unbind(ctx, map[string]types.BoundContract{bc2CustomID: bc2}))
	assert.Error(t, crByIDs.GetLatestValue(ctx, bc2CustomID, readName2, "", nil, nil))
}

func TestContractReaderByIDsQueryKey(t *testing.T) {
	ctx := tests.Context(t)
	mockReader := &ContractReaderMock{}
	crByIDs := &ContractReaderByIDs{
		bindings: sync.Map{},
		cr:       mockReader,
	}

	bc1 := types.BoundContract{Address: "0x123", Name: "testContract1"}
	bc2 := types.BoundContract{Address: "0x321", Name: "testContract2"}
	bc1CustomID, bc2CustomID := "customID1", "customID2"

	// Mock Bind function
	mockReader.bindFunc = func(ctx context.Context, contracts []types.BoundContract) error {
		expectedContracts := []types.BoundContract{bc1, bc2}
		assert.ElementsMatch(t, expectedContracts, contracts)
		return nil
	}
	require.NoError(t, crByIDs.Bind(ctx, map[string]types.BoundContract{bc1CustomID: bc1, bc2CustomID: bc2}))

	// Mock QueryKey function
	mockReader.queryKeyFunc = func(ctx context.Context, contract types.BoundContract, filter query.KeyFilter, limitAndSort query.LimitAndSort, sequenceDataType any) ([]types.Sequence, error) {
		if contract == bc1 {
			return []types.Sequence{{Data: "sequenceData1"}}, nil
		} else if contract == bc2 {
			return []types.Sequence{{Data: "sequenceData2"}}, nil
		}
		return nil, fmt.Errorf("not found")
	}

	// Test QueryKey for bc1
	filter := query.KeyFilter{}
	limitAndSort := query.LimitAndSort{}
	var sequenceDataType any
	sequences, err := crByIDs.QueryKey(ctx, bc1CustomID, filter, limitAndSort, sequenceDataType)
	assert.NoError(t, err)
	assert.Equal(t, []types.Sequence{{Data: "sequenceData1"}}, sequences)

	// Test QueryKey for bc2
	sequences2, err := crByIDs.QueryKey(ctx, bc2CustomID, filter, limitAndSort, sequenceDataType)
	assert.NoError(t, err)
	assert.Equal(t, []types.Sequence{{Data: "sequenceData2"}}, sequences2)

	// Unbind bc1 and test error case
	mockReader.unbindFunc = func(ctx context.Context, contracts []types.BoundContract) error {
		return nil
	}
	require.NoError(t, crByIDs.Unbind(ctx, map[string]types.BoundContract{bc1CustomID: bc1}))
	_, err = crByIDs.QueryKey(ctx, bc1CustomID, filter, limitAndSort, sequenceDataType)
	assert.Error(t, err)

	// Test that bc2 still works
	sequences2, err = crByIDs.QueryKey(ctx, bc2CustomID, filter, limitAndSort, sequenceDataType)
	assert.NoError(t, err)

	// Unbind bc2 and test error case
	require.NoError(t, crByIDs.Unbind(ctx, map[string]types.BoundContract{bc2CustomID: bc2}))
	_, err = crByIDs.QueryKey(ctx, bc2CustomID, filter, limitAndSort, sequenceDataType)
	assert.Error(t, err)
}

func TestContractReaderByIDsBatchGetLatestValues(t *testing.T) {
	ctx := tests.Context(t)
	mockReader := &ContractReaderMock{}
	crByIDs := &ContractReaderByIDs{
		bindings: sync.Map{},
		cr:       mockReader,
	}

	bc1 := types.BoundContract{Address: "0x123", Name: "testContract1"}
	bc2 := types.BoundContract{Address: "0x321", Name: "testContract2"}
	bc1CustomID, bc2CustomID := "customID1", "customID2"

	// Mock Bind function
	mockReader.bindFunc = func(ctx context.Context, contracts []types.BoundContract) error {
		expectedContracts := []types.BoundContract{bc1, bc2}
		assert.ElementsMatch(t, expectedContracts, contracts)
		return nil
	}
	require.NoError(t, crByIDs.Bind(ctx, map[string]types.BoundContract{bc1CustomID: bc1, bc2CustomID: bc2}))

	// Define request for BatchGetLatestValues
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

	// Define mock result
	bc1BatchResult1 := types.BatchReadResult{ReadName: bc1Batch[0].ReadName}
	bc1BatchResult1.SetResult("res-"+bc1Batch[0].ReadName, nil)
	bc1BatchResult2 := types.BatchReadResult{ReadName: bc1Batch[1].ReadName}
	bc1BatchResult2.SetResult(nil, fmt.Errorf("err"))

	bc2BatchResult1 := types.BatchReadResult{ReadName: bc2Batch[0].ReadName}
	bc2BatchResult1.SetResult("res-"+bc2Batch[0].ReadName, nil)

	result := types.BatchGetLatestValuesResult{
		bc1: {bc1BatchResult1, bc1BatchResult2},
		bc2: {bc2BatchResult1},
	}

	resultByCustomIDs := BatchGetLatestValuesResultByCustomID{
		bc1CustomID: {bc1BatchResult1, bc1BatchResult2},
		bc2CustomID: {bc2BatchResult1},
	}

	// Mock BatchGetLatestValues function
	mockReader.batchGetFunc = func(ctx context.Context, req types.BatchGetLatestValuesRequest) (types.BatchGetLatestValuesResult, error) {
		assert.Equal(t, request, req)
		return result, nil
	}

	// Call BatchGetLatestValues
	results, err := crByIDs.BatchGetLatestValues(ctx, requestByCustomIDs)
	assert.NoError(t, err)
	assert.Equal(t, results, resultByCustomIDs)

	// Test after unbinding bc1
	mockReader.unbindFunc = func(ctx context.Context, contracts []types.BoundContract) error {
		assert.Equal(t, []types.BoundContract{bc1}, contracts)
		return nil
	}
	require.NoError(t, crByIDs.Unbind(ctx, map[string]types.BoundContract{bc1CustomID: bc1}))

	_, err = crByIDs.BatchGetLatestValues(ctx, requestByCustomIDs)
	assert.Error(t, err)
}

// ContractReaderMock instead of using mockery because types.UnimplementedContractReader breaks mockery
type ContractReaderMock struct {
	bindFunc           func(context.Context, []types.BoundContract) error
	unbindFunc         func(context.Context, []types.BoundContract) error
	getLatestValueFunc func(context.Context, string, primitives.ConfidenceLevel, any, any) error
	queryKeyFunc       func(context.Context, types.BoundContract, query.KeyFilter, query.LimitAndSort, any) ([]types.Sequence, error)
	batchGetFunc       func(context.Context, types.BatchGetLatestValuesRequest) (types.BatchGetLatestValuesResult, error)
	types.UnimplementedContractReader
}

func (m *ContractReaderMock) Bind(ctx context.Context, contracts []types.BoundContract) error {
	if m.bindFunc != nil {
		return m.bindFunc(ctx, contracts)
	}
	return nil
}

func (m *ContractReaderMock) Unbind(ctx context.Context, contracts []types.BoundContract) error {
	if m.unbindFunc != nil {
		return m.unbindFunc(ctx, contracts)
	}
	return nil
}

func (m *ContractReaderMock) GetLatestValue(ctx context.Context, identifier string, confidence primitives.ConfidenceLevel, params, returnVal any) error {
	if m.getLatestValueFunc != nil {
		return m.getLatestValueFunc(ctx, identifier, confidence, params, returnVal)
	}
	return nil
}

func (m *ContractReaderMock) QueryKey(ctx context.Context, contract types.BoundContract, filter query.KeyFilter, limitAndSort query.LimitAndSort, sequenceDataType any) ([]types.Sequence, error) {
	if m.queryKeyFunc != nil {
		return m.queryKeyFunc(ctx, contract, filter, limitAndSort, sequenceDataType)
	}
	return nil, nil
}

func (m *ContractReaderMock) BatchGetLatestValues(ctx context.Context, request types.BatchGetLatestValuesRequest) (types.BatchGetLatestValuesResult, error) {
	if m.batchGetFunc != nil {
		return m.batchGetFunc(ctx, request)
	}
	return nil, nil
}
