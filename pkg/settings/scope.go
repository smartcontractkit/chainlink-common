package settings

import (
	"context"
	"fmt"

	"github.com/smartcontractkit/chainlink-common/pkg/contexts"
)

// GetScoped returns the value for key in the given scope, or an error if none is found.
// TODO special parse error? Or strings here?
type GetScoped[N any] func(ctx context.Context, scope Scope, key string) (value N, err error)

// Scope represents the levels at which settings can be overridden and applied.
type Scope int

const (
	ScopeGlobal Scope = iota //TODO better name?
	ScopeOrg
	ScopeUser
	ScopeWorkflow
)

func (s Scope) String() string {
	switch s {
	case ScopeGlobal:
		return "global"
	case ScopeOrg:
		return "organization"
	case ScopeUser:
		return "user"
	case ScopeWorkflow:
		return "workflow"
	default:
		return fmt.Sprintf("unknown Scope(%d)", s)
	}
}

// Value gets the tenant for this scope from ctx, or return empty string if not set.
func (s Scope) Value(ctx context.Context) string {
	switch s {
	case ScopeGlobal:
		return ""
	case ScopeOrg:
		return contexts.OrgValue(ctx)
	case ScopeUser:
		return contexts.UserValue(ctx)
	case ScopeWorkflow:
		return contexts.WorkflowValue(ctx)
	default:
		return ""
	}
}
