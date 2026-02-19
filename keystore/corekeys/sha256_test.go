package corekeys_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/keystore/corekeys"
)

func TestSha256Hash_MarshalJSON_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	hash := corekeys.MustSha256HashFromHex("f5bf259689b26f1374efb3c9a9868796953a0f814bb2d39b968d0e61b58620a5")
	json, err := hash.MarshalJSON()
	require.NoError(t, err)
	require.NotEmpty(t, json)

	var newHash corekeys.Sha256Hash
	err = newHash.UnmarshalJSON(json)
	require.NoError(t, err)

	require.Equal(t, hash, newHash)
}

func TestSha256Hash_Sha256HashFromHex(t *testing.T) {
	t.Parallel()

	_, err := corekeys.Sha256HashFromHex("abczzz")
	require.Error(t, err)

	_, err = corekeys.Sha256HashFromHex("f5bf259689b26f1374efb3c9a9868796953a0f814bb2d39b968d0e61b58620a5")
	require.NoError(t, err)

	_, err = corekeys.Sha256HashFromHex("f5bf259689b26f1374e6")
	require.NoError(t, err)
}

func TestSha256Hash_String(t *testing.T) {
	t.Parallel()

	hash := corekeys.MustSha256HashFromHex("f5bf259689b26f1374efb3c9a9868796953a0f814bb2d39b968d0e61b58620a5")
	assert.Equal(t, "f5bf259689b26f1374efb3c9a9868796953a0f814bb2d39b968d0e61b58620a5", hash.String())
}

func TestSha256Hash_Scan_Value(t *testing.T) {
	t.Parallel()

	hash := corekeys.MustSha256HashFromHex("f5bf259689b26f1374efb3c9a9868796953a0f814bb2d39b968d0e61b58620a5")
	val, err := hash.Value()
	require.NoError(t, err)

	var newHash corekeys.Sha256Hash
	err = newHash.Scan(val)
	require.NoError(t, err)

	require.Equal(t, hash, newHash)
}
