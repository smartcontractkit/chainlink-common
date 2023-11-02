package test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	libocr "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	"github.com/stretchr/testify/assert"

	"github.com/smartcontractkit/chainlink-relay/pkg/types"
)

type staticConfigProvider struct{}

// TODO validate start/Close calls?
func (s staticConfigProvider) Start(ctx context.Context) error { return nil }

func (s staticConfigProvider) Close() error { return nil }

func (s staticConfigProvider) Ready() error { panic("unimplemented") }

func (s staticConfigProvider) Name() string { panic("unimplemented") }

func (s staticConfigProvider) HealthReport() map[string]error { panic("unimplemented") }

func (s staticConfigProvider) OffchainConfigDigester() libocr.OffchainConfigDigester {
	return staticOffchainConfigDigester{}
}

func (s staticConfigProvider) ContractConfigTracker() libocr.ContractConfigTracker {
	return staticContractConfigTracker{}
}

type staticOffchainConfigDigester struct{}

func (s staticOffchainConfigDigester) ConfigDigest(config libocr.ContractConfig) (libocr.ConfigDigest, error) {
	if !assert.ObjectsAreEqual(contractConfig, config) {
		return libocr.ConfigDigest{}, fmt.Errorf("expected contract config %v but got %v", configDigest, config)
	}
	return configDigest, nil
}

func (s staticOffchainConfigDigester) ConfigDigestPrefix() (libocr.ConfigDigestPrefix, error) {
	return configDigestPrefix, nil
}

type staticContractConfigTracker struct{}

func (s staticContractConfigTracker) Notify() <-chan struct{} { return nil }

func (s staticContractConfigTracker) LatestConfigDetails(ctx context.Context) (uint64, libocr.ConfigDigest, error) {
	return changedInBlock, configDigest, nil
}

func (s staticContractConfigTracker) LatestConfig(ctx context.Context, cib uint64) (libocr.ContractConfig, error) {
	if changedInBlock != cib {
		return libocr.ContractConfig{}, fmt.Errorf("expected changed in block %d but got %d", changedInBlock, cib)
	}
	return contractConfig, nil
}

func (s staticContractConfigTracker) LatestBlockHeight(ctx context.Context) (uint64, error) {
	return blockHeight, nil
}

type staticContractTransmitter struct{}

func (s staticContractTransmitter) Transmit(ctx context.Context, rc libocr.ReportContext, r libocr.Report, ss []libocr.AttributedOnchainSignature) error {
	if !assert.ObjectsAreEqual(reportContext, rc) {
		return fmt.Errorf("expected report context %v but got %v", reportContext, report)
	}
	if !bytes.Equal(report, r) {
		return fmt.Errorf("expected report %x but got %x", report, r)
	}
	if !assert.ObjectsAreEqual(sigs, ss) {
		return fmt.Errorf("expected signatures %v but got %v", sigs, ss)
	}
	return nil
}

func (s staticContractTransmitter) LatestConfigDigestAndEpoch(ctx context.Context) (libocr.ConfigDigest, uint32, error) {
	return configDigest, epoch, nil
}

func (s staticContractTransmitter) FromAccount() (libocr.Account, error) {
	return account, nil
}

type staticChainReader struct{}

func (c staticChainReader) Encode(ctx context.Context, item any, itemType string) (libocr.Report, error) {
	return nil, errors.New("not used for these test")
}

func (c staticChainReader) Decode(ctx context.Context, raw []byte, into any, itemType string) error {
	return errors.New("not used for these test")
}

func (c staticChainReader) GetLatestValue(ctx context.Context, bc types.BoundContract, method string, params, retVal any) error {
	if !assert.ObjectsAreEqual(bc, boundContract) {
		return fmt.Errorf("expected report context %v but got %v", boundContract, bc)
	}
	if method != medianContractGenericMethod {
		return fmt.Errorf("expected generic contract method %v but got %v", medianContractGenericMethod, method)
	}

	jsonParams, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("Received non json-marshable params in GetLatestValue: %v", params)
	}
	var gotParams GetLatestValueParams
	err = json.Unmarshal(jsonParams, &gotParams)
	if err != nil {
		return fmt.Errorf("Invalid params received in GetLatestValue, must be unmarshable into type %T", reflect.TypeOf(getLatestValueParams))
	}

	if !assert.ObjectsAreEqual(gotParams, getLatestValueParams) {
		return fmt.Errorf("expected params %v but got %v", getLatestValueParams, gotParams)
	}

	ret, ok := retVal.(*map[string]any)
	if !ok {
		return fmt.Errorf("Wrong type passed for retVal param to GetLatestValue impl (expected %T instead of %T", reflect.TypeOf(retVal), reflect.TypeOf(ret))
	}

	if ret == nil {
		return fmt.Errorf("retVal should not be nil")
	}

	(*ret)["ConfigDigest"] = latestTransmissionDetails.ConfigDigest
	(*ret)["Epoch"] = latestTransmissionDetails.Epoch
	(*ret)["Round"] = latestTransmissionDetails.Round
	(*ret)["LatestAnswer"] = latestTransmissionDetails.LatestAnswer
	(*ret)["Timestamp"] = latestTransmissionDetails.Timestamp
	return nil
}
