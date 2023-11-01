package internal

import (
	"errors"
	"testing"

	"github.com/smartcontractkit/chainlink-relay/pkg/loop/internal/pb"
)

func TestVersionedBytesFunctionsBadPaths(t *testing.T) {
	t.Run("EncodeVersionedBytes unsupported type", func(t *testing.T) {
		expected := errors.New("json: unsupported type: chan int")
		invalidData := make(chan int)

		_, err := encodeVersionedBytes(invalidData, SimpleJsonEncodingVersion)
		if err == nil || err.Error() != expected.Error() {
			t.Errorf("expected error: %s, but got: %v", expected, err)
		}
	})

	t.Run("EncodeVersionedBytes unsupported encoding version", func(t *testing.T) {
		expected := errors.New("unsupported encoding version 2 for data map[key:value]")
		data := map[string]interface{}{
			"key": "value",
		}

		_, err := encodeVersionedBytes(data, 2)
		if err == nil || err.Error() != expected.Error() {
			t.Errorf("expected error: %s, but got: %v", expected, err)
		}
	})

	t.Run("DecodeVersionedBytes", func(t *testing.T) {
		var decodedData map[string]interface{}
		expected := errors.New("unsupported encoding version 2 for versionedData [97 98 99 100 102]")
		versionedBytes := &pb.VersionedBytes{
			Version: 2, // Unsupported version
			Data:    []byte("abcdf"),
		}

		err := decodeVersionedBytes(&decodedData, versionedBytes)
		if err == nil || err.Error() != expected.Error() {
			t.Errorf("expected error: %s, but got: %v", expected, err)
		}
	})
}
