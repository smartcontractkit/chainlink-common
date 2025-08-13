package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
	"github.com/stretchr/testify/assert"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	ocr3test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/ocr3/test"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test"
	testtypes "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test/types"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/ccipocr3"
)

type CCIPProviderEvaluator interface {
	types.CCIPProvider
	testtypes.Evaluator[types.CCIPProvider]
}

type CCIPProviderTester interface {
	types.CCIPProvider
	testtypes.Evaluator[types.CCIPProvider]
	testtypes.AssertEqualer[types.CCIPProvider]
}

// CCIPProvider is a static implementation of the CCIPProviderTester interface.
// It is to be used in tests to verify grpc implementations of the CCIPProvider interface.
func CCIPProvider(lggr logger.Logger) staticCCIPProvider {
	return newStaticCCIPProvider(lggr, staticCCIPProviderConfig{
		chainAccessor:       ChainAccessor(lggr),
		contractTransmitter: ocr3test.ContractTransmitter,
		codec:               Codec(lggr),
		codecEvaluator:      CodecEvaluator(lggr),
	})
}

var _ CCIPProviderTester = staticCCIPProvider{}

type staticCCIPProviderConfig struct {
	chainAccessor       ChainAccessorTester
	contractTransmitter testtypes.OCR3ContractTransmitterEvaluator
	codec               ccipocr3.Codec
	codecEvaluator      codecEvaluator
}

type staticCCIPProvider struct {
	services.Service
	staticCCIPProviderConfig
}

func newStaticCCIPProvider(lggr logger.Logger, cfg staticCCIPProviderConfig) staticCCIPProvider {
	lggr = logger.Named(lggr, "staticCCIPProvider")
	return staticCCIPProvider{
		Service:                  test.NewStaticService(lggr),
		staticCCIPProviderConfig: cfg,
	}
}

// ChainAccessor implements CCIPProviderEvaluator.
func (s staticCCIPProvider) ChainAccessor() ccipocr3.ChainAccessor {
	return s.chainAccessor
}

// ContractTransmitter implements CCIPProviderEvaluator.
func (s staticCCIPProvider) ContractTransmitter() ocr3types.ContractTransmitter[[]byte] {
	return s.contractTransmitter
}

// Codec implements CCIPProviderEvaluator.
func (s staticCCIPProvider) Codec() ccipocr3.Codec {
	return s.codec
}

// Evaluate implements CCIPProviderEvaluator.
func (s staticCCIPProvider) Evaluate(ctx context.Context, other types.CCIPProvider) error {
	// ChainAccessor test case
	err := s.chainAccessor.Evaluate(ctx, other.ChainAccessor())
	if err != nil {
		return evaluationError{err: err, component: "ChainAccessor"}
	}

	// ContractTransmitter test case
	err = s.contractTransmitter.Evaluate(ctx, other.ContractTransmitter())
	if err != nil {
		return evaluationError{err: err, component: "ContractTransmitter"}
	}

	// Codec test case
	err = s.codecEvaluator.Evaluate(ctx, other.Codec())
	if err != nil {
		return evaluationError{err: err, component: "Codec"}
	}

	return nil
}

// AssertEqual implements CCIPProviderTester.
func (s staticCCIPProvider) AssertEqual(ctx context.Context, t *testing.T, other types.CCIPProvider) {
	t.Run("StaticCCIPProvider", func(t *testing.T) {
		// ChainAccessor test case
		t.Run("ChainAccessor", func(t *testing.T) {
			s.chainAccessor.AssertEqual(ctx, t, other.ChainAccessor())
		})

		// ContractTransmitter test case
		t.Run("ContractTransmitter", func(t *testing.T) {
			assert.NoError(t, s.contractTransmitter.Evaluate(ctx, other.ContractTransmitter()))
		})

		// Codec test case
		t.Run("Codec", func(t *testing.T) {
			s.codecEvaluator.AssertEqual(ctx, t, other.Codec())
		})
	})
}

type evaluationError struct {
	err       error
	component string
}

func (e evaluationError) Error() string {
	return fmt.Sprintf("%s evaluation failed: %s", e.component, e.err.Error())
}
