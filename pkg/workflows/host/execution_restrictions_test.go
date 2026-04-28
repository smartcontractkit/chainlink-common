package host

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/utils/matches"
	"github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
)

func TestCallCapability(t *testing.T) {
	t.Run("allows when nil restrictions", func(t *testing.T) {
		er := newExecutionRestrictions(nil)
		assert.True(t, er.canCallCapability(&sdk.CapabilityRequest{Id: "cap@1.0.0", Method: "Foo"}))
	})

	t.Run("allows when no capabilities restrictions set", func(t *testing.T) {
		er := newExecutionRestrictions(&sdk.Restrictions{})
		assert.True(t, er.canCallCapability(&sdk.CapabilityRequest{Id: "cap@1.0.0", Method: "Foo"}))
	})

	t.Run("closed denies unmatched capability", func(t *testing.T) {
		er := newExecutionRestrictions(&sdk.Restrictions{
			Capabilities: &sdk.CapabilityRestrictions{
				MaxTotalCalls: 10,
				Type:          sdk.CapabilityRestrictionType_CAPABILITY_RESTRICTION_TYPE_CLOSED,
				Restrictions: []*sdk.CapabilityRestriction{
					{Restriction: &sdk.CapabilityRestriction_Method{
						Method: &sdk.MethodRestriction{Id: "cap@1.0.0", Method: "Foo", MaxCalls: 5},
					}},
				},
			},
		})
		assert.False(t, er.canCallCapability(&sdk.CapabilityRequest{Id: "other-cap@1.0.0", Method: "Bar"}))
	})

	t.Run("closed allows matched capability and decrements", func(t *testing.T) {
		er := newExecutionRestrictions(&sdk.Restrictions{
			Capabilities: &sdk.CapabilityRestrictions{
				MaxTotalCalls: 10,
				Type:          sdk.CapabilityRestrictionType_CAPABILITY_RESTRICTION_TYPE_CLOSED,
				Restrictions: []*sdk.CapabilityRestriction{
					{Restriction: &sdk.CapabilityRestriction_Method{
						Method: &sdk.MethodRestriction{Id: "cap@1.0.0", Method: "Foo", MaxCalls: 2},
					}},
				},
			},
		})
		req := &sdk.CapabilityRequest{Id: "cap@1.0.0", Method: "Foo"}
		key := methodKey{id: "cap@1.0.0", method: "Foo"}

		assert.True(t, er.canCallCapability(req))
		assert.Equal(t, int32(1), er.methods[key])
		assert.Equal(t, int32(9), er.maxTotalCalls)

		assert.True(t, er.canCallCapability(req))
		assert.Equal(t, int32(0), er.methods[key])
		assert.Equal(t, int32(8), er.maxTotalCalls)

		assert.False(t, er.canCallCapability(req))
	})

	t.Run("Denies when max total calls is zero", func(t *testing.T) {
		er := newExecutionRestrictions(&sdk.Restrictions{
			Capabilities: &sdk.CapabilityRestrictions{
				MaxTotalCalls: 0,
				Type:          sdk.CapabilityRestrictionType_CAPABILITY_RESTRICTION_TYPE_CLOSED,
				Restrictions: []*sdk.CapabilityRestriction{
					{Restriction: &sdk.CapabilityRestriction_Method{
						Method: &sdk.MethodRestriction{Id: "cap@1.0.0", Method: "Foo", MaxCalls: 5},
					}},
				},
			},
		})
		assert.False(t, er.canCallCapability(&sdk.CapabilityRequest{Id: "cap@1.0.0", Method: "Foo"}))
	})

	t.Run("open allows unmatched capability", func(t *testing.T) {
		er := newExecutionRestrictions(&sdk.Restrictions{
			Capabilities: &sdk.CapabilityRestrictions{
				MaxTotalCalls: 10,
				Type:          sdk.CapabilityRestrictionType_CAPABILITY_RESTRICTION_TYPE_OPEN,
				Restrictions: []*sdk.CapabilityRestriction{
					{Restriction: &sdk.CapabilityRestriction_Method{
						Method: &sdk.MethodRestriction{Id: "cap@1.0.0", Method: "Foo", MaxCalls: 2},
					}},
				},
			},
		})
		assert.True(t, er.canCallCapability(&sdk.CapabilityRequest{Id: "other-cap@1.0.0", Method: "Bar"}))
		assert.Equal(t, int32(9), er.maxTotalCalls)
	})

	t.Run("Denies when matched method has zero calls remaining", func(t *testing.T) {
		er := newExecutionRestrictions(&sdk.Restrictions{
			Capabilities: &sdk.CapabilityRestrictions{
				MaxTotalCalls: 10,
				Type:          sdk.CapabilityRestrictionType_CAPABILITY_RESTRICTION_TYPE_OPEN,
				Restrictions: []*sdk.CapabilityRestriction{
					{Restriction: &sdk.CapabilityRestriction_Method{
						Method: &sdk.MethodRestriction{Id: "cap@1.0.0", Method: "Foo", MaxCalls: 0},
					}},
				},
			},
		})
		assert.False(t, er.canCallCapability(&sdk.CapabilityRequest{Id: "cap@1.0.0", Method: "Foo"}))
	})

	t.Run("matches by both id and method", func(t *testing.T) {
		er := newExecutionRestrictions(&sdk.Restrictions{
			Capabilities: &sdk.CapabilityRestrictions{
				MaxTotalCalls: 10,
				Type:          sdk.CapabilityRestrictionType_CAPABILITY_RESTRICTION_TYPE_CLOSED,
				Restrictions: []*sdk.CapabilityRestriction{
					{Restriction: &sdk.CapabilityRestriction_Method{
						Method: &sdk.MethodRestriction{Id: "cap@1.0.0", Method: "Foo", MaxCalls: 5},
					}},
				},
			},
		})
		assert.False(t, er.canCallCapability(&sdk.CapabilityRequest{Id: "cap@1.0.0", Method: "Bar"}))
		assert.False(t, er.canCallCapability(&sdk.CapabilityRequest{Id: "cap@2.0.0", Method: "Foo"}))
	})

	t.Run("multiple different methods match independently", func(t *testing.T) {
		er := newExecutionRestrictions(&sdk.Restrictions{
			Capabilities: &sdk.CapabilityRestrictions{
				MaxTotalCalls: 10,
				Type:          sdk.CapabilityRestrictionType_CAPABILITY_RESTRICTION_TYPE_CLOSED,
				Restrictions: []*sdk.CapabilityRestriction{
					{Restriction: &sdk.CapabilityRestriction_Method{
						Method: &sdk.MethodRestriction{Id: "cap@1.0.0", Method: "Foo", MaxCalls: 1},
					}},
					{Restriction: &sdk.CapabilityRestriction_Method{
						Method: &sdk.MethodRestriction{Id: "cap@1.0.0", Method: "Bar", MaxCalls: 1},
					}},
				},
			},
		})

		assert.True(t, er.canCallCapability(&sdk.CapabilityRequest{Id: "cap@1.0.0", Method: "Foo"}))
		assert.False(t, er.canCallCapability(&sdk.CapabilityRequest{Id: "cap@1.0.0", Method: "Foo"}))
		assert.True(t, er.canCallCapability(&sdk.CapabilityRequest{Id: "cap@1.0.0", Method: "Bar"}))
		assert.False(t, er.canCallCapability(&sdk.CapabilityRequest{Id: "cap@1.0.0", Method: "Bar"}))
	})

	t.Run("total calls limit reached before method limit", func(t *testing.T) {
		er := newExecutionRestrictions(&sdk.Restrictions{
			Capabilities: &sdk.CapabilityRestrictions{
				MaxTotalCalls: 2,
				Type:          sdk.CapabilityRestrictionType_CAPABILITY_RESTRICTION_TYPE_CLOSED,
				Restrictions: []*sdk.CapabilityRestriction{
					{Restriction: &sdk.CapabilityRestriction_Method{
						Method: &sdk.MethodRestriction{Id: "cap@1.0.0", Method: "Foo", MaxCalls: 100},
					}},
				},
			},
		})
		req := &sdk.CapabilityRequest{Id: "cap@1.0.0", Method: "Foo"}

		assert.True(t, er.canCallCapability(req))
		assert.True(t, er.canCallCapability(req))
		assert.False(t, er.canCallCapability(req))
	})

	t.Run("negative max total calls means unlimited", func(t *testing.T) {
		er := newExecutionRestrictions(&sdk.Restrictions{
			Capabilities: &sdk.CapabilityRestrictions{
				MaxTotalCalls: -1,
				Type:          sdk.CapabilityRestrictionType_CAPABILITY_RESTRICTION_TYPE_CLOSED,
				Restrictions: []*sdk.CapabilityRestriction{
					{Restriction: &sdk.CapabilityRestriction_Method{
						Method: &sdk.MethodRestriction{Id: "cap@1.0.0", Method: "Foo", MaxCalls: 2},
					}},
				},
			},
		})
		req := &sdk.CapabilityRequest{Id: "cap@1.0.0", Method: "Foo"}

		assert.True(t, er.canCallCapability(req))
		assert.True(t, er.canCallCapability(req))
		assert.False(t, er.canCallCapability(req))
		assert.Equal(t, int32(-1), er.maxTotalCalls)
	})

	t.Run("negative max calls on method means unlimited for that method", func(t *testing.T) {
		er := newExecutionRestrictions(&sdk.Restrictions{
			Capabilities: &sdk.CapabilityRestrictions{
				MaxTotalCalls: 3,
				Type:          sdk.CapabilityRestrictionType_CAPABILITY_RESTRICTION_TYPE_CLOSED,
				Restrictions: []*sdk.CapabilityRestriction{
					{Restriction: &sdk.CapabilityRestriction_Method{
						Method: &sdk.MethodRestriction{Id: "cap@1.0.0", Method: "Foo", MaxCalls: -1},
					}},
				},
			},
		})
		req := &sdk.CapabilityRequest{Id: "cap@1.0.0", Method: "Foo"}

		for i := 0; i < 3; i++ {
			assert.True(t, er.canCallCapability(req))
		}
		assert.False(t, er.canCallCapability(req))
		assert.Equal(t, int32(-1), er.methods[methodKey{id: "cap@1.0.0", method: "Foo"}])
	})

	t.Run("duplicate restrictions keep smallest non-negative value", func(t *testing.T) {
		er := newExecutionRestrictions(&sdk.Restrictions{
			Capabilities: &sdk.CapabilityRestrictions{
				MaxTotalCalls: 10,
				Type:          sdk.CapabilityRestrictionType_CAPABILITY_RESTRICTION_TYPE_CLOSED,
				Restrictions: []*sdk.CapabilityRestriction{
					{Restriction: &sdk.CapabilityRestriction_Method{
						Method: &sdk.MethodRestriction{Id: "cap@1.0.0", Method: "Foo", MaxCalls: 5},
					}},
					{Restriction: &sdk.CapabilityRestriction_Method{
						Method: &sdk.MethodRestriction{Id: "cap@1.0.0", Method: "Foo", MaxCalls: 2},
					}},
				},
			},
		})
		key := methodKey{id: "cap@1.0.0", method: "Foo"}
		assert.Equal(t, int32(2), er.methods[key])
	})

	t.Run("duplicate restrictions non-negative overrides negative", func(t *testing.T) {
		er := newExecutionRestrictions(&sdk.Restrictions{
			Capabilities: &sdk.CapabilityRestrictions{
				MaxTotalCalls: 10,
				Type:          sdk.CapabilityRestrictionType_CAPABILITY_RESTRICTION_TYPE_CLOSED,
				Restrictions: []*sdk.CapabilityRestriction{
					{Restriction: &sdk.CapabilityRestriction_Method{
						Method: &sdk.MethodRestriction{Id: "cap@1.0.0", Method: "Foo", MaxCalls: -1},
					}},
					{Restriction: &sdk.CapabilityRestriction_Method{
						Method: &sdk.MethodRestriction{Id: "cap@1.0.0", Method: "Foo", MaxCalls: 3},
					}},
				},
			},
		})
		key := methodKey{id: "cap@1.0.0", method: "Foo"}
		assert.Equal(t, int32(3), er.methods[key])
	})

	t.Run("duplicate restrictions zero overrides positive", func(t *testing.T) {
		er := newExecutionRestrictions(&sdk.Restrictions{
			Capabilities: &sdk.CapabilityRestrictions{
				MaxTotalCalls: 10,
				Type:          sdk.CapabilityRestrictionType_CAPABILITY_RESTRICTION_TYPE_CLOSED,
				Restrictions: []*sdk.CapabilityRestriction{
					{Restriction: &sdk.CapabilityRestriction_Method{
						Method: &sdk.MethodRestriction{Id: "cap@1.0.0", Method: "Foo", MaxCalls: 5},
					}},
					{Restriction: &sdk.CapabilityRestriction_Method{
						Method: &sdk.MethodRestriction{Id: "cap@1.0.0", Method: "Foo", MaxCalls: 0},
					}},
				},
			},
		})
		assert.False(t, er.canCallCapability(&sdk.CapabilityRequest{Id: "cap@1.0.0", Method: "Foo"}))
	})

	t.Run("closed with no methods denies all", func(t *testing.T) {
		er := newExecutionRestrictions(&sdk.Restrictions{
			Capabilities: &sdk.CapabilityRestrictions{
				MaxTotalCalls: -1,
				Type:          sdk.CapabilityRestrictionType_CAPABILITY_RESTRICTION_TYPE_CLOSED,
			},
		})
		assert.False(t, er.canCallCapability(&sdk.CapabilityRequest{Id: "cap@1.0.0", Method: "Foo"}))
	})

	t.Run("open with no methods respects max total calls", func(t *testing.T) {
		er := newExecutionRestrictions(&sdk.Restrictions{
			Capabilities: &sdk.CapabilityRestrictions{
				MaxTotalCalls: 2,
				Type:          sdk.CapabilityRestrictionType_CAPABILITY_RESTRICTION_TYPE_OPEN,
			},
		})
		assert.True(t, er.canCallCapability(&sdk.CapabilityRequest{Id: "cap@1.0.0", Method: "Foo"}))
		assert.True(t, er.canCallCapability(&sdk.CapabilityRequest{Id: "cap@2.0.0", Method: "Bar"}))
		assert.False(t, er.canCallCapability(&sdk.CapabilityRequest{Id: "cap@3.0.0", Method: "Baz"}))
	})

	t.Run("open with zero max total calls denies all", func(t *testing.T) {
		er := newExecutionRestrictions(&sdk.Restrictions{
			Capabilities: &sdk.CapabilityRestrictions{
				MaxTotalCalls: 0,
				Type:          sdk.CapabilityRestrictionType_CAPABILITY_RESTRICTION_TYPE_OPEN,
			},
		})
		assert.False(t, er.canCallCapability(&sdk.CapabilityRequest{Id: "cap@1.0.0", Method: "Foo"}))
	})
}

func TestGetSecret(t *testing.T) {
	t.Run("allows when nil restrictions", func(t *testing.T) {
		er := newExecutionRestrictions(nil)
		assert.True(t, er.canGetSecret(&sdk.SecretRequest{Id: "my-secret", Namespace: "ns"}))
	})

	t.Run("allows when no secrets restrictions set", func(t *testing.T) {
		er := newExecutionRestrictions(&sdk.Restrictions{})
		assert.True(t, er.canGetSecret(&sdk.SecretRequest{Id: "my-secret", Namespace: "ns"}))
	})

	t.Run("denies when max secrets is zero", func(t *testing.T) {
		er := newExecutionRestrictions(&sdk.Restrictions{
			Secrets: &sdk.SecretsRestritions{
				MaxSecrets: 0,
				Restrictions: []*sdk.SecretRestriction{
					{Restriction: &sdk.SecretRestriction_ExactSecret{
						ExactSecret: &sdk.Secret{Id: "my-secret", Namespace: "ns"},
					}},
				},
			},
		})
		assert.False(t, er.canGetSecret(&sdk.SecretRequest{Id: "my-secret", Namespace: "ns"}))
	})

	t.Run("exact match allows and decrements max secrets", func(t *testing.T) {
		er := newExecutionRestrictions(&sdk.Restrictions{
			Secrets: &sdk.SecretsRestritions{
				MaxSecrets: 2,
				Restrictions: []*sdk.SecretRestriction{
					{Restriction: &sdk.SecretRestriction_ExactSecret{
						ExactSecret: &sdk.Secret{Id: "db-password", Namespace: "infra"},
					}},
				},
			},
		})
		req := &sdk.SecretRequest{Id: "db-password", Namespace: "infra"}

		assert.True(t, er.canGetSecret(req))
		assert.Equal(t, int32(1), er.maxSecrets)

		assert.True(t, er.canGetSecret(req))
		assert.Equal(t, int32(0), er.maxSecrets)

		assert.False(t, er.canGetSecret(req))
	})

	t.Run("exact match requires both id and namespace", func(t *testing.T) {
		er := newExecutionRestrictions(&sdk.Restrictions{
			Secrets: &sdk.SecretsRestritions{
				MaxSecrets: 10,
				Restrictions: []*sdk.SecretRestriction{
					{Restriction: &sdk.SecretRestriction_ExactSecret{
						ExactSecret: &sdk.Secret{Id: "db-password", Namespace: "infra"},
					}},
				},
			},
		})

		assert.False(t, er.canGetSecret(&sdk.SecretRequest{Id: "db-password", Namespace: "other"}))
		assert.False(t, er.canGetSecret(&sdk.SecretRequest{Id: "other", Namespace: "infra"}))
		assert.True(t, er.canGetSecret(&sdk.SecretRequest{Id: "db-password", Namespace: "infra"}))
	})

	t.Run("prefix match allows and decrements both counters", func(t *testing.T) {
		er := newExecutionRestrictions(&sdk.Restrictions{
			Secrets: &sdk.SecretsRestritions{
				MaxSecrets: 10,
				Restrictions: []*sdk.SecretRestriction{
					{Restriction: &sdk.SecretRestriction_PrefixedSecret{
						PrefixedSecret: &sdk.SecretPrefixRestriction{
							Prefix: "db-", Namespace: "infra", MaxSecrets: 2,
						},
					}},
				},
			},
		})

		assert.True(t, er.canGetSecret(&sdk.SecretRequest{Id: "db-password", Namespace: "infra"}))
		assert.Equal(t, int32(1), er.prefixSecrets[0].maxCalls)
		assert.Equal(t, int32(9), er.maxSecrets)

		assert.True(t, er.canGetSecret(&sdk.SecretRequest{Id: "db-host", Namespace: "infra"}))
		assert.Equal(t, int32(0), er.prefixSecrets[0].maxCalls)
		assert.Equal(t, int32(8), er.maxSecrets)

		assert.False(t, er.canGetSecret(&sdk.SecretRequest{Id: "db-port", Namespace: "infra"}))
	})

	t.Run("prefix match requires namespace match", func(t *testing.T) {
		er := newExecutionRestrictions(&sdk.Restrictions{
			Secrets: &sdk.SecretsRestritions{
				MaxSecrets: 10,
				Restrictions: []*sdk.SecretRestriction{
					{Restriction: &sdk.SecretRestriction_PrefixedSecret{
						PrefixedSecret: &sdk.SecretPrefixRestriction{
							Prefix: "db-", Namespace: "infra", MaxSecrets: 5,
						},
					}},
				},
			},
		})

		assert.False(t, er.canGetSecret(&sdk.SecretRequest{Id: "db-password", Namespace: "other"}))
	})

	t.Run("prefix match denied when global max secrets hits zero", func(t *testing.T) {
		er := newExecutionRestrictions(&sdk.Restrictions{
			Secrets: &sdk.SecretsRestritions{
				MaxSecrets: 1,
				Restrictions: []*sdk.SecretRestriction{
					{Restriction: &sdk.SecretRestriction_PrefixedSecret{
						PrefixedSecret: &sdk.SecretPrefixRestriction{
							Prefix: "db-", Namespace: "infra", MaxSecrets: 5,
						},
					}},
				},
			},
		})

		assert.True(t, er.canGetSecret(&sdk.SecretRequest{Id: "db-password", Namespace: "infra"}))
		assert.False(t, er.canGetSecret(&sdk.SecretRequest{Id: "db-host", Namespace: "infra"}))
	})

	t.Run("denies unmatched secret", func(t *testing.T) {
		er := newExecutionRestrictions(&sdk.Restrictions{
			Secrets: &sdk.SecretsRestritions{
				MaxSecrets: 10,
				Restrictions: []*sdk.SecretRestriction{
					{Restriction: &sdk.SecretRestriction_ExactSecret{
						ExactSecret: &sdk.Secret{Id: "db-password", Namespace: "infra"},
					}},
				},
			},
		})
		assert.False(t, er.canGetSecret(&sdk.SecretRequest{Id: "api-key", Namespace: "external"}))
	})

	t.Run("multiple restrictions match independently", func(t *testing.T) {
		er := newExecutionRestrictions(&sdk.Restrictions{
			Secrets: &sdk.SecretsRestritions{
				MaxSecrets: 10,
				Restrictions: []*sdk.SecretRestriction{
					{Restriction: &sdk.SecretRestriction_ExactSecret{
						ExactSecret: &sdk.Secret{Id: "db-password", Namespace: "infra"},
					}},
					{Restriction: &sdk.SecretRestriction_PrefixedSecret{
						PrefixedSecret: &sdk.SecretPrefixRestriction{
							Prefix: "api-", Namespace: "external", MaxSecrets: 5,
						},
					}},
				},
			},
		})

		assert.True(t, er.canGetSecret(&sdk.SecretRequest{Id: "db-password", Namespace: "infra"}))
		assert.True(t, er.canGetSecret(&sdk.SecretRequest{Id: "api-key", Namespace: "external"}))
		assert.False(t, er.canGetSecret(&sdk.SecretRequest{Id: "api-key", Namespace: "infra"}))
		assert.False(t, er.canGetSecret(&sdk.SecretRequest{Id: "other", Namespace: "external"}))
	})

	t.Run("global max secrets reached before individual limit", func(t *testing.T) {
		er := newExecutionRestrictions(&sdk.Restrictions{
			Secrets: &sdk.SecretsRestritions{
				MaxSecrets: 1,
				Restrictions: []*sdk.SecretRestriction{
					{Restriction: &sdk.SecretRestriction_ExactSecret{
						ExactSecret: &sdk.Secret{Id: "secret-a", Namespace: "ns"},
					}},
					{Restriction: &sdk.SecretRestriction_ExactSecret{
						ExactSecret: &sdk.Secret{Id: "secret-b", Namespace: "ns"},
					}},
				},
			},
		})

		assert.True(t, er.canGetSecret(&sdk.SecretRequest{Id: "secret-a", Namespace: "ns"}))
		assert.False(t, er.canGetSecret(&sdk.SecretRequest{Id: "secret-b", Namespace: "ns"}))
	})

	t.Run("negative max secrets means unlimited", func(t *testing.T) {
		er := newExecutionRestrictions(&sdk.Restrictions{
			Secrets: &sdk.SecretsRestritions{
				MaxSecrets: -1,
				Restrictions: []*sdk.SecretRestriction{
					{Restriction: &sdk.SecretRestriction_ExactSecret{
						ExactSecret: &sdk.Secret{Id: "db-password", Namespace: "infra"},
					}},
				},
			},
		})

		for i := 0; i < 100; i++ {
			assert.True(t, er.canGetSecret(&sdk.SecretRequest{Id: "db-password", Namespace: "infra"}))
		}
		assert.Equal(t, int32(-1), er.maxSecrets)
	})

	t.Run("negative prefix max secrets means unlimited for that prefix", func(t *testing.T) {
		er := newExecutionRestrictions(&sdk.Restrictions{
			Secrets: &sdk.SecretsRestritions{
				MaxSecrets: 3,
				Restrictions: []*sdk.SecretRestriction{
					{Restriction: &sdk.SecretRestriction_PrefixedSecret{
						PrefixedSecret: &sdk.SecretPrefixRestriction{
							Prefix: "db-", Namespace: "infra", MaxSecrets: -1,
						},
					}},
				},
			},
		})

		for i := 0; i < 3; i++ {
			assert.True(t, er.canGetSecret(&sdk.SecretRequest{Id: "db-password", Namespace: "infra"}))
		}
		assert.False(t, er.canGetSecret(&sdk.SecretRequest{Id: "db-password", Namespace: "infra"}))
		assert.Equal(t, int32(-1), er.prefixSecrets[0].maxCalls)
	})

	t.Run("secrets configured with only max secrets denies unmatched", func(t *testing.T) {
		er := newExecutionRestrictions(&sdk.Restrictions{
			Secrets: &sdk.SecretsRestritions{
				MaxSecrets: 10,
			},
		})
		assert.False(t, er.canGetSecret(&sdk.SecretRequest{Id: "any-secret", Namespace: "ns"}))
	})

	t.Run("secrets configured with zero max secrets denies even matched", func(t *testing.T) {
		er := newExecutionRestrictions(&sdk.Restrictions{
			Secrets: &sdk.SecretsRestritions{
				MaxSecrets: 0,
			},
		})
		assert.False(t, er.canGetSecret(&sdk.SecretRequest{Id: "any-secret", Namespace: "ns"}))
	})

	t.Run("exact match still respects and decrements prefix limits", func(t *testing.T) {
		er := newExecutionRestrictions(&sdk.Restrictions{
			Secrets: &sdk.SecretsRestritions{
				MaxSecrets: -1,
				Restrictions: []*sdk.SecretRestriction{
					{Restriction: &sdk.SecretRestriction_ExactSecret{
						ExactSecret: &sdk.Secret{Id: "db-password", Namespace: "infra"},
					}},
					{Restriction: &sdk.SecretRestriction_PrefixedSecret{
						PrefixedSecret: &sdk.SecretPrefixRestriction{
							Prefix: "db-", Namespace: "infra", MaxSecrets: 2,
						},
					}},
				},
			},
		})

		assert.True(t, er.canGetSecret(&sdk.SecretRequest{Id: "db-password", Namespace: "infra"}))
		assert.Equal(t, int32(1), er.prefixSecrets[0].maxCalls,
			"prefix maxCalls must decrement even though exact matched")

		assert.True(t, er.canGetSecret(&sdk.SecretRequest{Id: "db-password", Namespace: "infra"}))
		assert.Equal(t, int32(0), er.prefixSecrets[0].maxCalls)

		assert.False(t, er.canGetSecret(&sdk.SecretRequest{Id: "db-password", Namespace: "infra"}),
			"denied: prefix exhausted even though exact key exists")
	})

	t.Run("exact match denied when covering prefix has zero calls", func(t *testing.T) {
		er := newExecutionRestrictions(&sdk.Restrictions{
			Secrets: &sdk.SecretsRestritions{
				MaxSecrets: 10,
				Restrictions: []*sdk.SecretRestriction{
					{Restriction: &sdk.SecretRestriction_ExactSecret{
						ExactSecret: &sdk.Secret{Id: "db-password", Namespace: "infra"},
					}},
					{Restriction: &sdk.SecretRestriction_PrefixedSecret{
						PrefixedSecret: &sdk.SecretPrefixRestriction{
							Prefix: "db-", Namespace: "infra", MaxSecrets: 0,
						},
					}},
				},
			},
		})

		assert.False(t, er.canGetSecret(&sdk.SecretRequest{Id: "db-password", Namespace: "infra"}),
			"denied: covering prefix has zero budget, exact match is irrelevant")
	})

	t.Run("exact match without covering prefix still works", func(t *testing.T) {
		er := newExecutionRestrictions(&sdk.Restrictions{
			Secrets: &sdk.SecretsRestritions{
				MaxSecrets: 10,
				Restrictions: []*sdk.SecretRestriction{
					{Restriction: &sdk.SecretRestriction_ExactSecret{
						ExactSecret: &sdk.Secret{Id: "db-password", Namespace: "infra"},
					}},
					{Restriction: &sdk.SecretRestriction_PrefixedSecret{
						PrefixedSecret: &sdk.SecretPrefixRestriction{
							Prefix: "api-", Namespace: "external", MaxSecrets: 5,
						},
					}},
				},
			},
		})

		assert.True(t, er.canGetSecret(&sdk.SecretRequest{Id: "db-password", Namespace: "infra"}))
		assert.Equal(t, int32(5), er.prefixSecrets[0].maxCalls,
			"unrelated prefix not decremented")
		assert.Equal(t, int32(9), er.maxSecrets)
	})

	t.Run("multiple overlapping prefixes all decrement on match", func(t *testing.T) {
		er := newExecutionRestrictions(&sdk.Restrictions{
			Secrets: &sdk.SecretsRestritions{
				MaxSecrets: -1,
				Restrictions: []*sdk.SecretRestriction{
					{Restriction: &sdk.SecretRestriction_PrefixedSecret{
						PrefixedSecret: &sdk.SecretPrefixRestriction{
							Prefix: "db-", Namespace: "infra", MaxSecrets: 3,
						},
					}},
					{Restriction: &sdk.SecretRestriction_PrefixedSecret{
						PrefixedSecret: &sdk.SecretPrefixRestriction{
							Prefix: "db-pass", Namespace: "infra", MaxSecrets: 1,
						},
					}},
				},
			},
		})

		assert.True(t, er.canGetSecret(&sdk.SecretRequest{Id: "db-password", Namespace: "infra"}))
		assert.Equal(t, int32(2), er.prefixSecrets[0].maxCalls)
		assert.Equal(t, int32(0), er.prefixSecrets[1].maxCalls)

		assert.False(t, er.canGetSecret(&sdk.SecretRequest{Id: "db-password", Namespace: "infra"}),
			"second call denied: narrower prefix exhausted")

		assert.True(t, er.canGetSecret(&sdk.SecretRequest{Id: "db-host", Namespace: "infra"}),
			"different key only matches broader prefix which still has budget")
	})
}

func TestCallCapWithRestrictions(t *testing.T) {
	restrictions := &sdk.Restrictions{
		Capabilities: &sdk.CapabilityRestrictions{
			MaxTotalCalls: 10,
			Type:          sdk.CapabilityRestrictionType_CAPABILITY_RESTRICTION_TYPE_CLOSED,
			Restrictions: []*sdk.CapabilityRestriction{
				{Restriction: &sdk.CapabilityRestriction_Method{
					Method: &sdk.MethodRestriction{Id: "allowed@1.0.0", Method: "Foo", MaxCalls: 5},
				}},
			},
		},
	}

	t.Run("denies blocked capability", func(t *testing.T) {
		h := newRestrictedExecutionHelper(NewMockExecutionHelper(t), restrictions)
		_, err := h.CallCapability(t.Context(), &sdk.CapabilityRequest{Id: "blocked@1.0.0", Method: "Bar"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "denied by restrictions")
	})

	t.Run("allows permitted capability", func(t *testing.T) {
		inner := NewMockExecutionHelper(t)
		inner.EXPECT().CallCapability(matches.AnyContext, mock.Anything).Return(&sdk.CapabilityResponse{}, nil)
		h := newRestrictedExecutionHelper(inner, restrictions)
		_, err := h.CallCapability(t.Context(), &sdk.CapabilityRequest{Id: "allowed@1.0.0", Method: "Foo"})
		require.NoError(t, err)
	})

	t.Run("no restrictions allows everything", func(t *testing.T) {
		inner := NewMockExecutionHelper(t)
		inner.EXPECT().CallCapability(matches.AnyContext, mock.Anything).Return(&sdk.CapabilityResponse{}, nil)
		h := newRestrictedExecutionHelper(inner, nil)
		_, err := h.CallCapability(t.Context(), &sdk.CapabilityRequest{Id: "anything@1.0.0", Method: "Whatever"})
		require.NoError(t, err)
	})
}

func TestGetSecretsWithRestrictions(t *testing.T) {
	restrictions := &sdk.Restrictions{
		Secrets: &sdk.SecretsRestritions{
			MaxSecrets: 10,
			Restrictions: []*sdk.SecretRestriction{
				{Restriction: &sdk.SecretRestriction_ExactSecret{
					ExactSecret: &sdk.Secret{Id: "allowed-secret", Namespace: "ns"},
				}},
			},
		},
	}

	t.Run("blocked secret returns error response without calling inner", func(t *testing.T) {
		h := newRestrictedExecutionHelper(NewMockExecutionHelper(t), restrictions)
		resp, err := h.GetSecrets(t.Context(), &sdk.GetSecretsRequest{
			Requests: []*sdk.SecretRequest{{Id: "blocked-secret", Namespace: "ns"}},
		})
		require.NoError(t, err)
		require.Len(t, resp, 1)
		errResp := resp[0].GetError()
		require.NotNil(t, errResp)
		assert.Contains(t, errResp.Error, "denied by restrictions")
	})

	t.Run("allows permitted secret", func(t *testing.T) {
		inner := NewMockExecutionHelper(t)
		inner.EXPECT().GetSecrets(matches.AnyContext, mock.Anything).Return([]*sdk.SecretResponse{}, nil)
		h := newRestrictedExecutionHelper(inner, restrictions)
		_, err := h.GetSecrets(t.Context(), &sdk.GetSecretsRequest{
			Requests: []*sdk.SecretRequest{{Id: "allowed-secret", Namespace: "ns"}},
		})
		require.NoError(t, err)
	})

	t.Run("mixed batch: blocked gets error response, allowed goes to inner", func(t *testing.T) {
		inner := NewMockExecutionHelper(t)
		inner.EXPECT().GetSecrets(matches.AnyContext, mock.MatchedBy(func(r *sdk.GetSecretsRequest) bool {
			return len(r.Requests) == 1 && r.Requests[0].Id == "allowed-secret"
		})).Return([]*sdk.SecretResponse{{}}, nil)
		h := newRestrictedExecutionHelper(inner, restrictions)
		resp, err := h.GetSecrets(t.Context(), &sdk.GetSecretsRequest{
			Requests: []*sdk.SecretRequest{
				{Id: "allowed-secret", Namespace: "ns"},
				{Id: "blocked-secret", Namespace: "ns"},
			},
		})
		require.NoError(t, err)
		require.Len(t, resp, 2)
	})
}
