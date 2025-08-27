package evm_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/smartcontractkit/chainlink-common/pkg/chains/evm"
	evmtypes "github.com/smartcontractkit/chainlink-common/pkg/types/chains/evm"
)

func mkBytes(n int, fill byte) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = fill
	}
	return b
}

func TestAddressConverters(t *testing.T) {
	zero := make([]byte, evmtypes.AddressLength)

	t.Run("ConvertAddressFromProto", func(t *testing.T) {
		cases := []struct {
			name            string
			in              []byte
			wantErrParts    []string
			wantEqualsInput bool
		}{
			{"accepts 20-byte address", mkBytes(evmtypes.AddressLength, 0xAB), nil, true},
			{"rejects nil address", nil, []string{"address can't be nil"}, false},
			{"rejects wrong length and shows hex", mkBytes(evmtypes.AddressLength-1, 0x01), []string{"invalid address", "expected", "value=0x"}, false},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				got, err := evm.ConvertAddressFromProto(tc.in)
				if tc.wantErrParts != nil {
					if err == nil {
						t.Fatalf("expected error, got nil")
					}
					for _, p := range tc.wantErrParts {
						if !strings.Contains(err.Error(), p) {
							t.Fatalf("error missing %q, got: %v", p, err)
						}
					}
					return
				}
				if err != nil {
					t.Fatalf("unexpected err: %v", err)
				}
				if tc.wantEqualsInput && !bytes.Equal(got[:], tc.in) {
					t.Fatalf("mismatch: got=%x want=%x", got[:], tc.in)
				}
			})
		}
	})

	t.Run("ConvertOptionalAddressFromProto", func(t *testing.T) {
		cases := []struct {
			name         string
			in           []byte
			wantZero     bool
			wantErrParts []string
		}{
			{"nil input becomes zero address", nil, true, nil},
			{"empty input becomes zero address", []byte{}, true, nil},
			{"rejects wrong length", mkBytes(evmtypes.AddressLength-2, 0xFF), false, []string{"invalid address"}},
			{"accepts 20-byte address", mkBytes(evmtypes.AddressLength, 0x11), false, nil},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				got, err := evm.ConvertOptionalAddressFromProto(tc.in)
				if tc.wantErrParts != nil {
					if err == nil {
						t.Fatalf("expected error, got nil")
					}
					for _, p := range tc.wantErrParts {
						if !strings.Contains(err.Error(), p) {
							t.Fatalf("error missing %q, got: %v", p, err)
						}
					}
					return
				}
				if err != nil {
					t.Fatalf("unexpected err: %v", err)
				}
				if tc.wantZero {
					if !bytes.Equal(got[:], zero) {
						t.Fatalf("want zero address, got %x", got[:])
					}
				} else if !bytes.Equal(got[:], tc.in) {
					t.Fatalf("mismatch: got=%x want=%x", got[:], tc.in)
				}
			})
		}
	})

	t.Run("Roundtrip", func(t *testing.T) {
		addrs := []evmtypes.Address{
			evmtypes.Address(mkBytes(evmtypes.AddressLength, 0x01)),
			evmtypes.Address(mkBytes(evmtypes.AddressLength, 0x02)),
		}
		proto := evm.ConvertAddressesToProto(addrs)
		got, err := evm.ConvertAddressesFromProto(proto)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if len(got) != len(addrs) {
			t.Fatalf("len mismatch: got=%d want=%d", len(got), len(addrs))
		}
		for i := range addrs {
			if !bytes.Equal(got[i][:], addrs[i][:]) {
				t.Fatalf("idx %d mismatch: got=%x want=%x", i, got[i][:], addrs[i][:])
			}
		}
	})

	t.Run("AggregatesErrors", func(t *testing.T) {
		in := [][]byte{
			mkBytes(evmtypes.AddressLength, 0x01),
			mkBytes(evmtypes.AddressLength-1, 0x02),
			nil,
		}
		got, err := evm.ConvertAddressesFromProto(in)
		if got != nil {
			t.Fatalf("want nil result on aggregate error, got: %#v", got)
		}
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		msg := err.Error()
		if !strings.Contains(msg, "index 1") || !strings.Contains(msg, "index 2") {
			t.Fatalf("joined error should include both failing indices, got: %v", msg)
		}
	})

	t.Run("WrapsErrorsWithContext", func(t *testing.T) {
		in := [][]byte{mkBytes(evmtypes.AddressLength-1, 0x01)}
		_, err := evm.ConvertAddressesFromProto(in)
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "failed to convert address at index 0") {
			t.Fatalf("missing context wrapper, got: %v", err)
		}
		if !errors.Is(err, err) {
			t.Fatalf("unexpected errors.Is behavior")
		}
	})
}

func TestHashConverters(t *testing.T) {
	zero := make([]byte, evmtypes.HashLength)

	t.Run("ConvertHashFromProto", func(t *testing.T) {
		cases := []struct {
			name            string
			in              []byte
			wantErrParts    []string
			wantEqualsInput bool
		}{
			{"accepts 32-byte hash", mkBytes(evmtypes.HashLength, 0xCD), nil, true},
			{"rejects nil hash", nil, []string{"hash can't be nil"}, false},
			{"rejects wrong length and shows hex", mkBytes(evmtypes.HashLength+1, 0x02), []string{"invalid hash", "expected", "value=0x"}, false},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				got, err := evm.ConvertHashFromProto(tc.in)
				if tc.wantErrParts != nil {
					if err == nil {
						t.Fatalf("expected error, got nil")
					}
					for _, p := range tc.wantErrParts {
						if !strings.Contains(err.Error(), p) {
							t.Fatalf("error missing %q, got: %v", p, err)
						}
					}
					return
				}
				if err != nil {
					t.Fatalf("unexpected err: %v", err)
				}
				if tc.wantEqualsInput && !bytes.Equal(got[:], tc.in) {
					t.Fatalf("mismatch: got=%x want=%x", got[:], tc.in)
				}
			})
		}
	})

	t.Run("ConvertOptionalHashFromProto", func(t *testing.T) {
		cases := []struct {
			name         string
			in           []byte
			wantZero     bool
			wantErrParts []string
		}{
			{"nil input becomes zero hash", nil, true, nil},
			{"empty input becomes zero hash", []byte{}, true, nil},
			{"rejects wrong length", mkBytes(evmtypes.HashLength-1, 0xEE), false, []string{"invalid hash"}},
			{"accepts 32-byte hash", mkBytes(evmtypes.HashLength, 0xAA), false, nil},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				got, err := evm.ConvertOptionalHashFromProto(tc.in)
				if tc.wantErrParts != nil {
					if err == nil {
						t.Fatalf("expected error, got nil")
					}
					for _, p := range tc.wantErrParts {
						if !strings.Contains(err.Error(), p) {
							t.Fatalf("error missing %q, got: %v", p, err)
						}
					}
					return
				}
				if err != nil {
					t.Fatalf("unexpected err: %v", err)
				}
				if tc.wantZero {
					if !bytes.Equal(got[:], zero) {
						t.Fatalf("want zero hash, got %x", got[:])
					}
				} else if !bytes.Equal(got[:], tc.in) {
					t.Fatalf("mismatch: got=%x want=%x", got[:], tc.in)
				}
			})
		}
	})

	t.Run("Roundtrip", func(t *testing.T) {
		hashes := []evmtypes.Hash{
			evmtypes.Hash(mkBytes(evmtypes.HashLength, 0xAA)),
			evmtypes.Hash(mkBytes(evmtypes.HashLength, 0xBB)),
		}
		proto := evm.ConvertHashesToProto(hashes)
		got, err := evm.ConvertHashesFromProto(proto)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if len(got) != len(hashes) {
			t.Fatalf("len mismatch: got=%d want=%d", len(got), len(hashes))
		}
		for i := range hashes {
			if !bytes.Equal(got[i][:], hashes[i][:]) {
				t.Fatalf("idx %d mismatch: got=%x want=%x", i, got[i][:], hashes[i][:])
			}
		}
	})

	t.Run("AggregatesErrors", func(t *testing.T) {
		in := [][]byte{
			mkBytes(evmtypes.HashLength-1, 0x03),
			nil,
		}
		got, err := evm.ConvertHashesFromProto(in)
		if got != nil {
			t.Fatalf("want nil result on aggregate error, got: %#v", got)
		}
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "index 0") || !strings.Contains(err.Error(), "index 1") {
			t.Fatalf("joined error should include failing indices, got: %v", err)
		}
	})
}
