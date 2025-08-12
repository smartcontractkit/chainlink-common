package host

import (
	"context"
	"sync"
	"time"

	sdkpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
)

// timeFetcher safely retrieves DON or Node time from a background goroutine.
// It avoids calling into Go runtime APIs (e.g., context) inside Wasm trap handlers,
// which can panic if executed directly during WASI syscalls like clock_time_get.
// TODO: Add link reference if there is one and say "see this"
type timeFetcher struct {
	ctx             context.Context
	executor        ExecutionHelper
	timeRequestChan chan sdkpb.Mode
	timeResultChan  chan time.Time
	timeErrChan     chan error

	startOnce sync.Once
}

func newTimeFetcher(ctx context.Context, executor ExecutionHelper) *timeFetcher {
	return &timeFetcher{
		ctx:             ctx,
		executor:        executor,
		timeRequestChan: make(chan sdkpb.Mode, 1),
		timeResultChan:  make(chan time.Time, 1),
		timeErrChan:     make(chan error, 1),
	}
}

func (t *timeFetcher) GetTime(mode sdkpb.Mode) (time.Time, error) {
	select {
	case t.timeRequestChan <- mode:
	case <-t.ctx.Done():
		return time.Time{}, t.ctx.Err()
	}

	select {
	case donTime := <-t.timeResultChan:
		return donTime, nil
	case err := <-t.timeErrChan:
		return time.Time{}, err
	case <-t.ctx.Done():
		return time.Time{}, t.ctx.Err()
	}
}

func (t *timeFetcher) Start() {
	t.startOnce.Do(func() { go t.runLoop() })
}

func (t *timeFetcher) runLoop() {
	for {
		select {
		case <-t.ctx.Done():
			return
		case mode := <-t.timeRequestChan:
			var donTime time.Time
			var err error

			switch mode {
			case sdkpb.Mode_MODE_DON:
				donTime, err = t.executor.GetDONTime()
			default:
				donTime = t.executor.GetNodeTime()
			}

			if err != nil {
				select {
				case t.timeErrChan <- err:
				case <-t.ctx.Done():
					return
				}
			} else {
				select {
				case t.timeResultChan <- donTime:
				case <-t.ctx.Done():
					return
				}
			}
		}
	}
}
