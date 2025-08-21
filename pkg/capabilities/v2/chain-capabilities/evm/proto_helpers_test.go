package evm_test

import (
	"testing"

	protoevm "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/chain-capabilities/evm"
	chainevm "github.com/smartcontractkit/chainlink-common/pkg/chains/evm"
	evmtypes "github.com/smartcontractkit/chainlink-common/pkg/types/chains/evm"
	valuespb "github.com/smartcontractkit/chainlink-protos/cre/go/values/pb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func h32(b byte) evmtypes.Hash {
	var h [32]byte
	for i := 0; i < 32; i++ {
		h[i] = b + byte(i)
	}
	return h
}
func a20(b byte) evmtypes.Address {
	var a [20]byte
	for i := 0; i < 20; i++ {
		a[i] = b + byte(i)
	}
	return a
}
func b32(b byte) []byte {
	out := make([]byte, 32)
	for i := 0; i < 32; i++ {
		out[i] = b + byte(i)
	}
	return out
}
func b20(b byte) []byte {
	out := make([]byte, 20)
	for i := 0; i < 20; i++ {
		out[i] = b + byte(i)
	}
	return out
}

func TestConvertHeaderToProto(t *testing.T) {
	t.Run("nil header returns error", func(t *testing.T) {
		_, err := protoevm.ConvertHeaderToProto(nil)
		assert.ErrorIs(t, err, chainevm.ErrEmptyHead)
	})

	t.Run("happy path", func(t *testing.T) {
		h := &evmtypes.Header{
			Timestamp:  123456,
			Hash:       h32(0x10),
			ParentHash: h32(0x11),
			Number:     valuespb.NewIntFromBigInt(&valuespb.BigInt{AbsVal: []byte{0x01}, Sign: 0}),
		}
		p, err := protoevm.ConvertHeaderToProto(h)
		require.NoError(t, err)
		assert.Equal(t, h.Timestamp, p.Timestamp)
		assert.Equal(t, h.Hash[:], p.Hash)
		assert.Equal(t, h.ParentHash[:], p.ParentHash)
		assert.Equal(t, valuespb.NewBigIntFromInt(h.Number), p.BlockNumber)
	})
}

func TestConvertHeaderFromProto(t *testing.T) {
	t.Run("nil proto returns error", func(t *testing.T) {
		_, err := protoevm.ConvertHeaderFromProto(nil)
		assert.ErrorIs(t, err, chainevm.ErrEmptyHead)
	})

	t.Run("happy path", func(t *testing.T) {
		num := &valuespb.BigInt{AbsVal: []byte{0x01}, Sign: 0}
		p := &protoevm.Header{
			Timestamp:   42,
			BlockNumber: num,
			Hash:        b32(0x20),
			ParentHash:  b32(0x21),
		}
		h, err := protoevm.ConvertHeaderFromProto(p)
		require.NoError(t, err)
		assert.Equal(t, uint64(42), h.Timestamp)
		assert.Equal(t, evmtypes.Hash(b32(0x20)), h.Hash)
		assert.Equal(t, evmtypes.Hash(b32(0x21)), h.ParentHash)
		assert.Equal(t, valuespb.NewIntFromBigInt(num), h.Number)
	})
}

func TestConvertReceiptToProto(t *testing.T) {
	t.Run("nil receipt returns error", func(t *testing.T) {
		_, err := protoevm.ConvertReceiptToProto(nil)
		assert.ErrorIs(t, err, chainevm.ErrEmptyReceipt)
	})

	t.Run("happy path", func(t *testing.T) {
		r := &evmtypes.Receipt{
			Status:            1,
			Logs:              []*evmtypes.Log{{}},
			TxHash:            h32(0x30),
			ContractAddress:   a20(0x31),
			GasUsed:           21000,
			BlockHash:         h32(0x32),
			BlockNumber:       valuespb.NewIntFromBigInt(&valuespb.BigInt{AbsVal: []byte{0x02}, Sign: 0}),
			TransactionIndex:  9,
			EffectiveGasPrice: valuespb.NewIntFromBigInt(&valuespb.BigInt{AbsVal: []byte{0x03}, Sign: 0}),
		}
		p, err := protoevm.ConvertReceiptToProto(r)
		require.NoError(t, err)
		assert.Equal(t, r.Status, p.Status)
		assert.Equal(t, r.TxHash[:], p.TxHash)
		assert.Equal(t, r.ContractAddress[:], p.ContractAddress)
		assert.Equal(t, r.GasUsed, p.GasUsed)
		assert.Equal(t, r.BlockHash[:], p.BlockHash)
		assert.Equal(t, valuespb.NewBigIntFromInt(r.BlockNumber), p.BlockNumber)
		assert.Equal(t, r.TransactionIndex, p.TxIndex)
		assert.Equal(t, valuespb.NewBigIntFromInt(r.EffectiveGasPrice), p.EffectiveGasPrice)
	})
}

func TestConvertReceiptFromProto(t *testing.T) {
	t.Run("nil proto returns error", func(t *testing.T) {
		_, err := protoevm.ConvertReceiptFromProto(nil)
		assert.ErrorIs(t, err, chainevm.ErrEmptyReceipt)
	})

	t.Run("empty logs slice ok + happy path", func(t *testing.T) {
		num := &valuespb.BigInt{AbsVal: []byte{0x01}, Sign: 0}
		price := &valuespb.BigInt{AbsVal: []byte{0x02}, Sign: 0}
		p := &protoevm.Receipt{
			Status:            1,
			Logs:              []*protoevm.Log{},
			TxHash:            b32(0x40),
			ContractAddress:   b20(0x41),
			GasUsed:           21000,
			BlockHash:         b32(0x42),
			BlockNumber:       num,
			TxIndex:           3,
			EffectiveGasPrice: price,
		}
		r, err := protoevm.ConvertReceiptFromProto(p)
		require.NoError(t, err)
		assert.Equal(t, evmtypes.Hash(b32(0x40)), r.TxHash)
		assert.Equal(t, evmtypes.Address(b20(0x41)), r.ContractAddress)
		assert.Equal(t, evmtypes.Hash(b32(0x42)), r.BlockHash)
		assert.Equal(t, valuespb.NewIntFromBigInt(num), r.BlockNumber)
		assert.Equal(t, valuespb.NewIntFromBigInt(price), r.EffectiveGasPrice)
		assert.Equal(t, uint64(21000), r.GasUsed)
		assert.Equal(t, uint64(3), r.TransactionIndex)
	})
}

func TestConvertTransactionToProto(t *testing.T) {
	t.Run("nil tx returns error", func(t *testing.T) {
		_, err := protoevm.ConvertTransactionToProto(nil)
		assert.ErrorIs(t, err, chainevm.ErrEmptyTx)
	})

	t.Run("happy path", func(t *testing.T) {
		tx := &evmtypes.Transaction{
			To:       a20(0x50),
			Data:     []byte{0xDE, 0xAD},
			Hash:     h32(0x51),
			Nonce:    7,
			Gas:      50000,
			GasPrice: valuespb.NewIntFromBigInt(&valuespb.BigInt{AbsVal: []byte{0x0F}, Sign: 1}),
			Value:    valuespb.NewIntFromBigInt(&valuespb.BigInt{AbsVal: []byte{0x0E}, Sign: 1}),
		}
		p, err := protoevm.ConvertTransactionToProto(tx)
		require.NoError(t, err)
		assert.Equal(t, tx.To[:], p.To)
		assert.Equal(t, tx.Data, p.Data)
		assert.Equal(t, tx.Hash[:], p.Hash)
		assert.Equal(t, tx.Nonce, p.Nonce)
		assert.Equal(t, tx.Gas, p.Gas)
		assert.Equal(t, valuespb.NewBigIntFromInt(tx.GasPrice), p.GasPrice)
		assert.Equal(t, valuespb.NewBigIntFromInt(tx.Value), p.Value)
	})
}

func TestConvertTransactionFromProto(t *testing.T) {
	t.Run("nil proto returns error", func(t *testing.T) {
		_, err := protoevm.ConvertTransactionFromProto(nil)
		assert.ErrorIs(t, err, chainevm.ErrEmptyTx)
	})

	t.Run("happy path", func(t *testing.T) {
		gp := &valuespb.BigInt{AbsVal: []byte{0x05}, Sign: 1}
		val := &valuespb.BigInt{AbsVal: []byte{0x06}, Sign: 1}
		p := &protoevm.Transaction{
			To:       b20(0x60),
			Data:     []byte{0x01, 0x02},
			Hash:     b32(0x61),
			Nonce:    1,
			Gas:      30000,
			GasPrice: gp,
			Value:    val,
		}
		tx, err := protoevm.ConvertTransactionFromProto(p)
		require.NoError(t, err)
		assert.Equal(t, evmtypes.Address(b20(0x60)), tx.To)
		assert.Equal(t, []byte{0x01, 0x02}, tx.Data)
		assert.Equal(t, evmtypes.Hash(b32(0x61)), tx.Hash)
		assert.Equal(t, uint64(1), tx.Nonce)
		assert.Equal(t, uint64(30000), tx.Gas)
		assert.Equal(t, valuespb.NewIntFromBigInt(gp), tx.GasPrice)
		assert.Equal(t, valuespb.NewIntFromBigInt(val), tx.Value)
	})

	t.Run("optional To nil is accepted", func(t *testing.T) {
		p := &protoevm.Transaction{
			Hash:     b32(0x62),
			GasPrice: &valuespb.BigInt{AbsVal: []byte{0x01}, Sign: 1},
			Value:    &valuespb.BigInt{AbsVal: []byte{0x02}, Sign: 1},
		}
		tx, err := protoevm.ConvertTransactionFromProto(p)
		require.NoError(t, err)

		var zero evmtypes.Address
		assert.Equal(t, zero, tx.To)
	})
}
