package keys

import (
	"os"
	"testing"

	"github.com/smartcontractkit/chainlink-relay/core/server/types"
	"github.com/smartcontractkit/chainlink-relay/core/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShowKeys(t *testing.T) {
	t.Parallel()

	testKeys := map[string]string{}
	testKeys["P2PID"] = "something"

	// create new keys handler
	keys := New(&testKeys)

	// create gin context
	res, ctx, err := test.MockGinContext(nil)
	require.NoError(t, err)

	// test endpoint
	keys.ShowKeys(ctx)
	assert.Equal(t, 200, res.Code)
	resBody, err := test.RequestBody(testKeys) // create response payload
	require.NoError(t, err)
	assert.Equal(t, resBody, res.Body)
}

func TestSetEIKeys_Pass(t *testing.T) {
	t.Parallel()

	// create new handler
	keys := New(&map[string]string{})

	// create gin context
	res, ctx, err := test.MockGinContext(types.SetKeyData{
		"IC_ACCESSKEY", "IC_SECRET", "CI_ACCESSKEY", "CI_SECRET",
	})
	require.NoError(t, err)

	// test endpoint
	keys.SetEIKeys(ctx)
	assert.Equal(t, 201, res.Code)

	vars := []string{"IC_ACCESSKEY", "IC_SECRET", "CI_SECRET", "CI_ACCESSKEY"}
	for _, v := range vars {
		assert.Equal(t, v, os.Getenv(v))
	}
	assert.NoError(t, test.UnsetEIKeysSecrets()) // remove set secrets
}

func TestSetEIKeys_Fail_MissingKey(t *testing.T) {
	t.Parallel()

	// create new handler
	keys := New(&map[string]string{})

	// create gin context
	res, ctx, err := test.MockGinContext(types.SetKeyData{
		ICKey:    "IC_ACCESSKEY",
		ICSecret: "IC_SECRET",
		CIKey:    "CI_ACCESSKEY",
	})
	require.NoError(t, err)

	// test endpoint
	keys.SetEIKeys(ctx)
	assert.Equal(t, 400, res.Code)
}
