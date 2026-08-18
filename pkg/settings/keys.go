package settings

import (
	"context"
	"errors"
	"fmt"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// rawKeys returns a slice of keys to check for this scope, ordered by decreasing priority.
// Example: []string{"org.AcmeCorp.<key>", "<key>"}
// Example: []string{"owner.0xabcd.<key>", "<key>"}
func (s Scope) rawKeys(ctx context.Context, key string) (keys []string, err error) {
	// start with this Scope, and add each enclosing scope, except global.
	for i := s; i > ScopeGlobal; i-- {
		tenant := i.Value(ctx)
		if tenant == "" {
			err = errors.Join(err, tenantMissingError{Scope: i})
		} else {
			keys = append(keys, i.String()+"."+tenant+"."+key)
		}
	}
	keys = append(keys, ScopeGlobal.String()+"."+key) // ScopeGlobal
	return
}

type tenantMissingError struct {
	Scope Scope
}

func (t tenantMissingError) Error() string {
	return fmt.Sprintf("missing tenant for scope: %s", t.Scope)
}

type GetterConfig struct {
	Logger logger.Logger
}
