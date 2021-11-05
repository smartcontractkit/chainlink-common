package services

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/smartcontractkit/chainlink-relay/core/server/webhook"
	"github.com/smartcontractkit/chainlink-relay/core/services/types"
	"github.com/smartcontractkit/chainlink-relay/core/store/models"
	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/chainlink/core/utils"
)

type mockService struct {
	job       models.Job
	heartbeat time.Duration
	close     chan struct{}
	webhook   webhook.Trigger
	runData   chan *big.Int
	contract  mockTracker
	log       *logger.Logger
}

type mockTracker interface {
	Start()
	Close() error
}

// NewMockService creates a service that simulates the passing of messages between external client and CL node
func NewMockService(job models.Job, trigger webhook.Trigger, blockchain types.Blockchain) (mockService, error) {

	// create contract tracker
	contract, err := blockchain.NewContractTracker(job.ContractAddress, job.JobID)
	if err != nil {
		return mockService{}, err
	}

	heartbeat, _ := time.ParseDuration("15s")
	return mockService{
		job:       job,
		heartbeat: heartbeat,
		close:     make(chan struct{}),
		webhook:   trigger,
		runData:   make(chan *big.Int),
		contract:  contract,
		log:       logger.Default.Named("mock-service"),
	}, nil
}

// Start starts the service
func (ms *mockService) Start() {
	ms.log.Infof("[%s] Start mock service", ms.job.JobID)
	timer := utils.NewResettableTimer()
	timer.Reset(ms.heartbeat)
	defer timer.Stop()

	// start ws subscription to address
	go ms.contract.Start()

	// use dataSource
	ds := DataSource{
		id:      ms.job.JobID,
		webhook: &ms.webhook,
		runData: &ms.runData,
		log:     ms.log,
	}

	for {
		select {
		case <-timer.Ticks():
			// TODO: Implement a timeout context
			if _, err := ds.Observe(context.TODO()); err != nil {
				ms.log.Error(err)
			}
			timer.Reset(ms.heartbeat)
		case <-ms.close:
			return
		}
	}
}

// Run is a wrapper to pass data back to the original thread and unblock the function
func (ms *mockService) Run(data string) error {
	var val big.Int
	if _, ok := val.SetString(data, 10); !ok {
		return fmt.Errorf("[%s] Failed to parse *big.Int from %s", ms.job.JobID, data)
	}

	ms.runData <- &val // pass data back to job run
	return nil
}

// Stop ends the service
func (ms *mockService) Stop() error {
	ms.log.Infof("[%s] Stop mock service", ms.job.JobID)

	// unsubscribe ws connection
	if err := ms.contract.Close(); err != nil {
		return err
	}

	// trigger loop to quit
	close(ms.close)

	// close run channel for returning data source
	close(ms.runData)
	return nil
}
