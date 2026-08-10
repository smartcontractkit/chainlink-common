package host

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/settings/cresettings"
	"github.com/smartcontractkit/chainlink-common/pkg/settings/limits"
	sdkpb "github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
)

// TestRegressionPollOneoffRejectsExcessiveSubscriptions is a regression test
// for a host-memory amplification finding in pollOneoff. The implementation
// copies the subscriptions into Go's heap and allocates a proportional buffer for events.
// This means for for every subscription created in the host, we'd allocate 1.67 times
// that much memory on the host. We fix this by rejecting calls with large nsubscriptions.
func TestRegressionPollOneoffRejectsExcessiveSubscriptions(t *testing.T) {
	t.Parallel()

	limiter, err := limits.MakeUpperBoundLimiter(limits.Factory{Logger: logger.Test(t)}, cresettings.Default.WASMPollOneoffSubscriptionLimit)
	require.NoError(t, err)

	cfg := &ModuleConfig{MaxSubscriptionsLimiter: limiter}
	exec := &execution[*sdkpb.ExecutionResult]{ctx: t.Context(), module: &module{cfg: cfg}}
	legacyPollOneoff := createPollOneoff(t.Context(), cfg)

	tests := []struct {
		name           string
		nsubscriptions int32
	}{
		// From the finding: a guest filling a 100MB subscription buffer could
		// request this many clock subscriptions in a single call.
		{"finding's exploit size (100MB guest buffer)", 100_000_000 / subscriptionLen},
		// The old bound was only an int32-overflow guard; both ends of that
		// range must be rejected now too.
		{"just under the old overflow-guard bound", math.MaxInt32/subscriptionLen - 1},
		{"old overflow-guard bound", math.MaxInt32 / subscriptionLen},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run("execution.pollOneoff (V2 path)", func(t *testing.T) {
				var errno int32
				assert.NotPanics(t, func() {
					errno = exec.pollOneoff(nil, 0, 0, tt.nsubscriptions, 0)
				})
				assert.Equal(t, ErrnoInval, errno)
			})

			t.Run("createPollOneoff (legacy DAG path)", func(t *testing.T) {
				var errno int32
				assert.NotPanics(t, func() {
					errno = legacyPollOneoff(nil, 0, 0, tt.nsubscriptions, 0)
				})
				assert.Equal(t, ErrnoInval, errno)
			})
		})
	}
}
