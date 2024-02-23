package pluginprovider_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
	libocr "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	"github.com/stretchr/testify/assert"
)

type ConfigProviderTestConfig struct {
	OffchainConfigDigesterTestConfig
	ContractConfigTrackerTestConfig
}

type StaticConfigProvider struct {
	ConfigProviderTestConfig
	offchainDigester      StaticOffchainConfigDigester
	contractConfigTracker StaticContractConfigTracker
}

var _ types.ConfigProvider = StaticConfigProvider{}

// TODO validate start/Close calls?
func (s StaticConfigProvider) Start(ctx context.Context) error {
	// TODO lazy intialization
	s.offchainDigester = TestOffchainConfigDigester
	s.contractConfigTracker = TestContractConfigTracker
	return nil
}

func (s StaticConfigProvider) Close() error { return nil }

func (s StaticConfigProvider) Ready() error { panic("unimplemented") }

func (s StaticConfigProvider) Name() string { panic("unimplemented") }

func (s StaticConfigProvider) HealthReport() map[string]error { panic("unimplemented") }

func (s StaticConfigProvider) OffchainConfigDigester() libocr.OffchainConfigDigester {
	return StaticOffchainConfigDigester{s.OffchainConfigDigesterTestConfig}
}

func (s StaticConfigProvider) ContractConfigTracker() libocr.ContractConfigTracker {
	return StaticContractConfigTracker{s.ContractConfigTrackerTestConfig}
}

func (s StaticConfigProvider) AssertEqual(t *testing.T, cp types.ConfigProvider) {
	t.Run("OffchainConfigDigester", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, s.offchainDigester.Equal(cp.OffchainConfigDigester()))
	})
	t.Run("ContractConfigTracker", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, s.contractConfigTracker.Equal(context.Background(), cp.ContractConfigTracker()))
	})
}

/*
func (s StaticConfigProvider) Evaluate(cp types.ConfigProvider) error {
	if err :=  .Equal(cp.OffchainConfigDigester()); err != nil {
		return fmt.Errorf("OffchainConfigDigester: %w", err)
	}
	if err := s.ContractConfigTracker().Equal(context.Background(), cp.ContractConfigTracker()); err != nil {
		return fmt.Errorf("ContractConfigTracker: %w", err)
	}
	return nil

}

*/

type OffchainConfigDigesterTester interface {
	libocr.OffchainConfigDigester
	// Evaluate runs all the methods of the other OffchainConfigDigester and
	// checks for equality to this one
	Evaluate(ctx context.Context, other libocr.OffchainConfigDigester) error
}
type OffchainConfigDigesterTestConfig struct {
	ContractConfig     libocr.ContractConfig
	ConfigDigest       libocr.ConfigDigest
	ConfigDigestPrefix libocr.ConfigDigestPrefix
}

type StaticOffchainConfigDigester struct {
	OffchainConfigDigesterTestConfig
}

var _ libocr.OffchainConfigDigester = StaticOffchainConfigDigester{}

func (s StaticOffchainConfigDigester) ConfigDigest(config libocr.ContractConfig) (libocr.ConfigDigest, error) {
	if !assert.ObjectsAreEqual(s.ContractConfig, config) {
		return libocr.ConfigDigest{}, fmt.Errorf("expected contract config %v but got %v", s.ConfigDigest, config)
	}
	return s.OffchainConfigDigesterTestConfig.ConfigDigest, nil
}

func (s StaticOffchainConfigDigester) ConfigDigestPrefix() (libocr.ConfigDigestPrefix, error) {
	return s.OffchainConfigDigesterTestConfig.ConfigDigestPrefix, nil
}

func (s StaticOffchainConfigDigester) Equal(ocd libocr.OffchainConfigDigester) error {
	gotDigestPrefix, err := ocd.ConfigDigestPrefix()
	if err != nil {
		return fmt.Errorf("failed to get ConfigDigestPrefix: %w", err)
	}
	if gotDigestPrefix != s.OffchainConfigDigesterTestConfig.ConfigDigestPrefix {
		return fmt.Errorf("expected ConfigDigestPrefix %x but got %x", s.OffchainConfigDigesterTestConfig.ConfigDigestPrefix, gotDigestPrefix)
	}
	gotDigest, err := ocd.ConfigDigest(contractConfig)
	if err != nil {
		return fmt.Errorf("failed to get ConfigDigest: %w", err)
	}
	if gotDigest != s.OffchainConfigDigesterTestConfig.ConfigDigest {
		return fmt.Errorf("expected ConfigDigest %x but got %x", s.OffchainConfigDigesterTestConfig.ConfigDigest, gotDigest)
	}
	return nil
}

type ContractConfigTrackerTestConfig struct {
	ContractConfig libocr.ContractConfig
	ConfigDigest   libocr.ConfigDigest
	ChangedInBlock uint64
	BlockHeight    uint64
}

type StaticContractConfigTracker struct {
	ContractConfigTrackerTestConfig
}

func (s StaticContractConfigTracker) Notify() <-chan struct{} { return nil }

func (s StaticContractConfigTracker) LatestConfigDetails(ctx context.Context) (uint64, libocr.ConfigDigest, error) {
	return s.ChangedInBlock, s.ConfigDigest, nil
}

func (s StaticContractConfigTracker) LatestConfig(ctx context.Context, cib uint64) (libocr.ContractConfig, error) {
	if s.ChangedInBlock != cib {
		return libocr.ContractConfig{}, fmt.Errorf("expected changed in block %d but got %d", s.ChangedInBlock, cib)
	}
	return s.ContractConfig, nil
}

func (s StaticContractConfigTracker) LatestBlockHeight(ctx context.Context) (uint64, error) {
	return s.BlockHeight, nil
}

func (s StaticContractConfigTracker) Equal(ctx context.Context, cct libocr.ContractConfigTracker) error {
	gotCIB, gotDigest, err := cct.LatestConfigDetails(ctx)
	if err != nil {
		return fmt.Errorf("failed to get LatestConfigDetails: %w", err)
	}
	if gotCIB != s.ChangedInBlock {
		return fmt.Errorf("expected changed in block %d but got %d", s.ChangedInBlock, gotCIB)
	}
	if gotDigest != s.ConfigDigest {
		return fmt.Errorf("expected config digest %x but got %x", s.ConfigDigest, gotDigest)
	}
	gotBlockHeight, err := cct.LatestBlockHeight(ctx)
	if err != nil {
		return fmt.Errorf("failed to get LatestBlockHeight: %w", err)
	}
	if gotBlockHeight != s.BlockHeight {
		return fmt.Errorf("expected block height %d but got %d", s.BlockHeight, gotBlockHeight)
	}
	gotConfig, err := cct.LatestConfig(ctx, gotCIB)
	if err != nil {
		return fmt.Errorf("failed to get LatestConfig: %w", err)
	}
	if !reflect.DeepEqual(gotConfig, s.ContractConfig) {
		return fmt.Errorf("expected ContractConfig %v but got %v", s.ContractConfig, gotConfig)
	}
	return nil
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
