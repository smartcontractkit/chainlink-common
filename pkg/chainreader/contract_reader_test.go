package chainreader

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/mocks"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

func TestContractReaderByIDsGetLatestValue(t *testing.T) {
	ctx := tests.Context(t)
	mockReader := new(mocks.ContractReader)
	crByIDs := &contractReaderByIDs{
		bindings: sync.Map{},
		cr:       mockReader,
	}

	bc1, bc2 := types.BoundContract{Address: "0x123", Name: "testContract1"}, types.BoundContract{Address: "0x321", Name: "testContract2"}
	bc1CustomID, bcCustomID2 := "customID1", "customID2"
	mockReader.On("Bind", ctx, []types.BoundContract{bc1, bc2}).Return(nil)
	require.NoError(t, crByIDs.Bind(ctx, map[string]types.BoundContract{bc1CustomID: bc1, bcCustomID2: bc2}))

	readName1 := "readName1"
	mockReader.On("GetLatestValue", ctx, bc1.ReadIdentifier(readName1), primitives.ConfidenceLevel(""), nil, nil).Return(nil)
	assert.NoError(t, crByIDs.GetLatestValue(ctx, bc1CustomID, readName1, "", nil, nil))

	readName2 := "readName2"
	mockReader.On("GetLatestValue", ctx, bc2.ReadIdentifier(readName2), primitives.ConfidenceLevel(""), nil, nil).Return(nil)
	assert.NoError(t, crByIDs.GetLatestValue(ctx, bcCustomID2, readName2, "", nil, nil))

	// After unbinding the contract shouldn't be registered and should return error
	mockReader.On("Unbind", ctx, []types.BoundContract{bc1}).Return(nil)
	require.NoError(t, crByIDs.Unbind(ctx, map[string]types.BoundContract{bc1CustomID: bc1}))
	mockReader.On("GetLatestValue", ctx, bc1.ReadIdentifier(readName1), primitives.ConfidenceLevel(""), nil, nil).Return(nil)
	assert.Error(t, crByIDs.GetLatestValue(ctx, bc1CustomID, readName1, "", nil, nil))

	// contract 2 should still be registered
	mockReader.On("GetLatestValue", ctx, bc2.ReadIdentifier(readName2), primitives.ConfidenceLevel(""), nil, nil).Return(nil)
	assert.NoError(t, crByIDs.GetLatestValue(ctx, bcCustomID2, readName2, "", nil, nil))

	// After unbinding the contract2 shouldn't be registered and should return error
	mockReader.On("Unbind", ctx, []types.BoundContract{bc2}).Return(nil)
	require.NoError(t, crByIDs.Unbind(ctx, map[string]types.BoundContract{bcCustomID2: bc2}))
	mockReader.On("GetLatestValue", ctx, bc2.ReadIdentifier(readName2), primitives.ConfidenceLevel(""), nil, nil).Return(nil)
	assert.Error(t, crByIDs.GetLatestValue(ctx, bcCustomID2, readName2, "", nil, nil))

	mockReader.AssertExpectations(t)
}

func TestContractReaderByIDsQueryKey(t *testing.T) {
	ctx := tests.Context(t)
	mockReader := new(mocks.ContractReader)
	crByIDs := &contractReaderByIDs{
		bindings: sync.Map{},
		cr:       mockReader,
	}

	bc1, bc2 := types.BoundContract{Address: "0x123", Name: "testContract1"}, types.BoundContract{Address: "0x321", Name: "testContract2"}
	bc1CustomID, bcCustomID2 := "customID1", "customID2"
	mockReader.On("Bind", ctx, []types.BoundContract{bc1, bc2}).Return(nil)
	require.NoError(t, crByIDs.Bind(ctx, map[string]types.BoundContract{bc1CustomID: bc1, bcCustomID2: bc2}))

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
	sequences2, err := crByIDs.QueryKey(ctx, bcCustomID2, filter, limitAndSort, sequenceDataType)
	assert.NoError(t, err)
	assert.Equal(t, expectedSequences2, sequences2)

	// After unbinding contract1 shouldn't be registered and should return error
	mockReader.On("Unbind", ctx, []types.BoundContract{bc1}).Return(nil)
	require.NoError(t, crByIDs.Unbind(ctx, map[string]types.BoundContract{bc1CustomID: bc1}))
	_, err = crByIDs.QueryKey(ctx, bc1CustomID, filter, limitAndSort, sequenceDataType)
	assert.Error(t, err)

	// contract2 should still be registered
	mockReader.On("QueryKey", ctx, bc2, filter, limitAndSort, sequenceDataType)
	_, err = crByIDs.QueryKey(ctx, bcCustomID2, filter, limitAndSort, sequenceDataType)
	assert.NoError(t, err)

	// After unbinding contract2 shouldn't be registered and should return error
	mockReader.On("Unbind", ctx, []types.BoundContract{bc2}).Return(nil)
	require.NoError(t, crByIDs.Unbind(ctx, map[string]types.BoundContract{bcCustomID2: bc2}))
	_, err = crByIDs.QueryKey(ctx, bcCustomID2, filter, limitAndSort, sequenceDataType)
	assert.Error(t, err)

	mockReader.AssertExpectations(t)
}
