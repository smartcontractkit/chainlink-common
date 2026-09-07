package grafana

import "golang.org/x/sync/errgroup"

// parallelFor runs fn for each item with at most limit calls in flight and
// returns the first non-nil error returned by any invocation. In-flight calls are allowed to finish;
// items whose fn has not started yet may still run after an error occurs, which
// matches the pre-existing partial-apply behavior of the serial loops (a
// failure mid-loop leaves earlier items applied).
//
// Callers must ensure fn is safe for concurrent use: items must be independent
// (e.g. alert rules addressed by distinct UIDs).
func parallelFor[T any](items []T, limit int, fn func(T) error) error {
	if limit < 1 {
		limit = 1
	}

	var g errgroup.Group
	g.SetLimit(limit)

	for _, item := range items {
		i := item
		g.Go(func() error {
			return fn(i)
		})
	}

	return g.Wait()
}
