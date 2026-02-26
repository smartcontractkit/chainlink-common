package aptos_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	conv "github.com/smartcontractkit/chainlink-common/pkg/chains/aptos"
	typeaptos "github.com/smartcontractkit/chainlink-common/pkg/types/chains/aptos"
)

func mkBytes(n int, fill byte) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = fill
	}
	return b
}

func TestViewPayloadConverters(t *testing.T) {
	t.Run("Roundtrip ViewPayload with simple types", func(t *testing.T) {
		addr := mkBytes(typeaptos.AccountAddressLength, 0x01)

		domainPayload := &typeaptos.ViewPayload{
			Module: typeaptos.ModuleID{
				Address: [32]byte(addr),
				Name:    "aptos_account",
			},
			Function: "transfer",
			ArgTypes: []typeaptos.TypeTag{
				{Value: typeaptos.U64Tag{}},
				{Value: typeaptos.AddressTag{}},
			},
			Args: [][]byte{
				{0x01, 0x02, 0x03},
				{0x04, 0x05, 0x06},
			},
		}

		// To proto
		protoPayload, err := conv.ConvertViewPayloadToProto(domainPayload)
		require.NoError(t, err)
		require.Equal(t, "aptos_account", protoPayload.Module.Name)
		require.Len(t, protoPayload.ArgTypes, 2)
		require.Len(t, protoPayload.Args, 2)

		// From proto
		roundtrip, err := conv.ConvertViewPayloadFromProto(protoPayload)
		require.NoError(t, err)
		require.Equal(t, "aptos_account", roundtrip.Module.Name)
		require.Equal(t, "transfer", roundtrip.Function)
		require.Len(t, roundtrip.ArgTypes, 2)
		require.Len(t, roundtrip.Args, 2)
		require.True(t, bytes.Equal(domainPayload.Args[0], roundtrip.Args[0]))
	})

	t.Run("Nil payload error", func(t *testing.T) {
		_, err := conv.ConvertViewPayloadFromProto(nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "proto payload is nil")

		_, err = conv.ConvertViewPayloadToProto(nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "payload is nil")
	})
}

func TestTypeTagConverters(t *testing.T) {
	t.Run("Simple type tags roundtrip", func(t *testing.T) {
		tags := []typeaptos.TypeTag{
			{Value: typeaptos.BoolTag{}},
			{Value: typeaptos.U8Tag{}},
			{Value: typeaptos.U16Tag{}},
			{Value: typeaptos.U32Tag{}},
			{Value: typeaptos.U64Tag{}},
			{Value: typeaptos.U128Tag{}},
			{Value: typeaptos.U256Tag{}},
			{Value: typeaptos.AddressTag{}},
			{Value: typeaptos.SignerTag{}},
		}

		for i, tag := range tags {
			proto, err := conv.ConvertTypeTagToProto(&tag)
			require.NoError(t, err, "failed to convert tag %d", i)

			roundtrip, err := conv.ConvertTypeTagFromProto(proto)
			require.NoError(t, err, "failed to convert back tag %d", i)
			require.Equal(t, tag.Value.TypeTagType(), roundtrip.Value.TypeTagType())
		}
	})

	t.Run("VectorTag roundtrip", func(t *testing.T) {
		vectorTag := typeaptos.TypeTag{
			Value: typeaptos.VectorTag{
				ElementType: typeaptos.TypeTag{Value: typeaptos.U64Tag{}},
			},
		}

		proto, err := conv.ConvertTypeTagToProto(&vectorTag)
		require.NoError(t, err)
		require.NotNil(t, proto.GetVector())

		roundtrip, err := conv.ConvertTypeTagFromProto(proto)
		require.NoError(t, err)
		vec, ok := roundtrip.Value.(typeaptos.VectorTag)
		require.True(t, ok)
		require.Equal(t, typeaptos.TypeTagU64, vec.ElementType.Value.TypeTagType())
	})

	t.Run("StructTag roundtrip", func(t *testing.T) {
		addr := mkBytes(typeaptos.AccountAddressLength, 0x01)
		structTag := typeaptos.TypeTag{
			Value: typeaptos.StructTag{
				Address: [32]byte(addr),
				Module:  "coin",
				Name:    "Coin",
				TypeParams: []typeaptos.TypeTag{
					{Value: typeaptos.AddressTag{}},
				},
			},
		}

		proto, err := conv.ConvertTypeTagToProto(&structTag)
		require.NoError(t, err)
		require.NotNil(t, proto.GetStruct())
		require.Equal(t, "coin", proto.GetStruct().Module)
		require.Equal(t, "Coin", proto.GetStruct().Name)

		roundtrip, err := conv.ConvertTypeTagFromProto(proto)
		require.NoError(t, err)
		st, ok := roundtrip.Value.(typeaptos.StructTag)
		require.True(t, ok)
		require.Equal(t, "coin", st.Module)
		require.Equal(t, "Coin", st.Name)
		require.Len(t, st.TypeParams, 1)
	})

	t.Run("GenericTag roundtrip", func(t *testing.T) {
		genericTag := typeaptos.TypeTag{
			Value: typeaptos.GenericTag{Index: 42},
		}

		proto, err := conv.ConvertTypeTagToProto(&genericTag)
		require.NoError(t, err)
		require.NotNil(t, proto.GetGeneric())
		require.EqualValues(t, 42, proto.GetGeneric().Index)

		roundtrip, err := conv.ConvertTypeTagFromProto(proto)
		require.NoError(t, err)
		gen, ok := roundtrip.Value.(typeaptos.GenericTag)
		require.True(t, ok)
		require.EqualValues(t, 42, gen.Index)
	})

	t.Run("Nil type tag error", func(t *testing.T) {
		_, err := conv.ConvertTypeTagFromProto(nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "proto type tag is nil")

		_, err = conv.ConvertTypeTagToProto(nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "type tag or value is nil")
	})

	t.Run("Invalid struct address length", func(t *testing.T) {
		protoTag := &conv.TypeTag{
			Type: conv.TypeTagType_TYPE_TAG_STRUCT,
			Value: &conv.TypeTag_Struct{
				Struct: &conv.StructTag{
					Address:    mkBytes(typeaptos.AccountAddressLength-1, 0x01),
					Module:     "test",
					Name:       "Test",
					TypeParams: nil,
				},
			},
		}

		_, err := conv.ConvertTypeTagFromProto(protoTag)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid struct address length")
	})
}

func TestTransactionConverters(t *testing.T) {
	t.Run("Transaction roundtrip with all fields", func(t *testing.T) {
		version := uint64(12345)
		success := true
		tx := &typeaptos.Transaction{
			Type:    typeaptos.TransactionVariantUser,
			Hash:    "0xabcdef123456",
			Version: &version,
			Success: &success,
			Data:    []byte(`{"sender":"0x1"}`),
		}

		protoTx := conv.ConvertTransactionToProto(tx)
		require.NotNil(t, protoTx)
		require.Equal(t, "0xabcdef123456", protoTx.Hash)
		require.NotNil(t, protoTx.Version)
		require.Equal(t, uint64(12345), *protoTx.Version)
		require.NotNil(t, protoTx.Success)
		require.True(t, *protoTx.Success)

		roundtrip, err := conv.ConvertTransactionFromProto(protoTx)
		require.NoError(t, err)
		require.Equal(t, tx.Type, roundtrip.Type)
		require.Equal(t, tx.Hash, roundtrip.Hash)
		require.Equal(t, *tx.Version, *roundtrip.Version)
		require.Equal(t, *tx.Success, *roundtrip.Success)
		require.True(t, bytes.Equal(tx.Data, roundtrip.Data))
	})

	t.Run("Pending transaction (no version/success)", func(t *testing.T) {
		tx := &typeaptos.Transaction{
			Type:    typeaptos.TransactionVariantPending,
			Hash:    "0x999",
			Version: nil,
			Success: nil,
			Data:    []byte(`{"pending":true}`),
		}

		protoTx := conv.ConvertTransactionToProto(tx)
		require.NotNil(t, protoTx)
		require.Nil(t, protoTx.Version)
		require.Nil(t, protoTx.Success)

		roundtrip, err := conv.ConvertTransactionFromProto(protoTx)
		require.NoError(t, err)
		require.Equal(t, typeaptos.TransactionVariantPending, roundtrip.Type)
		require.Nil(t, roundtrip.Version)
		require.Nil(t, roundtrip.Success)
	})

	t.Run("TransactionVariant enum roundtrip", func(t *testing.T) {
		variants := []typeaptos.TransactionVariant{
			typeaptos.TransactionVariantPending,
			typeaptos.TransactionVariantUser,
			typeaptos.TransactionVariantGenesis,
			typeaptos.TransactionVariantBlockMetadata,
			typeaptos.TransactionVariantBlockEpilogue,
			typeaptos.TransactionVariantStateCheckpoint,
			typeaptos.TransactionVariantValidator,
			typeaptos.TransactionVariantUnknown,
		}

		for _, variant := range variants {
			proto := conv.ConvertTransactionVariantToProto(variant)
			roundtrip := conv.ConvertTransactionVariantFromProto(proto)
			require.Equal(t, variant, roundtrip)
		}
	})
}

func TestSubmitTransactionConverters(t *testing.T) {
	t.Run("SubmitTransactionRequest roundtrip with GasConfig", func(t *testing.T) {
		addr := mkBytes(typeaptos.AccountAddressLength, 0x01)
		req := &typeaptos.SubmitTransactionRequest{
			ReceiverModuleID: typeaptos.ModuleID{
				Address: [32]byte(addr),
				Name:    "receiver_module",
			},
			EncodedPayload: []byte{0x01, 0x02, 0x03, 0x04},
			GasConfig: &typeaptos.GasConfig{
				MaxGasAmount: 5000,
				GasUnitPrice: 100,
			},
		}

		protoReq, err := conv.ConvertSubmitTransactionRequestToProto(*req)
		require.NoError(t, err)
		require.Equal(t, "receiver_module", protoReq.ReceiverModuleId.Name)
		require.NotNil(t, protoReq.GasConfig)
		require.Equal(t, uint64(5000), protoReq.GasConfig.MaxGasAmount)
		require.Equal(t, uint64(100), protoReq.GasConfig.GasUnitPrice)

		roundtrip, err := conv.ConvertSubmitTransactionRequestFromProto(protoReq)
		require.NoError(t, err)
		require.Equal(t, "receiver_module", roundtrip.ReceiverModuleID.Name)
		require.NotNil(t, roundtrip.GasConfig)
		require.Equal(t, uint64(5000), roundtrip.GasConfig.MaxGasAmount)
		require.True(t, bytes.Equal(req.EncodedPayload, roundtrip.EncodedPayload))
	})

	t.Run("SubmitTransactionRequest without GasConfig", func(t *testing.T) {
		addr := mkBytes(typeaptos.AccountAddressLength, 0x02)
		req := &typeaptos.SubmitTransactionRequest{
			ReceiverModuleID: typeaptos.ModuleID{
				Address: [32]byte(addr),
				Name:    "test",
			},
			EncodedPayload: []byte{0xAA, 0xBB},
			GasConfig:      nil,
		}

		protoReq, err := conv.ConvertSubmitTransactionRequestToProto(*req)
		require.NoError(t, err)
		require.Nil(t, protoReq.GasConfig)

		roundtrip, err := conv.ConvertSubmitTransactionRequestFromProto(protoReq)
		require.NoError(t, err)
		require.Nil(t, roundtrip.GasConfig)
	})

	t.Run("SubmitTransactionReply roundtrip", func(t *testing.T) {
		reply := &typeaptos.SubmitTransactionReply{
			TxStatus:         typeaptos.TxSuccess,
			TxHash:           "0xabc123",
			TxIdempotencyKey: "key-456",
		}

		protoReply, err := conv.ConvertSubmitTransactionReplyToProto(reply)
		require.NoError(t, err)
		require.Equal(t, conv.TxStatus(typeaptos.TxSuccess), protoReply.TxStatus)
		require.Equal(t, "0xabc123", protoReply.TxHash)
		require.Equal(t, "key-456", protoReply.TxIdempotencyKey)

		roundtrip, err := conv.ConvertSubmitTransactionReplyFromProto(protoReply)
		require.NoError(t, err)
		require.Equal(t, reply.TxStatus, roundtrip.TxStatus)
		require.Equal(t, reply.TxHash, roundtrip.TxHash)
		require.Equal(t, reply.TxIdempotencyKey, roundtrip.TxIdempotencyKey)
	})

	t.Run("Invalid request errors", func(t *testing.T) {
		_, err := conv.ConvertSubmitTransactionRequestFromProto(nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "proto request is nil")

		protoReq := &conv.SubmitTransactionRequest{
			ReceiverModuleId: nil,
		}
		_, err = conv.ConvertSubmitTransactionRequestFromProto(protoReq)
		require.Error(t, err)
		require.Contains(t, err.Error(), "receiver module id is nil")

		protoReq.ReceiverModuleId = &conv.ModuleID{
			Address: mkBytes(typeaptos.AccountAddressLength-1, 0x01),
			Name:    "bad",
		}
		_, err = conv.ConvertSubmitTransactionRequestFromProto(protoReq)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid address length")
	})
}

func TestErrorJoinBehavior(t *testing.T) {
	t.Run("Aggregates multiple conversion errors", func(t *testing.T) {
		// Test that error wrapping works as expected
		protoPayload := &conv.ViewPayload{
			Module: &conv.ModuleID{
				Address: mkBytes(typeaptos.AccountAddressLength-5, 0x01),
				Name:    "bad",
			},
		}
		_, err := conv.ConvertViewPayloadFromProto(protoPayload)
		require.Error(t, err)
		require.True(t, errors.Is(err, err))
	})
}

func TestAccountTransactionsConverters(t *testing.T) {
	t.Run("Request roundtrip with start and limit", func(t *testing.T) {
		addr := mkBytes(typeaptos.AccountAddressLength, 0xAA)
		start := uint64(100)
		limit := uint64(25)
		req := typeaptos.AccountTransactionsRequest{
			Address: [32]byte(addr),
			Start:   &start,
			Limit:   &limit,
		}

		protoReq := conv.ConvertAccountTransactionsRequestToProto(req)
		require.Equal(t, addr, protoReq.Address)
		require.NotNil(t, protoReq.Start)
		require.Equal(t, uint64(100), *protoReq.Start)
		require.NotNil(t, protoReq.Limit)
		require.Equal(t, uint64(25), *protoReq.Limit)

		roundtrip, err := conv.ConvertAccountTransactionsRequestFromProto(protoReq)
		require.NoError(t, err)
		require.Equal(t, req.Address, roundtrip.Address)
		require.Equal(t, *req.Start, *roundtrip.Start)
		require.Equal(t, *req.Limit, *roundtrip.Limit)
	})

	t.Run("Request roundtrip without start and limit", func(t *testing.T) {
		addr := mkBytes(typeaptos.AccountAddressLength, 0xBB)
		req := typeaptos.AccountTransactionsRequest{
			Address: [32]byte(addr),
			Start:   nil,
			Limit:   nil,
		}

		protoReq := conv.ConvertAccountTransactionsRequestToProto(req)
		require.Nil(t, protoReq.Start)
		require.Nil(t, protoReq.Limit)

		roundtrip, err := conv.ConvertAccountTransactionsRequestFromProto(protoReq)
		require.NoError(t, err)
		require.Nil(t, roundtrip.Start)
		require.Nil(t, roundtrip.Limit)
	})

	t.Run("Request from proto nil error", func(t *testing.T) {
		_, err := conv.ConvertAccountTransactionsRequestFromProto(nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "proto request is nil")
	})

	t.Run("Request from proto invalid address length", func(t *testing.T) {
		protoReq := &conv.AccountTransactionsRequest{
			Address: mkBytes(10, 0x01),
		}
		_, err := conv.ConvertAccountTransactionsRequestFromProto(protoReq)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid address length")
	})

	t.Run("Reply roundtrip with multiple transactions", func(t *testing.T) {
		version1 := uint64(100)
		success1 := true
		version2 := uint64(200)
		success2 := false
		reply := &typeaptos.AccountTransactionsReply{
			Transactions: []*typeaptos.Transaction{
				{
					Type:    typeaptos.TransactionVariantUser,
					Hash:    "0xaaa",
					Version: &version1,
					Success: &success1,
					Data:    []byte(`{"tx":1}`),
				},
				{
					Type:    typeaptos.TransactionVariantUser,
					Hash:    "0xbbb",
					Version: &version2,
					Success: &success2,
					Data:    []byte(`{"tx":2}`),
				},
			},
		}

		protoReply := conv.ConvertAccountTransactionsReplyToProto(reply)
		require.Len(t, protoReply.Transactions, 2)
		require.Equal(t, "0xaaa", protoReply.Transactions[0].Hash)
		require.Equal(t, "0xbbb", protoReply.Transactions[1].Hash)

		roundtrip, err := conv.ConvertAccountTransactionsReplyFromProto(protoReply)
		require.NoError(t, err)
		require.Len(t, roundtrip.Transactions, 2)
		require.Equal(t, "0xaaa", roundtrip.Transactions[0].Hash)
		require.Equal(t, uint64(100), *roundtrip.Transactions[0].Version)
		require.True(t, *roundtrip.Transactions[0].Success)
		require.Equal(t, "0xbbb", roundtrip.Transactions[1].Hash)
		require.Equal(t, uint64(200), *roundtrip.Transactions[1].Version)
		require.False(t, *roundtrip.Transactions[1].Success)
	})

	t.Run("Reply roundtrip with empty transactions", func(t *testing.T) {
		reply := &typeaptos.AccountTransactionsReply{
			Transactions: []*typeaptos.Transaction{},
		}

		protoReply := conv.ConvertAccountTransactionsReplyToProto(reply)
		require.Len(t, protoReply.Transactions, 0)

		roundtrip, err := conv.ConvertAccountTransactionsReplyFromProto(protoReply)
		require.NoError(t, err)
		require.Len(t, roundtrip.Transactions, 0)
	})

	t.Run("Reply nil handling", func(t *testing.T) {
		result := conv.ConvertAccountTransactionsReplyToProto(nil)
		require.Nil(t, result)

		roundtrip, err := conv.ConvertAccountTransactionsReplyFromProto(nil)
		require.NoError(t, err)
		require.Nil(t, roundtrip)
	})
}

func TestNilHandling(t *testing.T) {
	t.Run("ConvertTransactionToProto with nil", func(t *testing.T) {
		result := conv.ConvertTransactionToProto(nil)
		require.Nil(t, result)
	})

	t.Run("ConvertTransactionFromProto with nil", func(t *testing.T) {
		result, err := conv.ConvertTransactionFromProto(nil)
		require.NoError(t, err)
		require.Nil(t, result)
	})
}
