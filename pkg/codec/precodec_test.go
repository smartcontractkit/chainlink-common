package codec_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/codec"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

var _ types.RemoteCodec = &ExampleCodec{}

type ExampleCodec struct {
	offChainType any
}

func (ec ExampleCodec) Encode(_ context.Context, item any, _ string) ([]byte, error) {
	return json.Marshal(item)
}

func (ec ExampleCodec) GetMaxEncodingSize(_ context.Context, n int, _ string) (int, error) {
	// not used in the example
	return math.MaxInt32, nil
}

func (ec ExampleCodec) Decode(_ context.Context, raw []byte, into any, _ string) error {
	err := json.Unmarshal(raw, into)
	if err != nil {
		return fmt.Errorf("%w: %w", types.ErrInvalidType, err)
	}
	return nil
}

func (ec ExampleCodec) GetMaxDecodingSize(ctx context.Context, n int, _ string) (int, error) {
	// not used in the example
	return math.MaxInt32, nil
}

func (ec ExampleCodec) CreateType(_ string, _ bool) (any, error) {
	// parameters here are unused in the example, but can be used to determine what type to expect.
	// this allows remote execution to know how to decode the incoming message
	// and for [codec.NewModifierCodec] to know what type to expect for intermediate phases.
	return ec.offChainType, nil
}

type testStructOff struct {
	Ask int
	Bid int
}

type nestedTestStructOff struct {
	Report    testStructOff
	FeedID    [32]byte
	Timestamp int64
}

type deepNestedTestStructOff struct {
	Reports []nestedTestStructOff
}

type testStructOn struct {
	Ask []byte
	Bid int
}

type nestedTestStructOn struct {
	Report    []byte
	FeedID    [32]byte
	Timestamp int64
}

type deepNestedTestStructOn struct {
	Reports []nestedTestStructOn
}

const (
	TestStructOffABI = "uint256 Ask, uint256 Bid"
)

func TestPreCodec(t *testing.T) {
	t.Parallel()

	preCodec := codec.NewPreCodec(
		map[string]string{"Ask": "uint256"},
		func(typeABI string) types.RemoteCodec {
			return ExampleCodec{offChainType: int(0)}
		},
	)
	nestedPreCodec := codec.NewPreCodec(
		map[string]string{"Report": TestStructOffABI},
		func(typeABI string) types.RemoteCodec { return ExampleCodec{offChainType: testStructOff{}} },
	)
	deepNestedPreCodec := codec.NewPreCodec(
		map[string]string{"Reports.Report": TestStructOffABI},
		func(typeABI string) types.RemoteCodec { return ExampleCodec{offChainType: testStructOff{}} },
	)
	invalidPreCodec := codec.NewPreCodec(
		map[string]string{"Unknown": TestStructOffABI},
		func(typeABI string) types.RemoteCodec { return ExampleCodec{offChainType: testStructOff{}} },
	)

	t.Run("RetypeToOffChain converts type to codec.CreateType type", func(t *testing.T) {
		offChainType, err := preCodec.RetypeToOffChain(reflect.TypeOf(testStructOn{}), "")

		require.NoError(t, err)

		require.Equal(t, 2, offChainType.NumField())
		field0 := offChainType.Field(0)
		assert.Equal(t, "Ask", field0.Name)
		assert.Equal(t, reflect.TypeOf(int(0)), field0.Type)
		field1 := offChainType.Field(1)
		assert.Equal(t, "Bid", field1.Name)
		assert.Equal(t, reflect.TypeOf(int(0)), field1.Type)
	})

	t.Run("RetypeToOffChain converts nested type to codec.CreateType type", func(t *testing.T) {
		offChainType, err := nestedPreCodec.RetypeToOffChain(reflect.TypeOf(nestedTestStructOn{}), "")

		require.NoError(t, err)

		require.Equal(t, 3, offChainType.NumField())
		field0 := offChainType.Field(0)
		assert.Equal(t, "Report", field0.Name)
		assert.Equal(t, reflect.TypeOf(testStructOff{}), field0.Type)
		field1 := offChainType.Field(1)
		assert.Equal(t, "FeedID", field1.Name)
		assert.Equal(t, reflect.TypeOf([32]byte{}), field1.Type)
		field2 := offChainType.Field(2)
		assert.Equal(t, "Timestamp", field2.Name)
		assert.Equal(t, reflect.TypeOf(int64(0)), field2.Type)
	})

	t.Run("RetypeToOffChain converts deep nested type to codec.CreateType type", func(t *testing.T) {
		offChainType, err := deepNestedPreCodec.RetypeToOffChain(reflect.TypeOf(deepNestedTestStructOn{}), "")

		require.NoError(t, err)

		reports, exists := offChainType.FieldByName("Reports")
		assert.True(t, exists)
		report := reports.Type.Elem()
		require.Equal(t, 3, report.NumField())
		field0 := report.Field(0)
		assert.Equal(t, "Report", field0.Name)
		assert.Equal(t, reflect.TypeOf(testStructOff{}), field0.Type)
		field1 := report.Field(1)
		assert.Equal(t, "FeedID", field1.Name)
		assert.Equal(t, reflect.TypeOf([32]byte{}), field1.Type)
		field2 := report.Field(2)
		assert.Equal(t, "Timestamp", field2.Name)
		assert.Equal(t, reflect.TypeOf(int64(0)), field2.Type)
	})

	t.Run("RetypeToOffChain only works on byte arrays", func(t *testing.T) {
		_, err := preCodec.RetypeToOffChain(reflect.TypeOf(testStructOff{}), "")
		require.Error(t, err)
		assert.Equal(t, err.Error(), "can only decode []byte from on-chain: int")
	})

	t.Run("RetypeToOffChain only works with a valid path", func(t *testing.T) {
		_, err := invalidPreCodec.RetypeToOffChain(reflect.TypeOf(testStructOn{}), "")
		require.Error(t, err)
		assert.Equal(t, err.Error(), "invalid type: cannot find Unknown")
	})

	t.Run("TransformToOnChain and TransformToOffChain returns error if input type was not from TransformToOnChain", func(t *testing.T) {
		incorrectVal := struct{}{}
		_, err := preCodec.TransformToOnChain(incorrectVal, "")
		assert.True(t, errors.Is(err, types.ErrInvalidType))
		_, err = preCodec.TransformToOffChain(incorrectVal, "")
		assert.True(t, errors.Is(err, types.ErrInvalidType))
	})

	t.Run("TransformToOnChain and TransformToOffChain works on structs", func(t *testing.T) {
		offChainType, err := preCodec.RetypeToOffChain(reflect.TypeOf(testStructOn{}), "")
		require.NoError(t, err)
		iOffchain := reflect.Indirect(reflect.New(offChainType))
		iOffchain.FieldByName("Ask").SetInt(20)
		iOffchain.FieldByName("Bid").SetInt(10)

		output, err := preCodec.TransformToOnChain(iOffchain.Interface(), "")
		require.NoError(t, err)

		jsonEncoded, err := json.Marshal(20)
		require.NoError(t, err)
		expected := testStructOn{
			Ask: jsonEncoded,
			Bid: 10,
		}
		assert.Equal(t, expected, output)
		newInput, err := preCodec.TransformToOffChain(expected, "")
		require.NoError(t, err)
		assert.Equal(t, iOffchain.Interface(), newInput)
	})
}
