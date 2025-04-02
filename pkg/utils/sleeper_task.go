package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/services"
)

// Worker is a simple interface that represents some work to do repeatedly
type Worker interface {
	Work()
	Name() string
}

// WorkerCtx is like Worker but includes [context.Context].
type WorkerCtx interface {
	Work(ctx context.Context)
	Name() string
}

// SleeperTask represents a task that waits in the background to process some work.
type SleeperTask struct {
	services.StateMachine
	worker     WorkerCtx
	chQueue    chan struct{}
	chStop     services.StopChan
	chDone     chan struct{}
	chWorkDone chan struct{}
}

// NewSleeperTask takes a worker and returns a SleeperTask.
//
// SleeperTask is guaranteed to call Work on the worker at least once for every
// WakeUp call.
// If the Worker is busy when WakeUp is called, the Worker will be called again
// immediately after it is finished. For this reason you should take care to
// make sure that Worker is idempotent.
// WakeUp does not block.
func NewSleeperTask(w Worker) *SleeperTask {
	return NewSleeperTaskCtx(&worker{w})
}

type worker struct {
	Worker
}

func (w *worker) Work(ctx context.Context) { w.Worker.Work() }

// NewSleeperTaskCtx is like NewSleeperTask but accepts a WorkerCtx with a [context.Context].
func NewSleeperTaskCtx(w WorkerCtx) *SleeperTask {
	s := &SleeperTask{
		worker:     w,
		chQueue:    make(chan struct{}, 1),
		chStop:     make(chan struct{}),
		chDone:     make(chan struct{}),
		chWorkDone: make(chan struct{}, 10),
	}

	_ = s.StartOnce("SleeperTask-"+w.Name(), func() error {
		go s.workerLoop()
		return nil
	})

	return s
}

// Stop stops the SleeperTask
func (s *SleeperTask) Stop() error {
	return s.StopOnce("SleeperTask-"+s.worker.Name(), func() error {
		close(s.chStop)
		select {
		case <-s.chDone:
		case <-time.After(15 * time.Second):
			return fmt.Errorf("SleeperTask-%s took too long to stop", s.worker.Name())
		}
		return nil
	})
}

func (s *SleeperTask) WakeUpIfStarted() bool {
	return s.IfStarted(func() {
		select {
		case s.chQueue <- struct{}{}:
		default:
		}
	})
}

// WakeUp wakes up the sleeper task, asking it to execute its Worker.
func (s *SleeperTask) WakeUp() {
	if !s.IfStarted(func() {
		select {
		case s.chQueue <- struct{}{}:
		default:
		}
	}) {
		panic("cannot wake up stopped sleeper task")
	}
}

func (s *SleeperTask) workDone() {
	select {
	case s.chWorkDone <- struct{}{}:
	default:
	}
}

// WorkDone isn't part of the SleeperTask interface, but can be
// useful in tests to assert that the work has been done.
func (s *SleeperTask) WorkDone() <-chan struct{} {
	return s.chWorkDone
}

func (s *SleeperTask) workerLoop() {
	defer close(s.chDone)

	ctx, cancel := s.chStop.NewCtx()
	defer cancel()

	for {
		select {
		case <-s.chQueue:
			s.worker.Work(ctx)
			s.workDone()
		case <-s.chStop:
			return
		}
	}
}

type sleeperTaskWorker struct {
	name string
	work func()
}

// SleeperFuncTask returns a Worker to execute the given work function.
func SleeperFuncTask(work func(), name string) Worker {
	return &sleeperTaskWorker{name: name, work: work}
}

func (w *sleeperTaskWorker) Name() string { return w.name }
func (w *sleeperTaskWorker) Work()        { w.work() }
