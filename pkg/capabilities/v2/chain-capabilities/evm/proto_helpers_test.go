package evm_test

import (
	"testing"
	"time"

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
func zero20() []byte { return make([]byte, 20) }

// Compares two *pb.BigInt by numeric value, ignoring internal Sign normalization.
func assertBigIntProtoEqual(t *testing.T, a, b *valuespb.BigInt) {
	t.Helper()
	assert.Equal(t, valuespb.NewIntFromBigInt(a), valuespb.NewIntFromBigInt(b))
}

func TestHeader_Conversions(t *testing.T) {
	t.Run("nil guards", func(t *testing.T) {
		_, err := protoevm.ConvertHeaderToProto(nil)
		require.ErrorIs(t, err, chainevm.ErrEmptyHead)
		_, err = protoevm.ConvertHeaderFromProto(nil)
		require.ErrorIs(t, err, chainevm.ErrEmptyHead)
	})

	t.Run("roundtrip to proto", func(t *testing.T) {
		h := &evmtypes.Header{
			Timestamp:  123456,
			Hash:       h32(0x10),
			ParentHash: h32(0x11),
			Number:     valuespb.NewIntFromBigInt(&valuespb.BigInt{AbsVal: []byte{0x01}, Sign: 0}),
		}
		p, err := protoevm.ConvertHeaderToProto(h)
		require.NoError(t, err)
		back, err := protoevm.ConvertHeaderFromProto(p)
		require.NoError(t, err)

		assert.Equal(t, h.Timestamp, back.Timestamp)
		assert.Equal(t, h.Hash, back.Hash)
		assert.Equal(t, h.ParentHash, back.ParentHash)
		assert.Equal(t, h.Number, back.Number)
	})

	t.Run("roundtrip from proto", func(t *testing.T) {
		num := &valuespb.BigInt{AbsVal: []byte{0x02}, Sign: 0}
		p := &protoevm.Header{
			Timestamp:   42,
			BlockNumber: num,
			Hash:        b32(0x20),
			ParentHash:  b32(0x21),
		}
		d, err := protoevm.ConvertHeaderFromProto(p)
		require.NoError(t, err)
		p2, err := protoevm.ConvertHeaderToProto(&d)
		require.NoError(t, err)

		assert.Equal(t, p.Timestamp, p2.Timestamp)
		assert.Equal(t, p.Hash, p2.Hash)
		assert.Equal(t, p.ParentHash, p2.ParentHash)
		assertBigIntProtoEqual(t, p.BlockNumber, p2.BlockNumber)
	})
}

func TestTransaction_Conversions(t *testing.T) {
	t.Run("nil guards", func(t *testing.T) {
		_, err := protoevm.ConvertTransactionToProto(nil)
		require.ErrorIs(t, err, chainevm.ErrEmptyTx)
		_, err = protoevm.ConvertTransactionFromProto(nil)
		require.ErrorIs(t, err, chainevm.ErrEmptyTx)
	})

	t.Run("roundtrip to proto", func(t *testing.T) {
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
		back, err := protoevm.ConvertTransactionFromProto(p)
		require.NoError(t, err)

		assert.Equal(t, tx.To, back.To)
		assert.Equal(t, tx.Data, back.Data)
		assert.Equal(t, tx.Hash, back.Hash)
		assert.Equal(t, tx.Nonce, back.Nonce)
		assert.Equal(t, tx.Gas, back.Gas)
		assert.Equal(t, tx.GasPrice, back.GasPrice)
		assert.Equal(t, tx.Value, back.Value)
	})

	t.Run("roundtrip from proto (nil To, nil Data allowed)", func(t *testing.T) {
		gp := &valuespb.BigInt{AbsVal: []byte{0x05}, Sign: 1}
		val := &valuespb.BigInt{AbsVal: []byte{0x06}, Sign: 1}
		p := &protoevm.Transaction{
			To:       nil, // optional
			Data:     nil, // allowed
			Hash:     b32(0x61),
			Nonce:    1,
			Gas:      30000,
			GasPrice: gp,
			Value:    val,
		}
		d, err := protoevm.ConvertTransactionFromProto(p)
		require.NoError(t, err)
		require.Nil(t, d.Data) // ensure nil stays nil
		p2, err := protoevm.ConvertTransactionToProto(d)
		require.NoError(t, err)

		assert.Equal(t, p.Hash, p2.Hash)
		assert.Equal(t, p.Nonce, p2.Nonce)
		assert.Equal(t, p.Gas, p2.Gas)
		assertBigIntProtoEqual(t, p.GasPrice, p2.GasPrice)
		assertBigIntProtoEqual(t, p.Value, p2.Value)
		// To should be zero-address bytes after roundtrip
		require.Len(t, p2.To, 20)
		assert.Equal(t, zero20(), p2.To)
	})
}

func TestCallMsg_Conversions(t *testing.T) {
	t.Run("nil guards", func(t *testing.T) {
		_, err := protoevm.ConvertCallMsgToProto(nil)
		require.ErrorIs(t, err, chainevm.ErrEmptyMsg)
		_, err = protoevm.ConvertCallMsgFromProto(nil)
		require.ErrorIs(t, err, chainevm.ErrEmptyMsg)
	})

	t.Run("roundtrip to proto", func(t *testing.T) {
		msg := &evmtypes.CallMsg{From: a20(0x01), To: a20(0x02), Data: []byte{1, 2, 3}}
		p, err := protoevm.ConvertCallMsgToProto(msg)
		require.NoError(t, err)
		back, err := protoevm.ConvertCallMsgFromProto(p)
		require.NoError(t, err)

		assert.Equal(t, msg.From, back.From)
		assert.Equal(t, msg.To, back.To)
		assert.Equal(t, msg.Data, back.Data)
	})

	t.Run("roundtrip from proto (optional To nil)", func(t *testing.T) {
		p := &protoevm.CallMsg{From: b20(0x01), To: nil, Data: []byte{0xAA}}
		d, err := protoevm.ConvertCallMsgFromProto(p)
		require.NoError(t, err)
		p2, err := protoevm.ConvertCallMsgToProto(d)
		require.NoError(t, err)

		assert.Equal(t, p.From, p2.From)
		assert.Equal(t, p.Data, p2.Data)
		// To should become zero-address bytes after roundtrip
		require.Len(t, p2.To, 20)
		assert.Equal(t, zero20(), p2.To)
	})
}

func TestReceipt_Conversions(t *testing.T) {
	t.Run("nil guards", func(t *testing.T) {
		_, err := protoevm.ConvertReceiptToProto(nil)
		require.ErrorIs(t, err, chainevm.ErrEmptyReceipt)
		_, err = protoevm.ConvertReceiptFromProto(nil)
		require.ErrorIs(t, err, chainevm.ErrEmptyReceipt)
	})

	t.Run("roundtrip to proto", func(t *testing.T) {
		r := &evmtypes.Receipt{
			Status: 1,
			Logs: []*evmtypes.Log{{
				LogIndex:    7,
				BlockHash:   h32(0xA0),
				BlockNumber: valuespb.NewIntFromBigInt(&valuespb.BigInt{AbsVal: []byte{0x09}, Sign: 0}),
				Topics:      []evmtypes.Hash{h32(0xA1), h32(0xA2)},
				EventSig:    h32(0xA3),
				Address:     a20(0xA4),
				TxHash:      h32(0xA5),
				Data:        []byte{0xDE, 0xAD, 0xBE, 0xEF},
				Removed:     true,
			}},
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
		back, err := protoevm.ConvertReceiptFromProto(p)
		require.NoError(t, err)

		assert.Equal(t, r.Status, back.Status)
		assert.Equal(t, r.TxHash, back.TxHash)
		assert.Equal(t, r.ContractAddress, back.ContractAddress)
		assert.Equal(t, r.GasUsed, back.GasUsed)
		assert.Equal(t, r.BlockHash, back.BlockHash)
		assert.Equal(t, r.BlockNumber, back.BlockNumber)
		assert.Equal(t, r.TransactionIndex, back.TransactionIndex)
		assert.Equal(t, r.EffectiveGasPrice, back.EffectiveGasPrice)

		require.Len(t, back.Logs, 1)
		gl := back.Logs[0]
		ol := r.Logs[0]
		assert.Equal(t, ol.LogIndex, gl.LogIndex)
		assert.Equal(t, ol.BlockHash, gl.BlockHash)
		assert.Equal(t, ol.BlockNumber, gl.BlockNumber)
		assert.Equal(t, ol.Topics, gl.Topics)
		assert.Equal(t, ol.EventSig, gl.EventSig)
		assert.Equal(t, ol.Address, gl.Address)
		assert.Equal(t, ol.TxHash, gl.TxHash)
		assert.Equal(t, ol.Data, gl.Data)
		assert.Equal(t, ol.Removed, gl.Removed)
	})

	t.Run("roundtrip from proto", func(t *testing.T) {
		num := &valuespb.BigInt{AbsVal: []byte{0x01}, Sign: 0}
		price := &valuespb.BigInt{AbsVal: []byte{0x02}, Sign: 0}
		plog := &protoevm.Log{
			Index:       5,
			BlockHash:   b32(0xB0),
			BlockNumber: &valuespb.BigInt{AbsVal: []byte{0x04}, Sign: 0},
			Topics:      [][]byte{b32(0xB1), b32(0xB2)},
			EventSig:    b32(0xB3),
			Address:     b20(0xB4),
			TxHash:      b32(0xB5),
			Data:        []byte{0xBE, 0xEF},
			Removed:     false,
		}
		p := &protoevm.Receipt{
			Status:            1,
			Logs:              []*protoevm.Log{plog},
			TxHash:            b32(0x40),
			ContractAddress:   b20(0x41),
			GasUsed:           21000,
			BlockHash:         b32(0x42),
			BlockNumber:       num,
			TxIndex:           3,
			EffectiveGasPrice: price,
		}
		d, err := protoevm.ConvertReceiptFromProto(p)
		require.NoError(t, err)
		p2, err := protoevm.ConvertReceiptToProto(d)
		require.NoError(t, err)

		assert.Equal(t, p.Status, p2.Status)
		assert.Equal(t, p.TxHash, p2.TxHash)
		assert.Equal(t, p.ContractAddress, p2.ContractAddress)
		assert.Equal(t, p.GasUsed, p2.GasUsed)
		assert.Equal(t, p.BlockHash, p2.BlockHash)
		assertBigIntProtoEqual(t, p.BlockNumber, p2.BlockNumber)
		assert.Equal(t, p.TxIndex, p2.TxIndex)
		assertBigIntProtoEqual(t, p.EffectiveGasPrice, p2.EffectiveGasPrice)

		require.Len(t, p2.Logs, 1)
		l2 := p2.Logs[0]
		assert.Equal(t, plog.Index, l2.Index)
		assert.Equal(t, plog.BlockHash, l2.BlockHash)
		assertBigIntProtoEqual(t, plog.BlockNumber, l2.BlockNumber)
		assert.Equal(t, plog.Topics, l2.Topics)
		assert.Equal(t, plog.EventSig, l2.EventSig)
		assert.Equal(t, plog.Address, l2.Address)
		assert.Equal(t, plog.TxHash, l2.TxHash)
		assert.Equal(t, plog.Data, l2.Data)
		assert.Equal(t, plog.Removed, l2.Removed)
	})
}

func TestFilterQuery_Conversions(t *testing.T) {
	t.Run("nil guard FromProto", func(t *testing.T) {
		_, err := protoevm.ConvertFilterFromProto(nil)
		require.ErrorIs(t, err, chainevm.ErrEmptyFilter)
	})

	t.Run("roundtrip to proto", func(t *testing.T) {
		d := evmtypes.FilterQuery{
			BlockHash: h32(0x90),
			FromBlock: valuespb.NewIntFromBigInt(&valuespb.BigInt{AbsVal: []byte{0x01}, Sign: 0}),
			ToBlock:   valuespb.NewIntFromBigInt(&valuespb.BigInt{AbsVal: []byte{0x02}, Sign: 0}),
			Addresses: []evmtypes.Address{a20(0x91), a20(0x92)},
			Topics: [][]evmtypes.Hash{
				{h32(0x93)},
				{}, // empty (non-nil) is allowed
			},
		}
		p, err := protoevm.ConvertFilterToProto(d)
		require.NoError(t, err)
		back, err := protoevm.ConvertFilterFromProto(p)
		require.NoError(t, err)

		assert.Equal(t, d.BlockHash, back.BlockHash)
		assert.Equal(t, d.FromBlock, back.FromBlock)
		assert.Equal(t, d.ToBlock, back.ToBlock)
		assert.Equal(t, d.Addresses, back.Addresses)
		assert.Equal(t, d.Topics, back.Topics)
	})

	t.Run("roundtrip from proto", func(t *testing.T) {
		p := &protoevm.FilterQuery{
			BlockHash: b32(0x9A),
			FromBlock: &valuespb.BigInt{AbsVal: []byte{0x03}, Sign: 0},
			ToBlock:   &valuespb.BigInt{AbsVal: []byte{0x04}, Sign: 0},
			Addresses: [][]byte{b20(0x9B), b20(0x9C)},
			Topics: []*protoevm.Topics{
				{Topic: [][]byte{b32(0x9D)}},
				{Topic: [][]byte{}}, // empty inner slice is OK
			},
		}
		d, err := protoevm.ConvertFilterFromProto(p)
		require.NoError(t, err)
		p2, err := protoevm.ConvertFilterToProto(d)
		require.NoError(t, err)

		assert.Equal(t, p.BlockHash, p2.BlockHash)
		assertBigIntProtoEqual(t, p.FromBlock, p2.FromBlock)
		assertBigIntProtoEqual(t, p.ToBlock, p2.ToBlock)
		assert.Equal(t, p.Addresses, p2.Addresses)
		require.Len(t, p2.Topics, len(p.Topics))
		for i := range p.Topics {
			assert.Equal(t, p.Topics[i].Topic, p2.Topics[i].Topic)
		}
	})
}

func TestLPFilter_Conversions(t *testing.T) {
	t.Run("roundtrip to proto", func(t *testing.T) {
		in := evmtypes.LPFilterQuery{
			Name:         "my-filter",
			Retention:    5 * time.Minute,
			Addresses:    []evmtypes.Address{a20(0x01)},
			EventSigs:    []evmtypes.Hash{h32(0x02)},
			Topic2:       []evmtypes.Hash{h32(0x03)},
			Topic3:       []evmtypes.Hash{h32(0x04)},
			Topic4:       []evmtypes.Hash{h32(0x05)},
			MaxLogsKept:  123,
			LogsPerBlock: 7,
		}
		p := protoevm.ConvertLPFilterToProto(in)
		out, err := protoevm.ConvertLPFilterFromProto(p)
		require.NoError(t, err)

		assert.Equal(t, in.Name, out.Name)
		assert.Equal(t, in.Retention, out.Retention)
		assert.Equal(t, in.MaxLogsKept, out.MaxLogsKept)
		assert.Equal(t, in.LogsPerBlock, out.LogsPerBlock)
		assert.Equal(t, in.Addresses, out.Addresses)
		assert.Equal(t, in.EventSigs, out.EventSigs)
		assert.Equal(t, in.Topic2, out.Topic2)
		assert.Equal(t, in.Topic3, out.Topic3)
		assert.Equal(t, in.Topic4, out.Topic4)
	})

	t.Run("roundtrip from proto", func(t *testing.T) {
		p := &protoevm.LPFilter{
			Name:          "other-filter",
			RetentionTime: int64(2 * time.Hour),
			Addresses:     [][]byte{b20(0x11), b20(0x12)},
			EventSigs:     [][]byte{b32(0x21)},
			Topic2:        [][]byte{b32(0x22)},
			Topic3:        [][]byte{b32(0x23)},
			Topic4:        [][]byte{b32(0x24)},
			MaxLogsKept:   456,
			LogsPerBlock:  3,
		}
		d, err := protoevm.ConvertLPFilterFromProto(p)
		require.NoError(t, err)
		p2 := protoevm.ConvertLPFilterToProto(d)

		assert.Equal(t, p.Name, p2.Name)
		assert.Equal(t, p.RetentionTime, p2.RetentionTime)
		assert.Equal(t, p.Addresses, p2.Addresses)
		assert.Equal(t, p.EventSigs, p2.EventSigs)
		assert.Equal(t, p.Topic2, p2.Topic2)
		assert.Equal(t, p.Topic3, p2.Topic3)
		assert.Equal(t, p.Topic4, p2.Topic4)
		assert.Equal(t, p.MaxLogsKept, p2.MaxLogsKept)
		assert.Equal(t, p.LogsPerBlock, p2.LogsPerBlock)
	})
}

func TestLog_Conversions(t *testing.T) {
	t.Run("roundtrip to proto", func(t *testing.T) {
		l := evmtypes.Log{
			LogIndex:    7,
			BlockHash:   h32(0xA0),
			BlockNumber: valuespb.NewIntFromBigInt(&valuespb.BigInt{AbsVal: []byte{0x09}, Sign: 0}),
			Topics:      []evmtypes.Hash{h32(0xA1), h32(0xA2)},
			EventSig:    h32(0xA3),
			Address:     a20(0xA4),
			TxHash:      h32(0xA5),
			Data:        []byte{0xDE, 0xAD, 0xBE, 0xEF},
			Removed:     true,
		}
		pl := protoevm.ConvertLogToProto(l)
		backLogs, err := protoevm.ConvertLogsFromProto([]*protoevm.Log{pl})
		require.NoError(t, err)
		require.Len(t, backLogs, 1)
		got := backLogs[0]

		assert.Equal(t, l.LogIndex, got.LogIndex)
		assert.Equal(t, l.BlockHash, got.BlockHash)
		assert.Equal(t, l.BlockNumber, got.BlockNumber)
		assert.Equal(t, l.Topics, got.Topics)
		assert.Equal(t, l.EventSig, got.EventSig)
		assert.Equal(t, l.Address, got.Address)
		assert.Equal(t, l.TxHash, got.TxHash)
		assert.Equal(t, l.Data, got.Data)
		assert.Equal(t, l.Removed, got.Removed)
	})

	t.Run("roundtrip from proto", func(t *testing.T) {
		pl := &protoevm.Log{
			Index:       5,
			BlockHash:   b32(0xB0),
			BlockNumber: &valuespb.BigInt{AbsVal: []byte{0x04}, Sign: 0},
			Topics:      [][]byte{b32(0xB1), b32(0xB2)},
			EventSig:    b32(0xB3),
			Address:     b20(0xB4),
			TxHash:      b32(0xB5),
			Data:        []byte{0xBE, 0xEF},
			Removed:     false,
		}
		dLogs, err := protoevm.ConvertLogsFromProto([]*protoevm.Log{pl})
		require.NoError(t, err)
		pLogs, err := protoevm.ConvertLogsToProto(dLogs)
		require.NoError(t, err)
		require.Len(t, pLogs, 1)

		got := pLogs[0]
		assert.Equal(t, pl.Index, got.Index)
		assert.Equal(t, pl.BlockHash, got.BlockHash)
		assertBigIntProtoEqual(t, pl.BlockNumber, got.BlockNumber)
		assert.Equal(t, pl.Topics, got.Topics)
		assert.Equal(t, pl.EventSig, got.EventSig)
		assert.Equal(t, pl.Address, got.Address)
		assert.Equal(t, pl.TxHash, got.TxHash)
		assert.Equal(t, pl.Data, got.Data)
		assert.Equal(t, pl.Removed, got.Removed)
	})
}
