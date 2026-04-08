package host

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/emptypb"

	sdkpb "github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
)

func TestTeeProvider(t *testing.T) {
	t.Parallel()
	t.Run("matches any", func(t *testing.T) {
		p := TeeProvider(sdkpb.TeeType_TEE_TYPE_AWS_NITRO)
		tee := &sdkpb.Tee{Type: &sdkpb.Tee_Any{Any: &emptypb.Empty{}}}
		assert.True(t, p.Provides(tee))
	})

	t.Run("matches type selection", func(t *testing.T) {
		p := TeeProvider(sdkpb.TeeType_TEE_TYPE_AWS_NITRO)
		tee := &sdkpb.Tee{Type: &sdkpb.Tee_TypeSelection{
			TypeSelection: &sdkpb.TeeTypeSelection{
				Types: []sdkpb.TeeType{sdkpb.TeeType(99), sdkpb.TeeType_TEE_TYPE_AWS_NITRO},
			},
		}}
		assert.True(t, p.Provides(tee))
	})

	t.Run("does not match any type", func(t *testing.T) {
		// Use a cast to an unknown value so we don't need a second enum variant.
		p := TeeProvider(sdkpb.TeeType(99))
		tee := &sdkpb.Tee{Type: &sdkpb.Tee_TypeSelection{
			TypeSelection: &sdkpb.TeeTypeSelection{
				Types: []sdkpb.TeeType{sdkpb.TeeType_TEE_TYPE_AWS_NITRO},
			},
		}}
		assert.False(t, p.Provides(tee))
	})
}
