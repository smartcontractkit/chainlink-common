package grafana

import "sync"

// parallelFor runs fn for each item with at most limit calls in flight and
// returns the first error encountered. In-flight calls are allowed to finish;
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

	sem := make(chan struct{}, limit)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error

	for _, item := range items {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			if err := fn(item); err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	return firstErr
}
