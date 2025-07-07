package evm_test

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/chain-capabilities/evm"
	evmtypes "github.com/smartcontractkit/chainlink-common/pkg/types/chains/evm"
	valuespb "github.com/smartcontractkit/chainlink-common/pkg/values/pb"
	"github.com/stretchr/testify/assert"
)

func TestConvertFilterFromProto(t *testing.T) {
	validBlockHash := make([]byte, 32)
	for i := 0; i < 32; i++ {
		validBlockHash[i] = byte(i)
	}

	validAddress := make([]byte, 20)
	for i := 0; i < 20; i++ {
		validAddress[i] = byte(i + 10)
	}

	validTopic := make([]byte, 32)
	for i := 0; i < 32; i++ {
		validTopic[i] = byte(i + 20)
	}

	t.Run("nil protoFilter returns error", func(t *testing.T) {
		_, err := evm.ConvertFilterFromProto(nil)
		assert.ErrorContains(t, err, "filter can't be nil")
	})

	t.Run("successful conversion", func(t *testing.T) {
		fromBlock := &valuespb.BigInt{AbsVal: []byte{1, 2, 3}, Sign: 0}
		toBlock := &valuespb.BigInt{AbsVal: []byte{1, 2, 4}, Sign: 0}
		validTopics := []*evm.Topics{{Topic: [][]byte{validTopic}}}
		input := &evm.FilterQuery{
			BlockHash: validBlockHash,
			FromBlock: fromBlock,
			ToBlock:   toBlock,
			Addresses: [][]byte{validAddress},
			Topics:    validTopics,
		}

		// expected outputs from conversions
		expectedHash := common.Hash(validBlockHash)
		expectedAddr := common.Address(validAddress)
		expectedTopics := evm.ConvertTopicsFromProto(validTopics)

		result, err := evm.ConvertFilterFromProto(input)
		assert.NoError(t, err)

		assert.ElementsMatch(t, expectedHash, result.BlockHash)
		assert.Equal(t, valuespb.NewIntFromBigInt(fromBlock), result.FromBlock)
		assert.Equal(t, valuespb.NewIntFromBigInt(toBlock), result.ToBlock)
		assert.ElementsMatch(t, [][20]byte{expectedAddr}, result.Addresses)
		assert.ElementsMatch(t, expectedTopics, result.Topics)
	})
}

func TestConvertAddressesFromProto(t *testing.T) {
	t.Run("empty input", func(t *testing.T) {
		addrs := evm.ConvertAddressesFromProto(nil)
		assert.Empty(t, addrs)
	})

	t.Run("invalid and valid addresses", func(t *testing.T) {
		valid := make([]byte, 20)
		for i := 0; i < 20; i++ {
			valid[i] = byte(i + 1)
		}
		invalid := []byte{0x01, 0x02}

		result := evm.ConvertAddressesFromProto([][]byte{valid, invalid})
		assert.Len(t, result, 1)
		assert.Equal(t, evmtypes.Address(valid), result[0])
	})
}

func TestConvertHashesFromProto(t *testing.T) {
	t.Run("empty input", func(t *testing.T) {
		hashes := evm.ConvertHashesFromProto(nil)
		assert.Empty(t, hashes)
	})

	t.Run("invalid and valid hashes", func(t *testing.T) {
		valid := make([]byte, 32)
		for i := 0; i < 32; i++ {
			valid[i] = byte(i + 10)
		}
		invalid := []byte{0xAA}

		result := evm.ConvertHashesFromProto([][]byte{valid, invalid})
		assert.Len(t, result, 1)
		assert.Equal(t, evmtypes.Hash(valid), result[0])
	})
}

func TestConvertTopicsFromProto(t *testing.T) {
	t.Run("single topic with one hash", func(t *testing.T) {
		topicBytes := make([]byte, 32)
		for i := 0; i < 32; i++ {
			topicBytes[i] = byte(i + 100)
		}
		input := []*evm.Topics{
			{Topic: [][]byte{topicBytes}},
		}
		result := evm.ConvertTopicsFromProto(input)
		assert.Len(t, result, 1)
		assert.Len(t, result[0], 1)
		assert.Equal(t, evmtypes.Hash(topicBytes), result[0][0])
	})
}

func TestConvertHeadFromProto(t *testing.T) {
	t.Run("nil input returns error", func(t *testing.T) {
		_, err := evm.ConvertHeadFromProto(nil)
		assert.ErrorContains(t, err, "head is nil")
	})

	t.Run("valid head", func(t *testing.T) {
		hash := make([]byte, 32)
		parent := make([]byte, 32)
		for i := 0; i < 32; i++ {
			hash[i] = byte(i + 30)
			parent[i] = byte(i + 60)
		}
		num := &valuespb.BigInt{AbsVal: []byte{0x01, 0x02, 0x03}, Sign: 1}
		timestamp := uint64(42)

		proto := &evm.Head{
			Timestamp:   timestamp,
			BlockNumber: num,
			Hash:        hash,
			ParentHash:  parent,
		}

		result, err := evm.ConvertHeadFromProto(proto)
		assert.NoError(t, err)
		assert.Equal(t, timestamp, result.Timestamp)
		assert.Equal(t, evmtypes.Hash(hash), result.Hash)
		assert.Equal(t, evmtypes.Hash(parent), result.ParentHash)
		assert.Equal(t, valuespb.NewIntFromBigInt(num), result.Number)
	})
}

func TestConvertReceiptFromProto(t *testing.T) {
	t.Run("nil input returns error", func(t *testing.T) {
		_, err := evm.ConvertReceiptFromProto(nil)
		assert.ErrorContains(t, err, "receipt is nil")
	})

	t.Run("valid receipt", func(t *testing.T) {
		txHash := make([]byte, 32)
		addr := make([]byte, 20)
		blockHash := make([]byte, 32)
		for i := range txHash {
			txHash[i] = byte(i + 1)
		}
		for i := range addr {
			addr[i] = byte(i + 50)
		}
		for i := range blockHash {
			blockHash[i] = byte(i + 100)
		}

		num := &valuespb.BigInt{AbsVal: []byte{0x01, 0x02, 0x03}, Sign: 1}
		price := &valuespb.BigInt{AbsVal: []byte{0x0A, 0x0B}, Sign: 1}

		proto := &evm.Receipt{
			Status:            1,
			Logs:              []*evm.Log{},
			TxHash:            txHash,
			ContractAddress:   addr,
			GasUsed:           21000,
			BlockHash:         blockHash,
			BlockNumber:       num,
			TxIndex:           3,
			EffectiveGasPrice: price,
		}

		result, err := evm.ConvertReceiptFromProto(proto)
		assert.NoError(t, err)
		assert.Equal(t, evmtypes.Hash(txHash), result.TxHash)
		assert.Equal(t, evmtypes.Address(addr), result.ContractAddress)
		assert.Equal(t, evmtypes.Hash(blockHash), result.BlockHash)
		assert.Equal(t, valuespb.NewIntFromBigInt(num), result.BlockNumber)
		assert.Equal(t, valuespb.NewIntFromBigInt(price), result.EffectiveGasPrice)
	})
}

func TestConvertTransactionFromProto(t *testing.T) {
	t.Run("nil input returns error", func(t *testing.T) {
		_, err := evm.ConvertTransactionFromProto(nil)
		assert.ErrorContains(t, err, "transaction is nil")
	})

	t.Run("valid transaction", func(t *testing.T) {
		to := make([]byte, 20)
		hash := make([]byte, 32)
		for i := 0; i < 20; i++ {
			to[i] = byte(i + 9)
		}
		for i := 0; i < 32; i++ {
			hash[i] = byte(i + 33)
		}

		gasPrice := &valuespb.BigInt{AbsVal: []byte{0x05, 0x06}, Sign: 1}
		value := &valuespb.BigInt{AbsVal: []byte{0x07, 0x08}, Sign: 1}

		proto := &evm.Transaction{
			To:       to,
			Data:     []byte{0x01, 0x02},
			Hash:     hash,
			Nonce:    1,
			Gas:      30000,
			GasPrice: gasPrice,
			Value:    value,
		}

		result, err := evm.ConvertTransactionFromProto(proto)
		assert.NoError(t, err)
		assert.Equal(t, evmtypes.Address(to), result.To)
		assert.Equal(t, []byte{0x01, 0x02}, result.Data)
		assert.Equal(t, evmtypes.Hash(hash), result.Hash)
		assert.Equal(t, uint64(1), result.Nonce)
		assert.Equal(t, uint64(30000), result.Gas)
		assert.Equal(t, valuespb.NewIntFromBigInt(gasPrice), result.GasPrice)
		assert.Equal(t, valuespb.NewIntFromBigInt(value), result.Value)
	})
}
