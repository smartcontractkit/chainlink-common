package host_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"google.golang.org/protobuf/types/known/anypb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/actions/vault"
	caperrors "github.com/smartcontractkit/chainlink-common/pkg/capabilities/errors"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/actions/confidentialhttp"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/matches"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/host"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/host/mocks"
	"github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
)

// stubEncryptionKeyFetcher is a no-op EncryptionKeyFetcher used to verify the fetcher is
// delegated through to the inner helper.
type stubEncryptionKeyFetcher struct{}

func (stubEncryptionKeyFetcher) GetEncryptionKeys(context.Context) ([]string, error) {
	return nil, nil
}

// capabilitySequence drives CallCapability through the public API, in order, against a
// single restricted helper. It returns, for each request, whether the call was allowed
// through to the inner helper (true) or denied by the restrictions (false).
func capabilitySequence(t *testing.T, r *sdk.Restrictions, reqs ...*sdk.CapabilityRequest) []bool {
	t.Helper()
	inner := mocks.NewMockExecutionHelper(t)
	inner.EXPECT().CallCapability(matches.AnyContext, mock.Anything).
		Return(&sdk.CapabilityResponse{}, nil).Maybe()
	h := host.NewRestrictedExecutionHelper(inner, r)

	allowed := make([]bool, len(reqs))
	for i, req := range reqs {
		_, err := h.CallCapability(t.Context(), req)
		allowed[i] = err == nil
	}
	return allowed
}

// secretSequence drives GetSecrets (one request per call) through the public API, in
// order, against a single restricted helper. It returns, for each request, whether the
// secret was allowed through to the inner helper (true) or denied by the restrictions
// (false).
func secretSequence(t *testing.T, r *sdk.Restrictions, reqs ...*sdk.SecretRequest) []bool {
	t.Helper()
	inner := mocks.NewMockExecutionHelper(t)
	inner.EXPECT().GetSecrets(matches.AnyContext, mock.Anything).
		Return([]*sdk.SecretResponse{{}}, nil).Maybe()
	h := host.NewRestrictedExecutionHelper(inner, r)

	allowed := make([]bool, len(reqs))
	for i, req := range reqs {
		resp, err := h.GetSecrets(t.Context(), &sdk.GetSecretsRequest{
			Requests: []*sdk.SecretRequest{req},
		})
		require.NoError(t, err)
		require.Len(t, resp, 1)
		// A denied secret is short-circuited into an error response; an allowed one is
		// forwarded to the inner helper which returns a non-error response.
		allowed[i] = resp[0].GetError() == nil
	}
	return allowed
}

func TestRequirementSelectingModule_CallCapWithRestrictions(t *testing.T) {
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

	t.Run("denied call returns a limit-exceeded error without calling inner", func(t *testing.T) {
		inner := mocks.NewMockExecutionHelper(t) // no expectations: inner must not be called
		h := host.NewRestrictedExecutionHelper(inner, restrictions)
		_, err := h.CallCapability(t.Context(), &sdk.CapabilityRequest{Id: "blocked@1.0.0", Method: "Bar"})
		var capErr caperrors.Error
		require.True(t, errors.As(err, &capErr))
		assert.Contains(t, capErr.Error(), "denied by user pre-hook restrictions")
		assert.Equal(t, caperrors.LimitExceeded, capErr.Code())
	})

	t.Run("allowed call reaches inner and returns its response", func(t *testing.T) {
		inner := mocks.NewMockExecutionHelper(t)
		want := &sdk.CapabilityResponse{}
		inner.EXPECT().CallCapability(matches.AnyContext, mock.Anything).Return(want, nil)
		h := host.NewRestrictedExecutionHelper(inner, restrictions)
		got, err := h.CallCapability(t.Context(), &sdk.CapabilityRequest{Id: "allowed@1.0.0", Method: "Foo"})
		require.NoError(t, err)
		assert.Same(t, want, got)
	})

	t.Run("no restrictions allows everything", func(t *testing.T) {
		got := capabilitySequence(t, nil, &sdk.CapabilityRequest{Id: "anything@1.0.0", Method: "Whatever"})
		assert.Equal(t, []bool{true}, got)
	})

	t.Run("allows when no capabilities restrictions set", func(t *testing.T) {
		got := capabilitySequence(t, &sdk.Restrictions{}, &sdk.CapabilityRequest{Id: "cap@1.0.0", Method: "Foo"})
		assert.Equal(t, []bool{true}, got)
	})

	t.Run("closed denies unmatched capability", func(t *testing.T) {
		got := capabilitySequence(t, &sdk.Restrictions{
			Capabilities: &sdk.CapabilityRestrictions{
				MaxTotalCalls: 10,
				Type:          sdk.CapabilityRestrictionType_CAPABILITY_RESTRICTION_TYPE_CLOSED,
				Restrictions: []*sdk.CapabilityRestriction{
					{Restriction: &sdk.CapabilityRestriction_Method{
						Method: &sdk.MethodRestriction{Id: "cap@1.0.0", Method: "Foo", MaxCalls: 5},
					}},
				},
			},
		}, &sdk.CapabilityRequest{Id: "other-cap@1.0.0", Method: "Bar"})
		assert.Equal(t, []bool{false}, got)
	})

	t.Run("closed allows matched capability until method limit is reached", func(t *testing.T) {
		req := &sdk.CapabilityRequest{Id: "cap@1.0.0", Method: "Foo"}
		got := capabilitySequence(t, &sdk.Restrictions{
			Capabilities: &sdk.CapabilityRestrictions{
				MaxTotalCalls: 10,
				Type:          sdk.CapabilityRestrictionType_CAPABILITY_RESTRICTION_TYPE_CLOSED,
				Restrictions: []*sdk.CapabilityRestriction{
					{Restriction: &sdk.CapabilityRestriction_Method{
						Method: &sdk.MethodRestriction{Id: "cap@1.0.0", Method: "Foo", MaxCalls: 2},
					}},
				},
			},
		}, req, req, req)
		assert.Equal(t, []bool{true, true, false}, got)
	})

	t.Run("denies when max total calls is zero", func(t *testing.T) {
		got := capabilitySequence(t, &sdk.Restrictions{
			Capabilities: &sdk.CapabilityRestrictions{
				MaxTotalCalls: 0,
				Type:          sdk.CapabilityRestrictionType_CAPABILITY_RESTRICTION_TYPE_CLOSED,
				Restrictions: []*sdk.CapabilityRestriction{
					{Restriction: &sdk.CapabilityRestriction_Method{
						Method: &sdk.MethodRestriction{Id: "cap@1.0.0", Method: "Foo", MaxCalls: 5},
					}},
				},
			},
		}, &sdk.CapabilityRequest{Id: "cap@1.0.0", Method: "Foo"})
		assert.Equal(t, []bool{false}, got)
	})

	t.Run("open allows unmatched capability", func(t *testing.T) {
		got := capabilitySequence(t, &sdk.Restrictions{
			Capabilities: &sdk.CapabilityRestrictions{
				MaxTotalCalls: 10,
				Type:          sdk.CapabilityRestrictionType_CAPABILITY_RESTRICTION_TYPE_OPEN,
				Restrictions: []*sdk.CapabilityRestriction{
					{Restriction: &sdk.CapabilityRestriction_Method{
						Method: &sdk.MethodRestriction{Id: "cap@1.0.0", Method: "Foo", MaxCalls: 2},
					}},
				},
			},
		}, &sdk.CapabilityRequest{Id: "other-cap@1.0.0", Method: "Bar"})
		assert.Equal(t, []bool{true}, got)
	})

	t.Run("denies when matched method has zero calls remaining", func(t *testing.T) {
		got := capabilitySequence(t, &sdk.Restrictions{
			Capabilities: &sdk.CapabilityRestrictions{
				MaxTotalCalls: 10,
				Type:          sdk.CapabilityRestrictionType_CAPABILITY_RESTRICTION_TYPE_OPEN,
				Restrictions: []*sdk.CapabilityRestriction{
					{Restriction: &sdk.CapabilityRestriction_Method{
						Method: &sdk.MethodRestriction{Id: "cap@1.0.0", Method: "Foo", MaxCalls: 0},
					}},
				},
			},
		}, &sdk.CapabilityRequest{Id: "cap@1.0.0", Method: "Foo"})
		assert.Equal(t, []bool{false}, got)
	})

	t.Run("matches by both id and method", func(t *testing.T) {
		got := capabilitySequence(t, &sdk.Restrictions{
			Capabilities: &sdk.CapabilityRestrictions{
				MaxTotalCalls: 10,
				Type:          sdk.CapabilityRestrictionType_CAPABILITY_RESTRICTION_TYPE_CLOSED,
				Restrictions: []*sdk.CapabilityRestriction{
					{Restriction: &sdk.CapabilityRestriction_Method{
						Method: &sdk.MethodRestriction{Id: "cap@1.0.0", Method: "Foo", MaxCalls: 5},
					}},
				},
			},
		},
			&sdk.CapabilityRequest{Id: "cap@1.0.0", Method: "Bar"},
			&sdk.CapabilityRequest{Id: "cap@2.0.0", Method: "Foo"},
		)
		assert.Equal(t, []bool{false, false}, got)
	})

	t.Run("multiple different methods match independently", func(t *testing.T) {
		got := capabilitySequence(t, &sdk.Restrictions{
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
		},
			&sdk.CapabilityRequest{Id: "cap@1.0.0", Method: "Foo"},
			&sdk.CapabilityRequest{Id: "cap@1.0.0", Method: "Foo"},
			&sdk.CapabilityRequest{Id: "cap@1.0.0", Method: "Bar"},
			&sdk.CapabilityRequest{Id: "cap@1.0.0", Method: "Bar"},
		)
		assert.Equal(t, []bool{true, false, true, false}, got)
	})

	t.Run("total calls limit reached before method limit", func(t *testing.T) {
		req := &sdk.CapabilityRequest{Id: "cap@1.0.0", Method: "Foo"}
		got := capabilitySequence(t, &sdk.Restrictions{
			Capabilities: &sdk.CapabilityRestrictions{
				MaxTotalCalls: 2,
				Type:          sdk.CapabilityRestrictionType_CAPABILITY_RESTRICTION_TYPE_CLOSED,
				Restrictions: []*sdk.CapabilityRestriction{
					{Restriction: &sdk.CapabilityRestriction_Method{
						Method: &sdk.MethodRestriction{Id: "cap@1.0.0", Method: "Foo", MaxCalls: 100},
					}},
				},
			},
		}, req, req, req)
		assert.Equal(t, []bool{true, true, false}, got)
	})

	t.Run("negative max total calls means unlimited (method limit still applies)", func(t *testing.T) {
		req := &sdk.CapabilityRequest{Id: "cap@1.0.0", Method: "Foo"}
		got := capabilitySequence(t, &sdk.Restrictions{
			Capabilities: &sdk.CapabilityRestrictions{
				MaxTotalCalls: -1,
				Type:          sdk.CapabilityRestrictionType_CAPABILITY_RESTRICTION_TYPE_CLOSED,
				Restrictions: []*sdk.CapabilityRestriction{
					{Restriction: &sdk.CapabilityRestriction_Method{
						Method: &sdk.MethodRestriction{Id: "cap@1.0.0", Method: "Foo", MaxCalls: 2},
					}},
				},
			},
		}, req, req, req)
		assert.Equal(t, []bool{true, true, false}, got)
	})

	t.Run("negative max calls on method means unlimited (total limit still applies)", func(t *testing.T) {
		req := &sdk.CapabilityRequest{Id: "cap@1.0.0", Method: "Foo"}
		got := capabilitySequence(t, &sdk.Restrictions{
			Capabilities: &sdk.CapabilityRestrictions{
				MaxTotalCalls: 3,
				Type:          sdk.CapabilityRestrictionType_CAPABILITY_RESTRICTION_TYPE_CLOSED,
				Restrictions: []*sdk.CapabilityRestriction{
					{Restriction: &sdk.CapabilityRestriction_Method{
						Method: &sdk.MethodRestriction{Id: "cap@1.0.0", Method: "Foo", MaxCalls: -1},
					}},
				},
			},
		}, req, req, req, req)
		assert.Equal(t, []bool{true, true, true, false}, got)
	})

	t.Run("duplicate restrictions keep smallest non-negative value", func(t *testing.T) {
		req := &sdk.CapabilityRequest{Id: "cap@1.0.0", Method: "Foo"}
		got := capabilitySequence(t, &sdk.Restrictions{
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
		}, req, req, req)
		assert.Equal(t, []bool{true, true, false}, got)
	})

	t.Run("duplicate restrictions non-negative overrides negative", func(t *testing.T) {
		req := &sdk.CapabilityRequest{Id: "cap@1.0.0", Method: "Foo"}
		got := capabilitySequence(t, &sdk.Restrictions{
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
		}, req, req, req, req)
		assert.Equal(t, []bool{true, true, true, false}, got)
	})

	t.Run("duplicate restrictions zero overrides positive", func(t *testing.T) {
		got := capabilitySequence(t, &sdk.Restrictions{
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
		}, &sdk.CapabilityRequest{Id: "cap@1.0.0", Method: "Foo"})
		assert.Equal(t, []bool{false}, got)
	})

	t.Run("closed with no methods denies all", func(t *testing.T) {
		got := capabilitySequence(t, &sdk.Restrictions{
			Capabilities: &sdk.CapabilityRestrictions{
				MaxTotalCalls: -1,
				Type:          sdk.CapabilityRestrictionType_CAPABILITY_RESTRICTION_TYPE_CLOSED,
			},
		}, &sdk.CapabilityRequest{Id: "cap@1.0.0", Method: "Foo"})
		assert.Equal(t, []bool{false}, got)
	})

	t.Run("open with no methods respects max total calls", func(t *testing.T) {
		got := capabilitySequence(t, &sdk.Restrictions{
			Capabilities: &sdk.CapabilityRestrictions{
				MaxTotalCalls: 2,
				Type:          sdk.CapabilityRestrictionType_CAPABILITY_RESTRICTION_TYPE_OPEN,
			},
		},
			&sdk.CapabilityRequest{Id: "cap@1.0.0", Method: "Foo"},
			&sdk.CapabilityRequest{Id: "cap@2.0.0", Method: "Bar"},
			&sdk.CapabilityRequest{Id: "cap@3.0.0", Method: "Baz"},
		)
		assert.Equal(t, []bool{true, true, false}, got)
	})

	t.Run("open with zero max total calls denies all", func(t *testing.T) {
		got := capabilitySequence(t, &sdk.Restrictions{
			Capabilities: &sdk.CapabilityRestrictions{
				MaxTotalCalls: 0,
				Type:          sdk.CapabilityRestrictionType_CAPABILITY_RESTRICTION_TYPE_OPEN,
			},
		}, &sdk.CapabilityRequest{Id: "cap@1.0.0", Method: "Foo"})
		assert.Equal(t, []bool{false}, got)
	})
}

// confidentialHTTPRequest builds a CapabilityRequest whose payload is a
// ConfidentialHTTPRequest referencing the given vault DON secrets. The payload's
// type URL is what drives the secret-reservation branch in reserveCapabilityCall.
func confidentialHTTPRequest(t *testing.T, id, method string, secrets ...*confidentialhttp.SecretIdentifier) *sdk.CapabilityRequest {
	t.Helper()
	payload, err := anypb.New(&confidentialhttp.ConfidentialHTTPRequest{VaultDonSecrets: secrets})
	require.NoError(t, err)
	return &sdk.CapabilityRequest{Id: id, Method: method, Payload: payload}
}

func TestRequirementSelectingModule_ConfidentialHTTPWithRestrictions(t *testing.T) {
	// Restrictions that allow the confidential HTTP capability method and a single
	// exact secret. A confidential HTTP call only succeeds if both the capability
	// method and every vault DON secret it references are permitted.
	restrictions := func() *sdk.Restrictions {
		return &sdk.Restrictions{
			Capabilities: &sdk.CapabilityRestrictions{
				MaxTotalCalls: 10,
				Type:          sdk.CapabilityRestrictionType_CAPABILITY_RESTRICTION_TYPE_CLOSED,
				Restrictions: []*sdk.CapabilityRestriction{
					{Restriction: &sdk.CapabilityRestriction_Method{
						Method: &sdk.MethodRestriction{Id: "confhttp@1.0.0", Method: "Call", MaxCalls: 5},
					}},
				},
			},
			Secrets: &sdk.SecretsRestritions{
				MaxSecrets: 10,
				Restrictions: []*sdk.SecretRestriction{
					{Restriction: &sdk.SecretRestriction_ExactSecret{
						ExactSecret: &sdk.Secret{Id: "allowed-secret", Namespace: "ns"},
					}},
				},
			},
		}
	}

	t.Run("allowed confidential http call reaches inner", func(t *testing.T) {
		inner := mocks.NewMockExecutionHelper(t)
		want := &sdk.CapabilityResponse{}
		inner.EXPECT().CallCapability(matches.AnyContext, mock.Anything).Return(want, nil)
		h := host.NewRestrictedExecutionHelper(inner, restrictions())

		req := confidentialHTTPRequest(t, "confhttp@1.0.0", "Call",
			&confidentialhttp.SecretIdentifier{Key: "allowed-secret", Namespace: "ns"})
		got, err := h.CallCapability(t.Context(), req)
		require.NoError(t, err)
		assert.Same(t, want, got)
	})

	t.Run("nil payload is not treated as confidential http and reaches inner", func(t *testing.T) {
		// A capability call sharing the confidential-http method id but carrying no
		// payload must skip the vault-secret reservation branch entirely (guarded by
		// request.Payload != nil) and fall through to the normal method check.
		inner := mocks.NewMockExecutionHelper(t)
		want := &sdk.CapabilityResponse{}
		inner.EXPECT().CallCapability(matches.AnyContext, mock.Anything).Return(want, nil)
		h := host.NewRestrictedExecutionHelper(inner, restrictions())

		got, err := h.CallCapability(t.Context(), &sdk.CapabilityRequest{Id: "confhttp@1.0.0", Method: "Call"})
		require.NoError(t, err)
		assert.Same(t, want, got)
	})

	t.Run("disallowed confidential http call is denied without calling inner", func(t *testing.T) {
		inner := mocks.NewMockExecutionHelper(t) // no expectations: inner must not be called
		h := host.NewRestrictedExecutionHelper(inner, restrictions())

		req := confidentialHTTPRequest(t, "confhttp@1.0.0", "Call",
			&confidentialhttp.SecretIdentifier{Key: "blocked-secret", Namespace: "ns"})
		_, err := h.CallCapability(t.Context(), req)
		var capErr caperrors.Error
		require.True(t, errors.As(err, &capErr))
		assert.Contains(t, capErr.Error(), "denied by user pre-hook restrictions")
		assert.Equal(t, caperrors.LimitExceeded, capErr.Code())
	})
}

func TestRequirementSelectingModule_GetSecretsWithRestrictions(t *testing.T) {
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
		inner := mocks.NewMockExecutionHelper(t) // no expectations: inner must not be called
		h := host.NewRestrictedExecutionHelper(inner, restrictions)
		resp, err := h.GetSecrets(t.Context(), &sdk.GetSecretsRequest{
			Requests: []*sdk.SecretRequest{{Id: "blocked-secret", Namespace: "ns"}},
		})
		require.NoError(t, err)
		require.Len(t, resp, 1)
		errResp := resp[0].GetError()
		require.NotNil(t, errResp)
		assert.Contains(t, errResp.Error, "denied by user pre-hook restrictions")
	})

	t.Run("allows permitted secret", func(t *testing.T) {
		inner := mocks.NewMockExecutionHelper(t)
		inner.EXPECT().GetSecrets(matches.AnyContext, mock.Anything).Return([]*sdk.SecretResponse{}, nil)
		h := host.NewRestrictedExecutionHelper(inner, restrictions)
		_, err := h.GetSecrets(t.Context(), &sdk.GetSecretsRequest{
			Requests: []*sdk.SecretRequest{{Id: "allowed-secret", Namespace: "ns"}},
		})
		require.NoError(t, err)
	})

	t.Run("mixed batch: blocked gets error response, allowed goes to inner", func(t *testing.T) {
		inner := mocks.NewMockExecutionHelper(t)
		inner.EXPECT().GetSecrets(matches.AnyContext, mock.MatchedBy(func(r *sdk.GetSecretsRequest) bool {
			return len(r.Requests) == 1 && r.Requests[0].Id == "allowed-secret"
		})).Return([]*sdk.SecretResponse{{}}, nil)
		h := host.NewRestrictedExecutionHelper(inner, restrictions)
		resp, err := h.GetSecrets(t.Context(), &sdk.GetSecretsRequest{
			Requests: []*sdk.SecretRequest{
				{Id: "allowed-secret", Namespace: "ns"},
				{Id: "blocked-secret", Namespace: "ns"},
			},
		})
		require.NoError(t, err)
		require.Len(t, resp, 2)
	})

	t.Run("allows when nil restrictions", func(t *testing.T) {
		got := secretSequence(t, nil, &sdk.SecretRequest{Id: "my-secret", Namespace: "ns"})
		assert.Equal(t, []bool{true}, got)
	})

	t.Run("allows when no secrets restrictions set", func(t *testing.T) {
		got := secretSequence(t, &sdk.Restrictions{}, &sdk.SecretRequest{Id: "my-secret", Namespace: "ns"})
		assert.Equal(t, []bool{true}, got)
	})

	t.Run("denies when max secrets is zero", func(t *testing.T) {
		got := secretSequence(t, &sdk.Restrictions{
			Secrets: &sdk.SecretsRestritions{
				MaxSecrets: 0,
				Restrictions: []*sdk.SecretRestriction{
					{Restriction: &sdk.SecretRestriction_ExactSecret{
						ExactSecret: &sdk.Secret{Id: "my-secret", Namespace: "ns"},
					}},
				},
			},
		}, &sdk.SecretRequest{Id: "my-secret", Namespace: "ns"})
		assert.Equal(t, []bool{false}, got)
	})

	t.Run("exact match allows until max secrets is reached", func(t *testing.T) {
		req := &sdk.SecretRequest{Id: "db-password", Namespace: "infra"}
		got := secretSequence(t, &sdk.Restrictions{
			Secrets: &sdk.SecretsRestritions{
				MaxSecrets: 2,
				Restrictions: []*sdk.SecretRestriction{
					{Restriction: &sdk.SecretRestriction_ExactSecret{
						ExactSecret: &sdk.Secret{Id: "db-password", Namespace: "infra"},
					}},
				},
			},
		}, req, req, req)
		assert.Equal(t, []bool{true, true, false}, got)
	})

	t.Run("exact match requires both id and namespace", func(t *testing.T) {
		got := secretSequence(t, &sdk.Restrictions{
			Secrets: &sdk.SecretsRestritions{
				MaxSecrets: 10,
				Restrictions: []*sdk.SecretRestriction{
					{Restriction: &sdk.SecretRestriction_ExactSecret{
						ExactSecret: &sdk.Secret{Id: "db-password", Namespace: "infra"},
					}},
				},
			},
		},
			&sdk.SecretRequest{Id: "db-password", Namespace: "other"},
			&sdk.SecretRequest{Id: "other", Namespace: "infra"},
			&sdk.SecretRequest{Id: "db-password", Namespace: "infra"},
		)
		assert.Equal(t, []bool{false, false, true}, got)
	})

	t.Run("prefix match allows until prefix limit is reached", func(t *testing.T) {
		got := secretSequence(t, &sdk.Restrictions{
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
		},
			&sdk.SecretRequest{Id: "db-password", Namespace: "infra"},
			&sdk.SecretRequest{Id: "db-host", Namespace: "infra"},
			&sdk.SecretRequest{Id: "db-port", Namespace: "infra"},
		)
		assert.Equal(t, []bool{true, true, false}, got)
	})

	t.Run("prefix match requires namespace match", func(t *testing.T) {
		got := secretSequence(t, &sdk.Restrictions{
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
		}, &sdk.SecretRequest{Id: "db-password", Namespace: "other"})
		assert.Equal(t, []bool{false}, got)
	})

	t.Run("prefix match denied when global max secrets hits zero", func(t *testing.T) {
		got := secretSequence(t, &sdk.Restrictions{
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
		},
			&sdk.SecretRequest{Id: "db-password", Namespace: "infra"},
			&sdk.SecretRequest{Id: "db-host", Namespace: "infra"},
		)
		assert.Equal(t, []bool{true, false}, got)
	})

	t.Run("denies unmatched secret", func(t *testing.T) {
		got := secretSequence(t, &sdk.Restrictions{
			Secrets: &sdk.SecretsRestritions{
				MaxSecrets: 10,
				Restrictions: []*sdk.SecretRestriction{
					{Restriction: &sdk.SecretRestriction_ExactSecret{
						ExactSecret: &sdk.Secret{Id: "db-password", Namespace: "infra"},
					}},
				},
			},
		}, &sdk.SecretRequest{Id: "api-key", Namespace: "external"})
		assert.Equal(t, []bool{false}, got)
	})

	t.Run("multiple restrictions match independently", func(t *testing.T) {
		got := secretSequence(t, &sdk.Restrictions{
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
		},
			&sdk.SecretRequest{Id: "db-password", Namespace: "infra"},
			&sdk.SecretRequest{Id: "api-key", Namespace: "external"},
			&sdk.SecretRequest{Id: "api-key", Namespace: "infra"},
			&sdk.SecretRequest{Id: "other", Namespace: "external"},
		)
		assert.Equal(t, []bool{true, true, false, false}, got)
	})

	t.Run("global max secrets reached before individual limit", func(t *testing.T) {
		got := secretSequence(t, &sdk.Restrictions{
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
		},
			&sdk.SecretRequest{Id: "secret-a", Namespace: "ns"},
			&sdk.SecretRequest{Id: "secret-b", Namespace: "ns"},
		)
		assert.Equal(t, []bool{true, false}, got)
	})

	t.Run("negative max secrets means unlimited", func(t *testing.T) {
		reqs := make([]*sdk.SecretRequest, 100)
		for i := range reqs {
			reqs[i] = &sdk.SecretRequest{Id: "db-password", Namespace: "infra"}
		}
		got := secretSequence(t, &sdk.Restrictions{
			Secrets: &sdk.SecretsRestritions{
				MaxSecrets: -1,
				Restrictions: []*sdk.SecretRestriction{
					{Restriction: &sdk.SecretRestriction_ExactSecret{
						ExactSecret: &sdk.Secret{Id: "db-password", Namespace: "infra"},
					}},
				},
			},
		}, reqs...)
		for i, allowed := range got {
			assert.Truef(t, allowed, "call %d should be allowed", i)
		}
	})

	t.Run("negative prefix max secrets means unlimited for that prefix (global limit applies)", func(t *testing.T) {
		req := &sdk.SecretRequest{Id: "db-password", Namespace: "infra"}
		got := secretSequence(t, &sdk.Restrictions{
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
		}, req, req, req, req)
		assert.Equal(t, []bool{true, true, true, false}, got)
	})

	t.Run("secrets configured with only max secrets denies unmatched", func(t *testing.T) {
		got := secretSequence(t, &sdk.Restrictions{
			Secrets: &sdk.SecretsRestritions{
				MaxSecrets: 10,
			},
		}, &sdk.SecretRequest{Id: "any-secret", Namespace: "ns"})
		assert.Equal(t, []bool{false}, got)
	})

	t.Run("secrets configured with zero max secrets denies even matched", func(t *testing.T) {
		got := secretSequence(t, &sdk.Restrictions{
			Secrets: &sdk.SecretsRestritions{
				MaxSecrets: 0,
			},
		}, &sdk.SecretRequest{Id: "any-secret", Namespace: "ns"})
		assert.Equal(t, []bool{false}, got)
	})

	t.Run("exact match still respects and decrements covering prefix limits", func(t *testing.T) {
		req := &sdk.SecretRequest{Id: "db-password", Namespace: "infra"}
		got := secretSequence(t, &sdk.Restrictions{
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
		}, req, req, req)
		assert.Equal(t, []bool{true, true, false}, got)
	})

	t.Run("exact match denied when covering prefix has zero calls", func(t *testing.T) {
		got := secretSequence(t, &sdk.Restrictions{
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
		}, &sdk.SecretRequest{Id: "db-password", Namespace: "infra"})
		assert.Equal(t, []bool{false}, got)
	})

	t.Run("exact match without covering prefix still works", func(t *testing.T) {
		got := secretSequence(t, &sdk.Restrictions{
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
		}, &sdk.SecretRequest{Id: "db-password", Namespace: "infra"})
		assert.Equal(t, []bool{true}, got)
	})

	t.Run("multiple overlapping prefixes all decrement on match", func(t *testing.T) {
		got := secretSequence(t, &sdk.Restrictions{
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
		},
			// First db-password matches both prefixes; the narrower db-pass prefix is then
			// exhausted, so a second db-password is denied while a db-host (only the broader
			// prefix) is still allowed.
			&sdk.SecretRequest{Id: "db-password", Namespace: "infra"},
			&sdk.SecretRequest{Id: "db-password", Namespace: "infra"},
			&sdk.SecretRequest{Id: "db-host", Namespace: "infra"},
		)
		assert.Equal(t, []bool{true, false, true}, got)
	})
}

func TestRequirementSelectingModule_GetRawSecretsWithRestrictions(t *testing.T) {
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

	fetcher := &stubEncryptionKeyFetcher{}

	newHelper := func(t *testing.T) (*mocks.MockExecutionHelperWithRawSecrets, host.ExecutionHelperWithRawSecrets) {
		inner := mocks.NewMockExecutionHelperWithRawSecrets(t)
		h := host.NewRestrictedExecutionHelper(inner, restrictions).(host.ExecutionHelperWithRawSecrets)
		return inner, h
	}

	t.Run("blocked secret returns error response without calling inner", func(t *testing.T) {
		inner, h := newHelper(t)
		inner.EXPECT().GetOwner().Return("owner-1")

		resp, err := h.GetRawSecrets(t.Context(), &sdk.GetSecretsRequest{
			Requests: []*sdk.SecretRequest{{Id: "blocked-secret", Namespace: "ns"}},
		}, fetcher)
		require.NoError(t, err)
		require.Len(t, resp, 1)
		assert.Contains(t, resp[0].GetError(), "denied by user pre-hook restrictions")
		assert.Equal(t, "blocked-secret", resp[0].GetId().GetKey())
		assert.Equal(t, "ns", resp[0].GetId().GetNamespace())
		assert.Equal(t, "owner-1", resp[0].GetId().GetOwner())
	})

	t.Run("allows permitted secret", func(t *testing.T) {
		inner, h := newHelper(t)
		inner.EXPECT().GetOwner().Return("owner-1")
		inner.EXPECT().GetRawSecrets(matches.AnyContext, mock.MatchedBy(func(r *sdk.GetSecretsRequest) bool {
			return len(r.Requests) == 1 && r.Requests[0].Id == "allowed-secret"
		}), mock.Anything).Return([]*vault.SecretResponse{{}}, nil)

		resp, err := h.GetRawSecrets(t.Context(), &sdk.GetSecretsRequest{
			Requests: []*sdk.SecretRequest{{Id: "allowed-secret", Namespace: "ns"}},
		}, fetcher)
		require.NoError(t, err)
		require.Len(t, resp, 1)
	})

	t.Run("mixed batch: blocked gets error response, allowed goes to inner", func(t *testing.T) {
		inner, h := newHelper(t)
		inner.EXPECT().GetOwner().Return("owner-1")
		inner.EXPECT().GetRawSecrets(matches.AnyContext, mock.MatchedBy(func(r *sdk.GetSecretsRequest) bool {
			return len(r.Requests) == 1 && r.Requests[0].Id == "allowed-secret"
		}), mock.Anything).Return([]*vault.SecretResponse{{}}, nil)

		resp, err := h.GetRawSecrets(t.Context(), &sdk.GetSecretsRequest{
			Requests: []*sdk.SecretRequest{
				{Id: "allowed-secret", Namespace: "ns"},
				{Id: "blocked-secret", Namespace: "ns"},
			},
		}, fetcher)
		require.NoError(t, err)
		require.Len(t, resp, 2)
	})

	t.Run("delegates the encryption key fetcher to inner", func(t *testing.T) {
		inner, h := newHelper(t)
		inner.EXPECT().GetOwner().Return("owner-1")
		var gotFetcher host.EncryptionKeyFetcher
		inner.EXPECT().GetRawSecrets(matches.AnyContext, mock.Anything, mock.Anything).
			RunAndReturn(func(_ context.Context, _ *sdk.GetSecretsRequest, f host.EncryptionKeyFetcher) ([]*vault.SecretResponse, error) {
				gotFetcher = f
				return []*vault.SecretResponse{{}}, nil
			})

		_, err := h.GetRawSecrets(t.Context(), &sdk.GetSecretsRequest{
			Requests: []*sdk.SecretRequest{{Id: "allowed-secret", Namespace: "ns"}},
		}, fetcher)
		require.NoError(t, err)
		assert.Same(t, fetcher, gotFetcher, "the fetcher passed in must be delegated unchanged to the inner helper")
	})

	t.Run("all blocked does not call inner", func(t *testing.T) {
		inner, h := newHelper(t)
		inner.EXPECT().GetOwner().Return("owner-1")

		resp, err := h.GetRawSecrets(t.Context(), &sdk.GetSecretsRequest{
			Requests: []*sdk.SecretRequest{
				{Id: "blocked-a", Namespace: "ns"},
				{Id: "blocked-b", Namespace: "ns"},
			},
		}, fetcher)
		require.NoError(t, err)
		require.Len(t, resp, 2)
		assert.Contains(t, resp[0].GetError(), "denied by user pre-hook restrictions")
		assert.Contains(t, resp[1].GetError(), "denied by user pre-hook restrictions")
	})

	t.Run("inner error is propagated", func(t *testing.T) {
		inner, h := newHelper(t)
		inner.EXPECT().GetOwner().Return("owner-1")
		inner.EXPECT().GetRawSecrets(matches.AnyContext, mock.Anything, mock.Anything).Return(nil, errors.New("boom"))

		resp, err := h.GetRawSecrets(t.Context(), &sdk.GetSecretsRequest{
			Requests: []*sdk.SecretRequest{{Id: "allowed-secret", Namespace: "ns"}},
		}, fetcher)
		require.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestRequirementSelectingModule_GetOwner(t *testing.T) {
	restrictions := &sdk.Restrictions{
		Secrets: &sdk.SecretsRestritions{
			MaxSecrets: 10,
		},
	}

	inner := mocks.NewMockExecutionHelperWithRawSecrets(t)
	inner.EXPECT().GetOwner().Return("owner-123")
	h := host.NewRestrictedExecutionHelper(inner, restrictions).(host.ExecutionHelperWithRawSecrets)

	owner := h.GetOwner()
	assert.Equal(t, "owner-123", owner)

}

func TestRequirementSelectingModule_NewCreatesTheRightInterface(t *testing.T) {
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

	t.Run("normal ExecutionHelper doesn't return ExecutionHelperWithRawSecrets", func(t *testing.T) {
		result := host.NewRestrictedExecutionHelper(mocks.NewMockExecutionHelper(t), restrictions)
		assert.NotImplements(t, (*host.ExecutionHelperWithRawSecrets)(nil), result)
	})

	t.Run("ExecutionHelperWithRawSecrets returns ExecutionHelperWithRawSecrets", func(t *testing.T) {
		result := host.NewRestrictedExecutionHelper(mocks.NewMockExecutionHelperWithRawSecrets(t), restrictions)
		assert.Implements(t, (*host.ExecutionHelperWithRawSecrets)(nil), result)
	})
}
