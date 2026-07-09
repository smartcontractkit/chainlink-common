package orgresolver

import (
	"context"

	log "github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// ResolveOrEmpty resolves owner to an organization ID, returning "" instead of
// an error. It is the canonical, fail-open resolution helper for metering
// producers: a nil resolver or empty owner yields "", a resolver error is
// logged and yields "", and a panic in the resolver is recovered and logged.
// Org attribution must never gate or fail the operation being metered.
func ResolveOrEmpty(ctx context.Context, resolver OrgResolver, owner string, lggr log.Logger) (orgID string) {
	if resolver == nil || owner == "" {
		return ""
	}
	sugared := log.Sugared(lggr)
	defer func() {
		if r := recover(); r != nil {
			sugared.Warnw("panic while resolving organization ID", "owner", owner, "panic", r)
			orgID = ""
		}
	}()
	resolved, err := resolver.Get(ctx, owner)
	if err != nil {
		sugared.Warnw("failed to resolve organization ID", "owner", owner, "error", err)
		return ""
	}
	return resolved
}
