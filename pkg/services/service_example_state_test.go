package services_test

import (
	"context"
	"fmt"
	"sync"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services"

	. "github.com/smartcontractkit/chainlink-common/pkg/internal/example"
)

type stateMachine struct {
	services.StateMachine

	lggr logger.Logger

	stop services.StopChan
	wg   sync.WaitGroup
}

func (f *stateMachine) HealthReport() map[string]error {
	return map[string]error{f.Name(): f.Healthy()}
}

func (f *stateMachine) Name() string { return f.lggr.Name() }

func NewStateMachine(lggr logger.Logger) services.Service {
	return &stateMachine{
		lggr: logger.Named(lggr, "StateMachine"),
		stop: make(services.StopChan),
	}
}

func (f *stateMachine) Start(ctx context.Context) error {
	return f.StartOnce("StateMachine", func() error {
		f.lggr.Info("Starting")
		f.wg.Add(1)
		go f.run()
		return nil
	})
}

func (f *stateMachine) Close() error {
	return f.StopOnce("StateMachine", func() error {
		f.lggr.Info("Closing")
		close(f.stop) // trigger goroutine cleanup
		f.wg.Wait()   // wait for cleanup to complete
		return nil
	})
}

func (f *stateMachine) run() {
	defer f.wg.Done()

	for {
		select {
		// ...
		case <-f.stop:
			return // stop the routine
		}
	}

}

func ExampleService() {
	lggr, err := Logger()
	if err != nil {
		fmt.Println("Failed to create logger:", err)
		return
	}
	s := NewStateMachine(lggr)
	if err = s.Start(context.Background()); err != nil {
		fmt.Println("Failed to start service:", err)
		return
	}
	if err = s.Close(); err != nil {
		fmt.Println("Failed to close service:", err)
		return
	}

	// Output:
	// INFO	StateMachine	Starting
	// INFO	StateMachine	Closing
}
