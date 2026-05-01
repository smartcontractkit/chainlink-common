package host

import (
	"testing"

	"github.com/stretchr/testify/assert"

	sdkpb "github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
)

func TestNewProviderFromSelection(t *testing.T) {
	t.Parallel()

	t.Run("returns false for nil selection", func(t *testing.T) {
		provider := NewProviderFromSelection(nil)
		assert.False(t, provider(&sdkpb.Tee{Item: &sdkpb.Tee_AnyRegions{AnyRegions: &sdkpb.Regions{Regions: []string{"us-west-2"}}}}))
	})

	t.Run("single type selection delegates to tee provider", func(t *testing.T) {
		provider := NewProviderFromSelection([]*sdkpb.TeeTypeAndRegions{{
			Type:    sdkpb.TeeType_TEE_TYPE_AWS_NITRO,
			Regions: []string{"us-west-2"},
		}})

		tee := &sdkpb.Tee{Item: &sdkpb.Tee_TeeTypesAndRegions{TeeTypesAndRegions: &sdkpb.TeeTypesAndRegions{TeeTypeAndRegions: []*sdkpb.TeeTypeAndRegions{{
			Type:    sdkpb.TeeType_TEE_TYPE_AWS_NITRO,
			Regions: []string{"us-west-2"},
		}}}}}
		assert.True(t, provider(tee))
	})

	t.Run("multiple types support any tee", func(t *testing.T) {
		provider := NewProviderFromSelection([]*sdkpb.TeeTypeAndRegions{
			{Type: sdkpb.TeeType_TEE_TYPE_AWS_NITRO, Regions: []string{"us-west-2"}},
			{Type: sdkpb.TeeType(999), Regions: []string{"eu-west-1"}},
		})

		tee := &sdkpb.Tee{Item: &sdkpb.Tee_AnyRegions{AnyRegions: &sdkpb.Regions{Regions: []string{"eu-west-1"}}}}
		assert.True(t, provider(tee))
	})

	t.Run("multiple types merges regions for same type", func(t *testing.T) {
		provider := NewProviderFromSelection([]*sdkpb.TeeTypeAndRegions{
			{Type: sdkpb.TeeType_TEE_TYPE_AWS_NITRO, Regions: []string{"us-west-2"}},
			{Type: sdkpb.TeeType_TEE_TYPE_AWS_NITRO, Regions: []string{"eu-west-1"}},
		})

		regions := []string{"eu-west-1"}
		tee := &sdkpb.Tee{Item: &sdkpb.Tee_TeeTypesAndRegions{TeeTypesAndRegions: &sdkpb.TeeTypesAndRegions{TeeTypeAndRegions: []*sdkpb.TeeTypeAndRegions{{
			Type:    sdkpb.TeeType_TEE_TYPE_AWS_NITRO,
			Regions: regions,
		}}}}}
		assert.True(t, provider(tee))
		regions[0] = "us-west-2"
		assert.True(t, provider(tee))
	})

	t.Run("multiple types returns false when requested type is not supplied", func(t *testing.T) {
		provider := NewProviderFromSelection([]*sdkpb.TeeTypeAndRegions{{
			Type:    sdkpb.TeeType_TEE_TYPE_AWS_NITRO,
			Regions: []string{"us-west-2"},
		}})

		tee := &sdkpb.Tee{Item: &sdkpb.Tee_TeeTypesAndRegions{TeeTypesAndRegions: &sdkpb.TeeTypesAndRegions{TeeTypeAndRegions: []*sdkpb.TeeTypeAndRegions{{
			Type:    sdkpb.TeeType(999),
			Regions: []string{"us-west-2"},
		}}}}}
		assert.False(t, provider(tee))
	})

	t.Run("returns false for unsupported tee shape", func(t *testing.T) {
		provider := NewProviderFromSelection([]*sdkpb.TeeTypeAndRegions{{
			Type: sdkpb.TeeType_TEE_TYPE_AWS_NITRO,
		}})
		assert.False(t, provider(&sdkpb.Tee{}))
	})

	t.Run("multi-type AnyRegions returns false when no provider matches", func(t *testing.T) {
		provider := NewProviderFromSelection([]*sdkpb.TeeTypeAndRegions{
			{Type: sdkpb.TeeType_TEE_TYPE_AWS_NITRO, Regions: []string{"us-west-2"}},
			{Type: sdkpb.TeeType(999), Regions: []string{"eu-west-1"}},
		})

		// AnyRegions with a region that doesn't match any provider's regions
		tee := &sdkpb.Tee{Item: &sdkpb.Tee_AnyRegions{AnyRegions: &sdkpb.Regions{Regions: []string{"ap-southeast-1"}}}}
		assert.False(t, provider(tee))
	})

	t.Run("multi-type TeeTypesAndRegions with nil TeeTypesAndRegions returns false", func(t *testing.T) {
		provider := NewProviderFromSelection([]*sdkpb.TeeTypeAndRegions{
			{Type: sdkpb.TeeType_TEE_TYPE_AWS_NITRO, Regions: []string{"us-west-2"}},
		})

		tee := &sdkpb.Tee{Item: &sdkpb.Tee_TeeTypesAndRegions{TeeTypesAndRegions: nil}}
		assert.False(t, provider(tee))
	})

	t.Run("multi-type TeeTypesAndRegions returns false when all requested types not in providers", func(t *testing.T) {
		provider := NewProviderFromSelection([]*sdkpb.TeeTypeAndRegions{
			{Type: sdkpb.TeeType_TEE_TYPE_AWS_NITRO, Regions: []string{"us-west-2"}},
		})

		// Request types that don't exist in providers
		tee := &sdkpb.Tee{Item: &sdkpb.Tee_TeeTypesAndRegions{TeeTypesAndRegions: &sdkpb.TeeTypesAndRegions{
			TeeTypeAndRegions: []*sdkpb.TeeTypeAndRegions{
				{Type: sdkpb.TeeType(999), Regions: []string{"us-west-2"}},
				{Type: sdkpb.TeeType(888), Regions: []string{"eu-west-1"}},
			},
		}}}
		assert.False(t, provider(tee))
	})

	t.Run("single type TeeTypesAndRegions with non-matching region returns false", func(t *testing.T) {
		provider := NewProviderFromSelection([]*sdkpb.TeeTypeAndRegions{{
			Type:    sdkpb.TeeType_TEE_TYPE_AWS_NITRO,
			Regions: []string{"us-west-2"},
		}})

		tee := &sdkpb.Tee{Item: &sdkpb.Tee_TeeTypesAndRegions{TeeTypesAndRegions: &sdkpb.TeeTypesAndRegions{TeeTypeAndRegions: []*sdkpb.TeeTypeAndRegions{{
			Type:    sdkpb.TeeType_TEE_TYPE_AWS_NITRO,
			Regions: []string{"eu-west-1"},
		}}}}}
		assert.False(t, provider(tee))
	})

	t.Run("multi-type TeeTypesAndRegions partial match skips non-providers", func(t *testing.T) {
		provider := NewProviderFromSelection([]*sdkpb.TeeTypeAndRegions{
			{Type: sdkpb.TeeType_TEE_TYPE_AWS_NITRO, Regions: []string{"us-west-2"}},
		})

		// Request both AWS_NITRO and an unknown type; AWS_NITRO should match
		tee := &sdkpb.Tee{Item: &sdkpb.Tee_TeeTypesAndRegions{TeeTypesAndRegions: &sdkpb.TeeTypesAndRegions{
			TeeTypeAndRegions: []*sdkpb.TeeTypeAndRegions{
				{Type: sdkpb.TeeType(999), Regions: []string{"eu-west-1"}},
				{Type: sdkpb.TeeType_TEE_TYPE_AWS_NITRO, Regions: []string{"us-west-2"}},
			},
		}}}
		assert.True(t, provider(tee))
	})

	t.Run("single type returns directly without closure for AnyRegions", func(t *testing.T) {
		provider := NewProviderFromSelection([]*sdkpb.TeeTypeAndRegions{{
			Type:    sdkpb.TeeType_TEE_TYPE_AWS_NITRO,
			Regions: []string{"us-west-2"},
		}})

		tee := &sdkpb.Tee{Item: &sdkpb.Tee_AnyRegions{AnyRegions: &sdkpb.Regions{Regions: []string{"us-west-2"}}}}
		assert.True(t, provider(tee))
	})

	t.Run("single type returns false for non-matching AnyRegions", func(t *testing.T) {
		provider := NewProviderFromSelection([]*sdkpb.TeeTypeAndRegions{{
			Type:    sdkpb.TeeType_TEE_TYPE_AWS_NITRO,
			Regions: []string{"us-west-2"},
		}})

		tee := &sdkpb.Tee{Item: &sdkpb.Tee_AnyRegions{AnyRegions: &sdkpb.Regions{Regions: []string{"eu-west-1"}}}}
		assert.False(t, provider(tee))
	})

	t.Run("multi-type TeeTypesAndRegions with empty TeeTypeAndRegions array returns false", func(t *testing.T) {
		provider := NewProviderFromSelection([]*sdkpb.TeeTypeAndRegions{
			{Type: sdkpb.TeeType_TEE_TYPE_AWS_NITRO, Regions: []string{"us-west-2"}},
		})

		tee := &sdkpb.Tee{Item: &sdkpb.Tee_TeeTypesAndRegions{TeeTypesAndRegions: &sdkpb.TeeTypesAndRegions{
			TeeTypeAndRegions: []*sdkpb.TeeTypeAndRegions{},
		}}}
		assert.False(t, provider(tee))
	})

	t.Run("multi-type AnyRegions with empty regions list returns false", func(t *testing.T) {
		provider := NewProviderFromSelection([]*sdkpb.TeeTypeAndRegions{
			{Type: sdkpb.TeeType_TEE_TYPE_AWS_NITRO, Regions: []string{"us-west-2"}},
			{Type: sdkpb.TeeType(999), Regions: []string{"eu-west-1"}},
		})

		tee := &sdkpb.Tee{Item: &sdkpb.Tee_AnyRegions{AnyRegions: &sdkpb.Regions{}}}
		assert.False(t, provider(tee))
	})

	t.Run("multiple types with no regions in first type", func(t *testing.T) {
		provider := NewProviderFromSelection([]*sdkpb.TeeTypeAndRegions{
			{Type: sdkpb.TeeType_TEE_TYPE_AWS_NITRO},
			{Type: sdkpb.TeeType(999), Regions: []string{"eu-west-1"}},
		})

		tee := &sdkpb.Tee{Item: &sdkpb.Tee_AnyRegions{AnyRegions: &sdkpb.Regions{Regions: []string{"eu-west-1"}}}}
		assert.True(t, provider(tee))
	})

	t.Run("multi-type TeeTypesAndRegions with first type not matching then match on second", func(t *testing.T) {
		provider := NewProviderFromSelection([]*sdkpb.TeeTypeAndRegions{
			{Type: sdkpb.TeeType_TEE_TYPE_AWS_NITRO, Regions: []string{"us-west-2"}},
			{Type: sdkpb.TeeType(999), Regions: []string{"eu-west-1"}},
		})

		tee := &sdkpb.Tee{Item: &sdkpb.Tee_TeeTypesAndRegions{TeeTypesAndRegions: &sdkpb.TeeTypesAndRegions{
			TeeTypeAndRegions: []*sdkpb.TeeTypeAndRegions{
				{Type: sdkpb.TeeType(888), Regions: []string{"us-west-2"}},
				{Type: sdkpb.TeeType(999), Regions: []string{"eu-west-1"}},
			},
		}}}
		assert.True(t, provider(tee))
	})

	t.Run("unsupported item type in multi-type scenario", func(t *testing.T) {
		provider := NewProviderFromSelection([]*sdkpb.TeeTypeAndRegions{
			{Type: sdkpb.TeeType_TEE_TYPE_AWS_NITRO, Regions: []string{"us-west-2"}},
			{Type: sdkpb.TeeType(999), Regions: []string{"eu-west-1"}},
		})

		tee := &sdkpb.Tee{}
		assert.False(t, provider(tee))
	})

	t.Run("multi-type TeeTypesAndRegions all types not in providers with continue path", func(t *testing.T) {
		provider := NewProviderFromSelection([]*sdkpb.TeeTypeAndRegions{
			{Type: sdkpb.TeeType_TEE_TYPE_AWS_NITRO, Regions: []string{"us-west-2"}},
			{Type: sdkpb.TeeType(555), Regions: []string{"us-west-2"}},
		})

		// Request a type that is never in providers - forces continue on every iteration
		tee := &sdkpb.Tee{Item: &sdkpb.Tee_TeeTypesAndRegions{TeeTypesAndRegions: &sdkpb.TeeTypesAndRegions{
			TeeTypeAndRegions: []*sdkpb.TeeTypeAndRegions{
				{Type: sdkpb.TeeType(777), Regions: []string{"us-west-2"}},
			},
		}}}
		assert.False(t, provider(tee))
	})
}
