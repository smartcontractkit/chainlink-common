package host

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protowire"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
)

func Test_CheckRequirements(t *testing.T) {
	t.Parallel()
	t.Run("unknown proto fields", func(t *testing.T) {
		// Encode a field number (99) unknown to Requirements so proto.Unmarshal
		// preserves it as unknown bytes.
		b := protowire.AppendTag(nil, 99, protowire.VarintType)
		b = protowire.AppendVarint(b, 1)
		req := &sdk.Requirements{}
		require.NoError(t, proto.Unmarshal(b, req))

		assert.False(t, CheckRequirements(context.Background(), RequirementsHandler{}, req))
	})

	t.Run("no fields always passes", func(t *testing.T) {
		assert.True(t, CheckRequirements(context.Background(), RequirementsHandler{}, &sdk.Requirements{}))
	})

	t.Run("handler not set returns false", func(t *testing.T) {
		req := &sdk.Requirements{Tee: &sdk.Tee{}}
		assert.False(t, CheckRequirements(context.Background(), RequirementsHandler{}, req))
	})

	t.Run("handler returns false causes false return value", func(t *testing.T) {
		req := &sdk.Requirements{Tee: &sdk.Tee{}}
		handler := RequirementsHandler{Tee: func(context.Context, *sdk.Tee) bool { return false }}
		assert.False(t, CheckRequirements(context.Background(), handler, req))
	})

	t.Run("handler returns true causes true return value", func(t *testing.T) {
		req := &sdk.Requirements{Tee: &sdk.Tee{}}
		handler := RequirementsHandler{Tee: func(context.Context, *sdk.Tee) bool { return true }}
		assert.True(t, CheckRequirements(context.Background(), handler, req))
	})
}
