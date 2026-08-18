package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	testtypes "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/ccipocr3"
)

// CodecEvaluatorInterface defines what a codec evaluator should implement
type CodecEvaluatorInterface interface {
	testtypes.Evaluator[ccipocr3.Codec]
	testtypes.AssertEqualer[ccipocr3.Codec]
}

// Codec is a static implementation of the CodecTester interface.
// It is to be used in tests to verify grpc implementations of the Codec interface.
func Codec(lggr logger.Logger) ccipocr3.Codec {
	return ccipocr3.Codec{
		ChainSpecificAddressCodec: ChainSpecificAddressCodec(lggr),
		CommitPluginCodec:         CommitPluginCodec(lggr),
		ExecutePluginCodec:        ExecutePluginCodec(lggr),
		TokenDataEncoder:          TokenDataEncoder(lggr),
		SourceChainExtraDataCodec: SourceChainExtraDataCodec(lggr),
		MessageHasher:             MessageHasher(lggr),
	}
}

// CodecEvaluator is a helper to evaluate Codec instances
func CodecEvaluator(lggr logger.Logger) codecEvaluator {
	return codecEvaluator{
		chainSpecificAddressCodec: ChainSpecificAddressCodec(lggr),
		commitPluginCodec:         CommitPluginCodec(lggr),
		executePluginCodec:        ExecutePluginCodec(lggr),
		tokenDataEncoder:          TokenDataEncoder(lggr),
		sourceChainExtraDataCodec: SourceChainExtraDataCodec(lggr),
		messageHasher:             MessageHasher(lggr),
	}
}

type codecEvaluator struct {
	chainSpecificAddressCodec ChainSpecificAddressCodecTester
	commitPluginCodec         CommitPluginCodecTester
	executePluginCodec        ExecutePluginCodecTester
	tokenDataEncoder          TokenDataEncoderTester
	sourceChainExtraDataCodec SourceChainExtraDataCodecTester
	messageHasher             MessageHasherTester
}

// Evaluate implements CodecEvaluator.
func (s codecEvaluator) Evaluate(ctx context.Context, other ccipocr3.Codec) error {
	// Test ChainSpecificAddressCodec
	err := s.chainSpecificAddressCodec.Evaluate(ctx, other.ChainSpecificAddressCodec)
	if err != nil {
		return fmt.Errorf("ChainSpecificAddressCodec evaluation failed: %w", err)
	}

	// Test CommitPluginCodec
	err = s.commitPluginCodec.Evaluate(ctx, other.CommitPluginCodec)
	if err != nil {
		return fmt.Errorf("CommitPluginCodec evaluation failed: %w", err)
	}

	// Test ExecutePluginCodec
	err = s.executePluginCodec.Evaluate(ctx, other.ExecutePluginCodec)
	if err != nil {
		return fmt.Errorf("ExecutePluginCodec evaluation failed: %w", err)
	}

	// Test TokenDataEncoder
	err = s.tokenDataEncoder.Evaluate(ctx, other.TokenDataEncoder)
	if err != nil {
		return fmt.Errorf("TokenDataEncoder evaluation failed: %w", err)
	}

	// Test SourceChainExtraDataCodec
	err = s.sourceChainExtraDataCodec.Evaluate(ctx, other.SourceChainExtraDataCodec)
	if err != nil {
		return fmt.Errorf("SourceChainExtraDataCodec evaluation failed: %w", err)
	}

	// Test MessageHasher
	err = s.messageHasher.Evaluate(ctx, other.MessageHasher)
	if err != nil {
		return fmt.Errorf("MessageHasher evaluation failed: %w", err)
	}

	return nil
}

// AssertEqual implements CodecTester.
func (s codecEvaluator) AssertEqual(ctx context.Context, t *testing.T, other ccipocr3.Codec) {
	t.Run("Codec", func(t *testing.T) {
		assert.NoError(t, s.Evaluate(ctx, other))
	})
}
