package poller

import (
	"github.com/smartcontractkit/chainlink-common/pkg/utils/mailbox"
	"sync"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
)

// Poller is a component that polls a function at a given interval
// and delivers the result to subscribers
type Poller[T any] struct {
	services.StateMachine
	pollingInterval time.Duration
	pollingFunc     func() (T, error)
	logger          logger.Logger

	subscribers []*mailbox.Mailbox[T]
	stopCh      chan struct{}
	wg          sync.WaitGroup
}

// NewPoller creates a new Poller instance
func NewPoller[T any](pollingInterval time.Duration, pollingFunc func() (T, error), logger logger.Logger) *Poller[T] {
	return &Poller[T]{
		pollingInterval: pollingInterval,
		pollingFunc:     pollingFunc,
		logger:          logger,
		subscribers:     make([]*mailbox.Mailbox[T], 0),
		stopCh:          make(chan struct{}),
	}
}

// Subscribe adds a new subscriber to the Poller
func (p *Poller[T]) Subscribe(subscriber *mailbox.Mailbox[T]) {
	p.subscribers = append(p.subscribers, subscriber)
}

func (p *Poller[T]) Unsubscribe(subscriber *mailbox.Mailbox[T]) {
	if subscriber == nil {
		return
	}
	for i, sub := range p.subscribers {
		if sub == subscriber {
			p.subscribers = append(p.subscribers[:i], p.subscribers[i+1:]...)
			return
		}
	}
}

// Start starts the polling process
func (p *Poller[T]) Start() error {
	return p.StartOnce("Poller", func() error {
		p.wg.Add(1)
		go p.pollingLoop()
		return nil
	})
}

func (p *Poller[T]) pollingLoop() {
	ticker := time.NewTicker(p.pollingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			result, err := p.pollingFunc()
			if err != nil {
				p.logger.Error("Error polling:", err)
				continue
			}
			for _, sub := range p.subscribers {
				sub.Deliver(result)
			}
		case <-p.stopCh:
			p.wg.Done()
			return
		}
	}
}

// Stop stops the polling process and waits for the polling goroutine to stop
func (p *Poller[T]) Stop() error {
	return p.StopOnce("Poller", func() error {
		close(p.stopCh)
		p.wg.Wait()
		return nil
	})
}
