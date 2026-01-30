package limits

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/contexts"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/settings"
)

func Test_updater(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		scope settings.Scope
		cre   contexts.CRE
	}{
		{settings.ScopeGlobal, contexts.CRE{}},
		{settings.ScopeOwner, contexts.CRE{Owner: "ow-id"}},
	} {
		t.Run(tt.scope.String(), func(t *testing.T) {
			t.Parallel()
			t.Run("static", func(t *testing.T) {
				t.Parallel()
				var got []int
				u := newUpdater[int](logger.Test(t), func(ctx context.Context) (int, error) { return 13, nil }, nil)
				u.recordLimit = func(ctx context.Context, i int) { got = append(got, i) }

				go u.updateLoop(contexts.WithCRE(t.Context(), tt.cre))
				time.Sleep(2 * pollPeriod)
				require.NoError(t, u.Close())

				assert.GreaterOrEqual(t, len(got), 1)
				for i := range got {
					assert.Equal(t, got[i], 13)
				}
			})
			t.Run("dynamic", func(t *testing.T) {
				t.Parallel()
				var limit atomic.Int64
				limit.Store(13)
				var got []int
				u := newUpdater[int](logger.Test(t), func(ctx context.Context) (int, error) { return int(limit.Load()), nil }, nil)
				u.recordLimit = func(ctx context.Context, i int) { got = append(got, i) }

				go u.updateLoop(contexts.WithCRE(t.Context(), tt.cre))
				time.Sleep(2 * pollPeriod)
				limit.Store(42)
				ctx2, cancel := context.WithCancel(contexts.WithCRE(t.Context(), contexts.CRE{Org: "org-id"}))
				t.Cleanup(cancel)
				u.updateCtx(ctx2)
				time.Sleep(2 * pollPeriod)
				require.NoError(t, u.Close())

				assert.GreaterOrEqual(t, len(got), 2)
				assert.Equal(t, got[0], 13)
				assert.Equal(t, got[len(got)-1], 42)

				cancel()
				select {
				case <-time.After(5 * time.Second):
					t.Error("timed out waiting for updater to close")
				case <-u.done:
					// success
				}
			})
			t.Run("sub", func(t *testing.T) {
				t.Parallel()
				updates := make(chan settings.Update[int])
				var got []int
				u := newUpdater[int](logger.Test(t), func(ctx context.Context) (int, error) { return 13, nil },
					func(ctx context.Context) (<-chan settings.Update[int], func()) { return updates, func() {} })
				u.recordLimit = func(ctx context.Context, i int) { got = append(got, i) }

				go u.updateLoop(contexts.WithCRE(t.Context(), tt.cre))
				updates <- settings.Update[int]{Value: 42}
				updates <- settings.Update[int]{Value: 100}
				ctx2, cancel := context.WithCancel(contexts.WithCRE(t.Context(), contexts.CRE{Org: "org-id"}))
				t.Cleanup(cancel)
				u.updateCtx(ctx2)
				require.NoError(t, u.Close())

				assert.GreaterOrEqual(t, len(got), 2)
				assert.Equal(t, got[0], 42)
				assert.Equal(t, got[len(got)-1], 100)

				cancel()
				select {
				case <-time.After(5 * time.Second):
					t.Error("timed out waiting for updater to close")
				case <-u.done:
					// success
				}
			})
		})
	}

}
