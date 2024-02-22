package pluginprovider_test

import (
	"context"
	"errors"
	"fmt"

	libocr "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	"github.com/stretchr/testify/assert"
)

type ConfigProviderTestConfig struct {
	OffchainConfigDigesterTestConfig
	ContractConfigTrackerTestConfig
}

type StaticConfigProvider struct {
	ConfigProviderTestConfig
}

// TODO validate start/Close calls?
func (s StaticConfigProvider) Start(ctx context.Context) error { return nil }

func (s StaticConfigProvider) Close() error { return nil }

func (s StaticConfigProvider) Ready() error { panic("unimplemented") }

func (s StaticConfigProvider) Name() string { panic("unimplemented") }

func (s StaticConfigProvider) HealthReport() map[string]error { panic("unimplemented") }

func (s StaticConfigProvider) OffchainConfigDigester() libocr.OffchainConfigDigester {
	return staticOffchainConfigDigester{s.OffchainConfigDigesterTestConfig}
}

func (s StaticConfigProvider) ContractConfigTracker() libocr.ContractConfigTracker {
	return staticContractConfigTracker{s.ContractConfigTrackerTestConfig}
}

type OffchainConfigDigesterTestConfig struct {
	ContractConfig     libocr.ContractConfig
	ConfigDigest       libocr.ConfigDigest
	ConfigDigestPrefix libocr.ConfigDigestPrefix
}

type staticOffchainConfigDigester struct {
	OffchainConfigDigesterTestConfig
}

func (s staticOffchainConfigDigester) ConfigDigest(config libocr.ContractConfig) (libocr.ConfigDigest, error) {
	if !assert.ObjectsAreEqual(s.ContractConfig, config) {
		return libocr.ConfigDigest{}, fmt.Errorf("expected contract config %v but got %v", s.ConfigDigest, config)
	}
	return s.OffchainConfigDigesterTestConfig.ConfigDigest, nil
}

func (s staticOffchainConfigDigester) ConfigDigestPrefix() (libocr.ConfigDigestPrefix, error) {
	return s.OffchainConfigDigesterTestConfig.ConfigDigestPrefix, nil
}

type ContractConfigTrackerTestConfig struct {
	ContractConfig libocr.ContractConfig
	ConfigDigest   libocr.ConfigDigest
	ChangedInBlock uint64
	BlockHeight    uint64
}

type staticContractConfigTracker struct {
	ContractConfigTrackerTestConfig
}

func (s staticContractConfigTracker) Notify() <-chan struct{} { return nil }

func (s staticContractConfigTracker) LatestConfigDetails(ctx context.Context) (uint64, libocr.ConfigDigest, error) {
	return s.ChangedInBlock, s.ConfigDigest, nil
}

func (s staticContractConfigTracker) LatestConfig(ctx context.Context, cib uint64) (libocr.ContractConfig, error) {
	if s.ChangedInBlock != cib {
		return libocr.ContractConfig{}, fmt.Errorf("expected changed in block %d but got %d", s.ChangedInBlock, cib)
	}
	return s.ContractConfig, nil
}

func (s staticContractConfigTracker) LatestBlockHeight(ctx context.Context) (uint64, error) {
	return s.BlockHeight, nil
}

type staticCodec struct{}

func (c staticCodec) GetMaxEncodingSize(ctx context.Context, n int, itemType string) (int, error) {
	return 0, errors.New("not used for these test")
}

func (c staticCodec) GetMaxDecodingSize(ctx context.Context, n int, itemType string) (int, error) {
	return 0, errors.New("not used for these test")
}

func (c staticCodec) Encode(ctx context.Context, item any, itemType string) ([]byte, error) {
	return nil, errors.New("not used for these test")
}

func (c staticCodec) Decode(ctx context.Context, raw []byte, into any, itemType string) error {
	return errors.New("not used for these test")
}
