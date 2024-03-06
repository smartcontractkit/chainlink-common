package ccip

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHash_String(t *testing.T) {
	tests := []struct {
		name string
		h    Hash
		want string
	}{
		{
			name: "empty",
			h:    Hash{},
			want: "0x0000000000000000000000000000000000000000000000000000000000000000",
		},
		{
			name: "1..",
			h:    Hash{1},
			want: "0x0100000000000000000000000000000000000000000000000000000000000000",
		},
		{
			name: "1..000..1",
			h:    [32]byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
			want: "0x0100000000000000000000000000000000000000000000000000000000000001",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.h.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddress_JSONUnmarshal(t *testing.T) {
	addr1 := "0x507877c2e26f1387432d067d2daafa7d0420d90a"
	addr1Eip55 := "0x507877C2E26f1387432D067D2DaAfa7d0420d90a"

	t.Run("arrays lower case unmarshalled to eip55", func(t *testing.T) {
		js := []byte(`["` + addr1 + `"]`)
		exp := []Address{Address(addr1Eip55)}
		var res []Address
		err := json.Unmarshal(js, &res)
		assert.NoError(t, err)
		assert.Equal(t, exp, res)
	})

	t.Run("maps lower case unmarshalled to eip55", func(t *testing.T) {
		js := []byte(`{"` + addr1 + `": 123}`)
		exp := map[Address]int{Address(addr1Eip55): 123}
		var res map[Address]int
		err := json.Unmarshal(js, &res)
		assert.NoError(t, err)
		assert.Equal(t, exp, res)
	})

	t.Run("non evm", func(t *testing.T) {
		js := []byte(`{"abCDefg": ["lalala"]}`)
		exp := map[Address][]Address{Address("abCDefg"): {"lalala"}}
		var res map[Address][]Address
		err := json.Unmarshal(js, &res)
		assert.NoError(t, err)
		assert.Equal(t, exp, res)
	})
}

func TestAddress_JSONMarshal(t *testing.T) {
	addr1 := "0x507877c2e26f1387432d067d2daafa7d0420d90a"
	addr1Eip55 := "0x507877C2E26f1387432D067D2DaAfa7d0420d90a"

	testCases := []struct {
		name    string
		inp     any
		expJson string
		expErr  bool
	}{
		{
			name:    "array values",
			inp:     []Address{Address(addr1), Address(strings.ToLower(addr1))},
			expJson: `["0x507877c2e26f1387432d067d2daafa7d0420d90a","0x507877c2e26f1387432d067d2daafa7d0420d90a"]`,
			expErr:  false,
		},
		{
			name: "map key lower should remain lower",
			inp: map[Address]int{
				Address(addr1): 1234,
			},
			expJson: `{"0x507877c2e26f1387432d067d2daafa7d0420d90a":1234}`,
			expErr:  false,
		},
		{
			name: "map key eip55 will be eip55 in json",
			inp: map[Address]int{
				Address(addr1Eip55): 1234,
			},
			expJson: `{"0x507877C2E26f1387432D067D2DaAfa7d0420d90a":1234}`,
			expErr:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := json.Marshal(tc.inp)
			if tc.expErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.expJson, string(b))
		})
	}
}

func TestAddress_UnmarshalText(t *testing.T) {
	testCases := []struct {
		name   string
		txt    string
		exp    Address
		expErr bool
	}{
		{
			name: "non-evm",
			txt:  "something",
			exp:  Address("something"),
		},
		{
			name: "very large",
			txt:  strings.Repeat("abc", 1000),
			exp:  Address(strings.Repeat("abc", 1000)),
		},
		{
			name: "lower evm leads to eip",
			txt:  "0x507877c2e26f1387432d067d2daafa7d0420d90a",
			exp:  Address("0x507877C2E26f1387432D067D2DaAfa7d0420d90a"),
		},
		{
			name: "eip55 leads to eip55",
			txt:  "0x507877C2E26f1387432D067D2DaAfa7d0420d90a",
			exp:  Address("0x507877C2E26f1387432D067D2DaAfa7d0420d90a"),
		},
		{
			name: "all caps leads to eip55",
			txt:  "0x507877C2E26F1387432D067D2DAAFA7D0420D90A",
			exp:  Address("0x507877C2E26f1387432D067D2DaAfa7d0420d90a"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var a Address
			err := a.UnmarshalText([]byte(tc.txt))
			if tc.expErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.exp, a)
		})
	}
}

func TestAddress_MarshalText(t *testing.T) {
	testCases := []struct {
		name   string
		addr   Address
		exp    string
		expErr bool
	}{
		{
			name: "non-evm",
			addr: Address("something"),
			exp:  "something",
		},
		{
			name: "eip55 leads to lower",
			addr: Address("0x507877C2E26f1387432D067D2DaAfa7d0420d90a"),
			exp:  "0x507877c2e26f1387432d067d2daafa7d0420d90a",
		},
		{
			name: "lower to lower",
			addr: Address("0x507877c2e26f1387432d067d2daafa7d0420d90a"),
			exp:  "0x507877c2e26f1387432d067d2daafa7d0420d90a",
		},
		{
			name: "all caps to lower",
			addr: Address("0x507877C2E26F1387432D067D2DAAFA7D0420D90A"),
			exp:  "0x507877c2e26f1387432d067d2daafa7d0420d90a",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := tc.addr.MarshalText()
			if tc.expErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.exp, string(res))
		})

	}

}

func TestEIP55(t *testing.T) {
	testCases := []struct {
		inp    string
		exp    string
		expErr bool
	}{
		{
			inp:    "0x507877c2e26f1387432d067d2daafa7d0420d90a",
			exp:    "0x507877C2E26f1387432D067D2DaAfa7d0420d90a",
			expErr: false,
		},
		{
			inp:    "0x001d3f1ef827552ae1114027bd3ecf1f086ba0f9",
			exp:    "0x001d3F1ef827552Ae1114027BD3ECF1f086bA0F9",
			expErr: false,
		},
		{
			inp:    "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			exp:    "0xaAaAaAaaAaAaAaaAaAAAAAAAAaaaAaAaAaaAaaAa",
			expErr: false,
		},
		{
			inp:    "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			exp:    "0xaAaAaAaaAaAaAaaAaAAAAAAAAaaaAaAaAaaAaaAa",
			expErr: false,
		},
		{
			inp:    "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", // 39 chars
			expErr: true,
		},
		{
			inp:    "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", // 41 chars
			expErr: true,
		},
		{
			inp:    "aaaaaaaaaaaaagaaaaaaaaaaaaaaaaaaaaaaaaaa", // contains g
			expErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.inp, func(t *testing.T) {
			res, err := EIP55(tc.inp)

			if tc.expErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.exp, res)
		})
	}
}
