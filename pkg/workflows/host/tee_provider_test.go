package host

import (
	"testing"

	"github.com/stretchr/testify/assert"

	sdkpb "github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
)

func TestNewTeeProvider(t *testing.T) {
	t.Parallel()
	t.Run("matches any", func(t *testing.T) {
		p := teeProvider{TeeType: sdkpb.TeeType_TEE_TYPE_AWS_NITRO, regions: map[string]bool{"us-west-2": true}}
		tee := &sdkpb.Tee{Item: &sdkpb.Tee_AnyRegions{AnyRegions: &sdkpb.Regions{Regions: []string{"us-west-2"}}}}
		assert.True(t, p.Provides(tee))
	})

	t.Run("matches type selection with matching region", func(t *testing.T) {
		p := teeProvider{TeeType: sdkpb.TeeType_TEE_TYPE_AWS_NITRO, regions: map[string]bool{"us-west-2": true}}
		tee := &sdkpb.Tee{Item: &sdkpb.Tee_TeeTypesAndRegions{
			TeeTypesAndRegions: &sdkpb.TeeTypesAndRegions{
				TeeTypeAndRegions: []*sdkpb.TeeTypeAndRegions{
					{Type: sdkpb.TeeType(99)},
					{Type: sdkpb.TeeType_TEE_TYPE_AWS_NITRO, Regions: []string{"us-west-2"}},
				},
			},
		}}
		assert.True(t, p.Provides(tee))
	})

	t.Run("does not match different types", func(t *testing.T) {
		p := teeProvider{TeeType: sdkpb.TeeType(99)}
		tee := &sdkpb.Tee{Item: &sdkpb.Tee_TeeTypesAndRegions{
			TeeTypesAndRegions: &sdkpb.TeeTypesAndRegions{
				TeeTypeAndRegions: []*sdkpb.TeeTypeAndRegions{
					{Type: sdkpb.TeeType_TEE_TYPE_AWS_NITRO, Regions: []string{"us-west-2"}},
				},
			},
		}}
		assert.False(t, p.Provides(tee))
	})

	t.Run("matches type and region", func(t *testing.T) {
		p := teeProvider{TeeType: sdkpb.TeeType_TEE_TYPE_AWS_NITRO, regions: map[string]bool{"us-west-2": true}}
		tee := &sdkpb.Tee{Item: &sdkpb.Tee_TeeTypesAndRegions{
			TeeTypesAndRegions: &sdkpb.TeeTypesAndRegions{
				TeeTypeAndRegions: []*sdkpb.TeeTypeAndRegions{
					{Type: sdkpb.TeeType_TEE_TYPE_AWS_NITRO, Regions: []string{"us-west-2"}},
				},
			},
		}}
		assert.True(t, p.Provides(tee))
	})

	t.Run("matches type but not region", func(t *testing.T) {
		p := teeProvider{TeeType: sdkpb.TeeType_TEE_TYPE_AWS_NITRO, regions: map[string]bool{"us-west-2": true}}
		tee := &sdkpb.Tee{Item: &sdkpb.Tee_TeeTypesAndRegions{
			TeeTypesAndRegions: &sdkpb.TeeTypesAndRegions{
				TeeTypeAndRegions: []*sdkpb.TeeTypeAndRegions{
					{Type: sdkpb.TeeType_TEE_TYPE_AWS_NITRO, Regions: []string{"eu-west-1"}},
				},
			},
		}}
		assert.False(t, p.Provides(tee))
	})

	t.Run("matches one of multiple requested regions", func(t *testing.T) {
		p := teeProvider{TeeType: sdkpb.TeeType_TEE_TYPE_AWS_NITRO, regions: map[string]bool{"eu-west-1": true}}
		tee := &sdkpb.Tee{Item: &sdkpb.Tee_TeeTypesAndRegions{
			TeeTypesAndRegions: &sdkpb.TeeTypesAndRegions{
				TeeTypeAndRegions: []*sdkpb.TeeTypeAndRegions{
					{Type: sdkpb.TeeType_TEE_TYPE_AWS_NITRO, Regions: []string{"us-west-2", "eu-west-1"}},
				},
			},
		}}
		assert.True(t, p.Provides(tee))
	})

	t.Run("provider has multiple regions and one matches", func(t *testing.T) {
		p := teeProvider{
			TeeType: sdkpb.TeeType_TEE_TYPE_AWS_NITRO,
			regions: map[string]bool{"us-west-2": true, "us-east-1": true},
		}
		tee := &sdkpb.Tee{Item: &sdkpb.Tee_TeeTypesAndRegions{
			TeeTypesAndRegions: &sdkpb.TeeTypesAndRegions{
				TeeTypeAndRegions: []*sdkpb.TeeTypeAndRegions{
					{Type: sdkpb.TeeType_TEE_TYPE_AWS_NITRO, Regions: []string{"us-east-1"}},
				},
			},
		}}
		assert.True(t, p.Provides(tee))
	})

	t.Run("no matching region across multiple provider regions", func(t *testing.T) {
		p := teeProvider{
			TeeType: sdkpb.TeeType_TEE_TYPE_AWS_NITRO,
			regions: map[string]bool{"us-west-2": true, "us-east-1": true},
		}
		tee := &sdkpb.Tee{Item: &sdkpb.Tee_TeeTypesAndRegions{
			TeeTypesAndRegions: &sdkpb.TeeTypesAndRegions{
				TeeTypeAndRegions: []*sdkpb.TeeTypeAndRegions{
					{Type: sdkpb.TeeType_TEE_TYPE_AWS_NITRO, Regions: []string{"ap-southeast-1"}},
				},
			},
		}}
		assert.False(t, p.Provides(tee))
	})

	t.Run("type mismatch ignores region match", func(t *testing.T) {
		p := teeProvider{TeeType: sdkpb.TeeType(99), regions: map[string]bool{"us-west-2": true}}
		tee := &sdkpb.Tee{Item: &sdkpb.Tee_TeeTypesAndRegions{
			TeeTypesAndRegions: &sdkpb.TeeTypesAndRegions{
				TeeTypeAndRegions: []*sdkpb.TeeTypeAndRegions{
					{Type: sdkpb.TeeType_TEE_TYPE_AWS_NITRO, Regions: []string{"us-west-2"}},
				},
			},
		}}
		assert.False(t, p.Provides(tee))
	})

	t.Run("matches any tee", func(t *testing.T) {
		provides := NewTeeProvider(sdkpb.TeeType_TEE_TYPE_AWS_NITRO, []string{"us-west-2"})
		tee := &sdkpb.Tee{Item: &sdkpb.Tee_AnyRegions{AnyRegions: &sdkpb.Regions{Regions: []string{"us-west-2"}}}}
		assert.True(t, provides(tee))
	})

	t.Run("returns a function that checks regions", func(t *testing.T) {
		provides := NewTeeProvider(sdkpb.TeeType_TEE_TYPE_AWS_NITRO, []string{"us-west-2"})
		tee := &sdkpb.Tee{Item: &sdkpb.Tee_TeeTypesAndRegions{
			TeeTypesAndRegions: &sdkpb.TeeTypesAndRegions{
				TeeTypeAndRegions: []*sdkpb.TeeTypeAndRegions{
					{Type: sdkpb.TeeType_TEE_TYPE_AWS_NITRO, Regions: []string{"eu-west-1"}},
				},
			},
		}}
		assert.False(t, provides(tee))
	})

	t.Run("returns false when tee item is nil", func(t *testing.T) {
		provides := NewTeeProvider(sdkpb.TeeType_TEE_TYPE_AWS_NITRO, []string{"us-west-2"})
		tee := &sdkpb.Tee{}
		assert.True(t, provides(tee))
	})

	t.Run("AnyRegions with empty region list returns false", func(t *testing.T) {
		provides := NewTeeProvider(sdkpb.TeeType_TEE_TYPE_AWS_NITRO, []string{"us-west-2"})
		tee := &sdkpb.Tee{Item: &sdkpb.Tee_AnyRegions{AnyRegions: &sdkpb.Regions{}}}
		assert.True(t, provides(tee))
	})

	t.Run("TeeTypesAndRegions with empty region list returns true", func(t *testing.T) {
		provides := NewTeeProvider(sdkpb.TeeType_TEE_TYPE_AWS_NITRO, []string{"us-west-2"})
		tee := &sdkpb.Tee{Item: &sdkpb.Tee_TeeTypesAndRegions{TeeTypesAndRegions: &sdkpb.TeeTypesAndRegions{
			TeeTypeAndRegions: []*sdkpb.TeeTypeAndRegions{
				{Type: sdkpb.TeeType_TEE_TYPE_AWS_NITRO},
			},
		}}}
		assert.True(t, provides(tee))
	})

	t.Run("TeeTypesAndRegions with nil regions returns false", func(t *testing.T) {
		provides := NewTeeProvider(sdkpb.TeeType_TEE_TYPE_AWS_NITRO, []string{"us-west-2"})
		tee := &sdkpb.Tee{Item: &sdkpb.Tee_TeeTypesAndRegions{TeeTypesAndRegions: &sdkpb.TeeTypesAndRegions{
			TeeTypeAndRegions: []*sdkpb.TeeTypeAndRegions{
				{Type: sdkpb.TeeType_TEE_TYPE_AWS_NITRO, Regions: nil},
			},
		}}}
		assert.True(t, provides(tee))
	})
}
