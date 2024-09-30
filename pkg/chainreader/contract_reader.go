package chainreader

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
)

var (
	ErrNoBindings = errors.New("no bindings found")
)

// WrapContractReaderByIDs returns types.ContractReader behind ContractReaderByIDs interface.
func WrapContractReaderByIDs(contractReader types.ContractReader) ContractReaderByIDs {
	return ContractReaderByIDs{
		cr: contractReader,
	}
}

// ContractReaderByIDs wraps types.ContractReader to allow the caller to set custom contractIDs with Bind.
type ContractReaderByIDs struct {
	bindings sync.Map
	cr       types.ContractReader
}

// Bind accepts a map of bound contracts where map keys are custom contractIDs which will be used by the wrapper to reference contracts.
func (crByIds *ContractReaderByIDs) Bind(ctx context.Context, bindings map[string]types.BoundContract) error {
	var toBind []types.BoundContract
	for customContractID, boundContract := range bindings {
		crByIds.bindings.Store(customContractID, boundContract)
		toBind = append(toBind, boundContract)
	}
	return crByIds.cr.Bind(ctx, toBind)
}

func (crByIds *ContractReaderByIDs) Unbind(ctx context.Context, bindings map[string]types.BoundContract) error {
	var toUnbind []types.BoundContract
	for _, boundContract := range bindings {
		toUnbind = append(toUnbind, boundContract)
	}

	err := crByIds.cr.Unbind(ctx, toUnbind)
	if err == nil {
		for customContractID := range bindings {
			crByIds.bindings.Delete(customContractID)
		}
	}

	return err
}

func (crByIds *ContractReaderByIDs) GetLatestValue(ctx context.Context, contractID, readName string, confidenceLevel primitives.ConfidenceLevel, params, returnVal any) error {
	boundContract, err := crByIds.getBoundContract(contractID)
	if err != nil {
		return err
	}

	return crByIds.cr.GetLatestValue(ctx, boundContract.ReadIdentifier(readName), confidenceLevel, params, returnVal)
}

func (crByIds *ContractReaderByIDs) QueryKey(ctx context.Context, contractID string, filter query.KeyFilter, limitAndSort query.LimitAndSort, sequenceDataType any) ([]types.Sequence, error) {
	boundContract, err := crByIds.getBoundContract(contractID)
	if err != nil {
		return nil, err
	}

	return crByIds.cr.QueryKey(ctx, boundContract, filter, limitAndSort, sequenceDataType)
}

type BatchGetLatestValuesRequestByCustomID map[string]types.ContractBatch
type BatchGetLatestValuesResultByCustomID map[string]types.ContractBatchResults

func (crByIds *ContractReaderByIDs) BatchGetLatestValues(ctx context.Context, request BatchGetLatestValuesRequestByCustomID) (BatchGetLatestValuesResultByCustomID, error) {
	bcToID := make(map[string]string)
	req := make(types.BatchGetLatestValuesRequest)
	for contractID, batch := range request {
		boundContract, err := crByIds.getBoundContract(contractID)
		if err != nil {
			return nil, err
		}

		req[boundContract] = batch
		bcToID[boundContract.String()] = contractID
	}

	res, err := crByIds.cr.BatchGetLatestValues(ctx, req)
	if err != nil {
		return nil, err
	}

	wrappedRes := make(BatchGetLatestValuesResultByCustomID)
	for bc, batchResp := range res {
		wrappedRes[bcToID[bc.String()]] = batchResp
	}

	return wrappedRes, nil
}

func (crByIds *ContractReaderByIDs) getBoundContract(contractID string) (types.BoundContract, error) {
	binding, ok := crByIds.bindings.Load(contractID)
	if !ok {
		return types.BoundContract{}, fmt.Errorf("%w for contractID: %s", ErrNoBindings, contractID)
	}

	boundContract, ok := binding.(types.BoundContract)
	if !ok {
		return types.BoundContract{}, fmt.Errorf("binding found for contractID %s, but is malformed", contractID)
	}
	return boundContract, nil
}
