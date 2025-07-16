package report

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

func TestGenerateReportContext(t *testing.T) {
	seqNr := uint64(12345)
	configDigest := types.ConfigDigest{}
	copy(configDigest[:], []byte("testconfigdigest"))
	result := generateReportContext(seqNr, configDigest)
	expectedResult, err := hex.DecodeString("74657374636f6e6669676469676573740000000000000000000000000000000000000000000000000000000000000000000000000000000000000000003039000000000000000000000000000000000000000000000000000000000000000000")
	require.NoError(t, err)
	assert.Equal(t, expectedResult, result)
}
