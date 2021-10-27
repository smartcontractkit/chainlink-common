package webhook

import (
	"testing"

	"github.com/smartcontractkit/chainlink-relay/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthenticate(t *testing.T) {
	t.Parallel()
	auth := Authenticate("test-key", "test-secret")

	keys := []struct {
		name   string
		key    string
		secret string
		code   int
	}{
		{"correct", "test-key", "test-secret", 200},
		{"incorrect-secret", "test-key", "wrong-secet", 401},
		{"incorrect-key", "wrong-key", "test-secet", 401},
		{"incorrect", "wrong-key", "wrong-secet", 401},
	}

	for _, k := range keys {
		t.Run(k.name, func(t *testing.T) {
			// create response recorder and gin context
			res, ctx, err := test.MockGinContext([]byte{})
			require.NoError(t, err)
			ctx.Request.Header.Add(webhookAccessKeyHeader, k.key)
			ctx.Request.Header.Add(webhookSecretHeader, k.secret)

			auth(ctx)
			assert.Equal(t, k.code, res.Code)
		})
	}
}
