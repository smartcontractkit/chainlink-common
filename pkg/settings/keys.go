package settings

import (
	"context"
	"errors"
	"fmt"
)

// rawKeys returns a slice of keys to check for this scope, ordered by decreasing priority.
// Example: []string{"org.AcmeCorp.<key>", "<key>"}
// Example: []string{"owner.0xabcd.<key>", "<key>"}
func (s Scope) rawKeys(ctx context.Context, key string) (keys []string, err error) {
	// start with this Scope, and add each enclosing scope, except global.
	for i := s; i > ScopeGlobal; i-- {
		tenant := i.Value(ctx)
		if tenant == "" {
			if i.IsTenantRequired() {
				err = errors.Join(err, fmt.Errorf("empty %s key", i))
			}
		} else {
			keys = append(keys, i.String()+"."+tenant+"."+key)
		}
	}
	keys = append(keys, ScopeGlobal.String()+"."+key) // ScopeGlobal
	return
}
