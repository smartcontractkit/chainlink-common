package test

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/types/ccipocr3"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
)

func TestChainAccessor(t *testing.T) {
	ctx := context.Background()
	lggr := logger.Test(t)

	chainAccessor := ChainAccessor(lggr)

	t.Run("GetContractAddress", func(t *testing.T) {
		addr, err := chainAccessor.GetContractAddress("test-contract")
		assert.NoError(t, err)
		assert.Equal(t, []byte("test-contract-address"), addr)
	})

	t.Run("GetAllConfigsLegacy", func(t *testing.T) {
		snapshot, configs, err := chainAccessor.GetAllConfigsLegacy(ctx, 1, []ccipocr3.ChainSelector{2, 3})
		assert.NoError(t, err)
		assert.NotNil(t, snapshot)
		assert.NotNil(t, configs)
	})

	t.Run("GetChainFeeComponents", func(t *testing.T) {
		feeComponents, err := chainAccessor.GetChainFeeComponents(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, feeComponents)
	})

	t.Run("Sync", func(t *testing.T) {
		err := chainAccessor.Sync(ctx, "test-contract", ccipocr3.UnknownAddress("test-address"))
		assert.NoError(t, err)
	})

	t.Run("CommitReportsGTETimestamp", func(t *testing.T) {
		reports, err := chainAccessor.CommitReportsGTETimestamp(ctx, time.Now(), primitives.Unconfirmed, 10)
		assert.NoError(t, err)
		assert.NotNil(t, reports)
	})

	t.Run("ExecutedMessages", func(t *testing.T) {
		ranges := map[ccipocr3.ChainSelector][]ccipocr3.SeqNumRange{
			ccipocr3.ChainSelector(1): {{ccipocr3.SeqNum(1), ccipocr3.SeqNum(10)}},
		}
		executed, err := chainAccessor.ExecutedMessages(ctx, ranges, primitives.Unconfirmed)
		assert.NoError(t, err)
		assert.NotNil(t, executed)
	})

	t.Run("NextSeqNum", func(t *testing.T) {
		seqNums, err := chainAccessor.NextSeqNum(ctx, []ccipocr3.ChainSelector{1, 2})
		assert.NoError(t, err)
		assert.NotNil(t, seqNums)
	})

	t.Run("Nonces", func(t *testing.T) {
		addresses := map[ccipocr3.ChainSelector][]ccipocr3.UnknownEncodedAddress{
			ccipocr3.ChainSelector(1): {ccipocr3.UnknownEncodedAddress("addr1")},
		}
		nonces, err := chainAccessor.Nonces(ctx, addresses)
		assert.NoError(t, err)
		assert.NotNil(t, nonces)
	})

	t.Run("GetChainFeePriceUpdate", func(t *testing.T) {
		updates := chainAccessor.GetChainFeePriceUpdate(ctx, []ccipocr3.ChainSelector{1, 2})
		assert.NotNil(t, updates)
	})

	t.Run("GetLatestPriceSeqNr", func(t *testing.T) {
		seqNr, err := chainAccessor.GetLatestPriceSeqNr(ctx)
		assert.NoError(t, err)
		assert.Equal(t, uint64(42), seqNr)
	})

	t.Run("MsgsBetweenSeqNums", func(t *testing.T) {
		msgs, err := chainAccessor.MsgsBetweenSeqNums(ctx, 1, ccipocr3.SeqNumRange{ccipocr3.SeqNum(1), ccipocr3.SeqNum(10)})
		assert.NoError(t, err)
		assert.NotNil(t, msgs)
	})

	t.Run("LatestMessageTo", func(t *testing.T) {
		seqNr, err := chainAccessor.LatestMessageTo(ctx, 1)
		assert.NoError(t, err)
		assert.Equal(t, ccipocr3.SeqNum(43), seqNr)
	})

	t.Run("GetExpectedNextSequenceNumber", func(t *testing.T) {
		seqNr, err := chainAccessor.GetExpectedNextSequenceNumber(ctx, 1)
		assert.NoError(t, err)
		assert.Equal(t, ccipocr3.SeqNum(44), seqNr)
	})

	t.Run("GetTokenPriceUSD", func(t *testing.T) {
		price, err := chainAccessor.GetTokenPriceUSD(ctx, ccipocr3.UnknownAddress("token1"))
		assert.NoError(t, err)
		assert.NotNil(t, price)
	})

	t.Run("GetFeeQuoterDestChainConfig", func(t *testing.T) {
		config, err := chainAccessor.GetFeeQuoterDestChainConfig(ctx, 1)
		assert.NoError(t, err)
		assert.NotNil(t, config)
	})

	// USDCMessageReader tests
	t.Run("MessagesByTokenID", func(t *testing.T) {
		tokens := map[ccipocr3.MessageTokenID]ccipocr3.RampTokenAmount{
			ccipocr3.NewMessageTokenID(1, 0): {
				SourcePoolAddress: ccipocr3.UnknownAddress("test-source-pool"),
				DestTokenAddress:  ccipocr3.UnknownAddress("test-dest-token"),
				ExtraData:         ccipocr3.Bytes("test-extra-data"),
				Amount:            ccipocr3.NewBigInt(big.NewInt(12345)),
			},
		}
		messages, err := chainAccessor.MessagesByTokenID(ctx, ccipocr3.ChainSelector(1), ccipocr3.ChainSelector(2), tokens)
		assert.NoError(t, err)
		assert.NotNil(t, messages)
		assert.Len(t, messages, 1)
	})

	// PriceReader tests
	t.Run("GetFeedPricesUSD", func(t *testing.T) {
		tokens := []ccipocr3.UnknownEncodedAddress{"token1", "token2", "token3"}
		prices, err := chainAccessor.GetFeedPricesUSD(ctx, tokens)
		assert.NoError(t, err)
		assert.NotNil(t, prices)
		assert.Len(t, prices, 3)
	})

	t.Run("GetFeeQuoterTokenUpdates", func(t *testing.T) {
		tokens := []ccipocr3.UnknownEncodedAddress{"token1", "token2"}
		updates, err := chainAccessor.GetFeeQuoterTokenUpdates(ctx, tokens, ccipocr3.ChainSelector(1))
		assert.NoError(t, err)
		assert.NotNil(t, updates)
		assert.Len(t, updates, 2)
	})
}

func TestChainAccessorEvaluate(t *testing.T) {
	ctx := context.Background()
	lggr := logger.Test(t)

	chainAccessor1 := ChainAccessor(lggr)
	chainAccessor2 := ChainAccessor(lggr)

	err := chainAccessor1.Evaluate(ctx, chainAccessor2)
	assert.NoError(t, err)
}

func TestChainAccessorAssertEqual(t *testing.T) {
	ctx := context.Background()
	lggr := logger.Test(t)

	chainAccessor1 := ChainAccessor(lggr)
	chainAccessor2 := ChainAccessor(lggr)

	chainAccessor1.AssertEqual(ctx, t, chainAccessor2)
}
