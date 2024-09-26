package chainreader

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/mocks"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

func TestGetLatestValue(t *testing.T) {
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
