package settings

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/contexts"
)

func TestTenant_rawKeys(t *testing.T) {
	const (
		org      = "acmecorporation"
		owner    = "1234abcd"
		workflow = "abcdefgh"
		key      = "foo"
	)
	for _, test := range []struct {
		tenant Scope
		expect []string
	}{
		{ScopeWorkflow, []string{
			"workflow." + workflow + "." + key,
			"owner." + owner + "." + key,
			"org." + org + "." + key,
			"global." + key,
		}},
		{ScopeOwner, []string{
			"owner." + owner + "." + key,
			"org." + org + "." + key,
			"global." + key,
		}},
		{ScopeOrg, []string{
			"org." + org + "." + key,
			"global." + key,
		}},
		{ScopeGlobal, []string{"global." + key}},
	} {
		t.Run(test.tenant.String(), func(t *testing.T) {
			ctx := contexts.WithCRE(t.Context(), contexts.CRE{Org: org, Owner: owner, Workflow: workflow})
			got, err := test.tenant.rawKeys(ctx, key)
			require.NoError(t, err)
			require.Equal(t, test.expect, got)
		})
	}
}
