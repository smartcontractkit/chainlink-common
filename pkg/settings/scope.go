package settings

import (
	"context"
	"fmt"

	"github.com/smartcontractkit/chainlink-common/pkg/contexts"
)

// Scope represents the levels at which settings can be overridden and applied.
type Scope int

const (
	ScopeGlobal Scope = iota
	ScopeOrg
	ScopeOwner
	ScopeWorkflow
)

func ParseScope(scope string) (Scope, error) {
	switch scope {
	case "global":
		return ScopeGlobal, nil
	case "org":
		return ScopeOrg, nil
	case "owner":
		return ScopeOwner, nil
	case "workflow":
		return ScopeWorkflow, nil
	default:
		return Scope(-1), fmt.Errorf("invalid scope: %s", scope)
	}
}

func (s Scope) String() string {
	switch s {
	case ScopeGlobal:
		return "global"
	case ScopeOrg:
		return "org"
	case ScopeOwner:
		return "owner"
	case ScopeWorkflow:
		return "workflow"
	default:
		return fmt.Sprintf("unknown Scope(%d)", s)
	}
}

// Value gets the tenant for this scope from ctx, or return empty string if not set.
func (s Scope) Value(ctx context.Context) string {
	cre := contexts.CREValue(ctx)
	switch s {
	case ScopeGlobal:
		return ""
	case ScopeOrg:
		return cre.Org
	case ScopeOwner:
		return cre.Owner
	case ScopeWorkflow:
		return cre.Workflow
	default:
		return ""
	}
}

// RoundCRE returns a modified CRE with out-of-scope tenants omitted.
func (s Scope) RoundCRE(c contexts.CRE) contexts.CRE {
	switch s {
	case ScopeGlobal:
		c.Org = ""
		fallthrough
	case ScopeOrg:
		c.Owner = ""
		fallthrough
	case ScopeOwner:
		c.Workflow = ""
		fallthrough
	case ScopeWorkflow:
	}
	return c
}

func (s Scope) IsTenantRequired() bool {
	return s != ScopeOrg // org may not be available
}
