package services_test

import (
	"context"
	"fmt"
	"time"

	. "github.com/smartcontractkit/chainlink-common/pkg/internal/example"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
)

type configured struct {
	services.Service
	eng *services.Engine

	subA services.Service
	subB services.Service

	workCh chan func() (name string, err error)
}

func (c *configured) start(context.Context) error {
	c.eng.GoTick(services.NewTicker(time.Minute), c.do)
	return nil
}

func (c *configured) close() error {
	close(c.workCh)
	return nil
}

// do processes all outstanding work
func (c *configured) do(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case work, ok := <-c.workCh:
			if !ok {
				return
			}
			name, err := work()
			if err != nil {
				c.eng.SetHealthCond(name, err)
			} else {
				c.eng.ClearHealthCond(name)
			}
		default:
			return
		}
	}
}

func newFakeService(lggr logger.Logger, name string) services.Service {
	return services.Config{Name: name}.NewService(lggr)
}

func NewConfigured(lggr logger.Logger) services.Service {
	e := &configured{
		workCh: make(chan func() (string, error)),
	}
	e.Service, e.eng = services.Config{
		Name:  "Configured",
		Start: e.start,
		Close: e.close,
		NewSubServices: func(lggr logger.Logger) []services.Service {
			e.subA = newFakeService(lggr, "Sub-service-A")
			e.subB = newFakeService(lggr, "Sub-service-B")
			return []services.Service{e.subA, e.subB}
		},
	}.NewServiceEngine(lggr)
	return e
}

func ExampleConfig_NewService() {
	lggr, err := Logger()
	if err != nil {
		fmt.Println("Failed to create logger:", err)
		return
	}
	s := NewConfigured(lggr)
	if err = s.Start(context.Background()); err != nil {
		fmt.Println("Failed to start service:", err)
		return
	}
	if err = s.Close(); err != nil {
		fmt.Println("Failed to close service:", err)
		return
	}
	/* commented out because the log output is non-deterministic
	// Output:
	// INFO	Configured	Starting
	// INFO	Configured	Starting 2 sub-services
	// INFO	Configured.Sub-service-A	Starting
	// INFO	Configured.Sub-service-A	Started
	// INFO	Configured.Sub-service-B	Starting
	// INFO	Configured.Sub-service-B	Started
	// INFO	Configured	Started
	// INFO	Configured	Closing
	// INFO	Configured	Closing 2 sub-services
	// INFO	Configured.Sub-service-B	Closing
	// INFO	Configured.Sub-service-B	Closed
	// INFO	Configured.Sub-service-A	Closing
	// INFO	Configured.Sub-service-A	Closed
	// INFO	Configured	Closed
	*/
}
