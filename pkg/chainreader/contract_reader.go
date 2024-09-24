package chainreader

import (
	"context"
	"fmt"
	"sync"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
)

type ContractReaderByIDs interface {
	Bind(ctx context.Context, bindings map[string]types.BoundContract) error
	Unbind(ctx context.Context, bindings map[string]types.BoundContract) error
	GetLatestValue(ctx context.Context, contractID, readName string, confidenceLevel primitives.ConfidenceLevel, params, returnVal any) error
	BatchGetLatestValues(ctx context.Context, request BatchGetLatestValuesRequestByCustomID) (BatchGetLatestValuesResultByCustomID, error)
	QueryKey(ctx context.Context, contractID string, filter query.KeyFilter, limitAndSort query.LimitAndSort, sequenceDataType any) ([]types.Sequence, error)
}

func WrapContractReaderByIDs(contractReader types.ContractReader) ContractReaderByIDs {
	return &contractReaderByIDs{
		ContractReader: contractReader,
	}
}

type contractReaderByIDs struct {
	bindings sync.Map
	types.ContractReader
}

var _ types.ContractReader = (*contractReaderByIDs)(nil)

type BatchGetLatestValuesRequestByCustomID map[string]types.ContractBatch
type BatchGetLatestValuesResultByCustomID map[string]types.ContractBatchResults

func (crByIds *contractReaderByIDs) Bind(ctx context.Context, bindings map[string]types.BoundContract) error {
	var toBind []types.BoundContract
	for customContractID, boundContract := range bindings {
		crByIds.bindings.Store(customContractID, boundContract)
		toBind = append(toBind, boundContract)
	}
	return crByIds.ContractReader.Bind(ctx, toBind)
}

func (crByIds *contractReaderByIDs) Unbind(ctx context.Context, bindings map[string]types.BoundContract) error {
	var toUnbind []types.BoundContract
	for customContractID, boundContract := range bindings {
		crByIds.bindings.Delete(customContractID)
		toUnbind = append(toUnbind, boundContract)
	}
	return crByIds.ContractReader.Unbind(ctx, toUnbind)
}

func (crByIds *contractReaderByIDs) GetLatestValue(ctx context.Context, contractID, readName string, confidenceLevel primitives.ConfidenceLevel, params, returnVal any) error {
	boundContract, err := crByIds.getBoundContract(contractID)
	if err != nil {
		return err
	}

	return crByIds.ContractReader.GetLatestValue(ctx, boundContract.ReadIdentifier(readName), confidenceLevel, params, returnVal)
}

func (crByIds *contractReaderByIDs) QueryKey(ctx context.Context, contractID string, filter query.KeyFilter, limitAndSort query.LimitAndSort, sequenceDataType any) ([]types.Sequence, error) {
	boundContract, err := crByIds.getBoundContract(contractID)
	if err != nil {
		return nil, err
	}

	return crByIds.ContractReader.QueryKey(ctx, boundContract, filter, limitAndSort, sequenceDataType)
}

func (crByIds *contractReaderByIDs) BatchGetLatestValues(ctx context.Context, request BatchGetLatestValuesRequestByCustomID) (BatchGetLatestValuesResultByCustomID, error) {
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

	res, err := crByIds.ContractReader.BatchGetLatestValues(ctx, req)
	if err != nil {
		return nil, err
	}

	wrappedRes := make(BatchGetLatestValuesResultByCustomID)
	for bc, batchResp := range res {
		wrappedRes[bcToID[bc.String()]] = batchResp
	}

	return wrappedRes, nil
}

func (crByIds *contractReaderByIDs) getBoundContract(contractID string) (types.BoundContract, error) {
	binding, ok := crByIds.bindings.Load(contractID)
	if !ok {
		return types.BoundContract{}, fmt.Errorf("binding not found for contractID %s", contractID)
	}

	boundContract, ok := binding.(types.BoundContract)
	if !ok {
		return types.BoundContract{}, fmt.Errorf("binding found for contractID %s, but is malformed", contractID)
	}
	return boundContract, nil
}
