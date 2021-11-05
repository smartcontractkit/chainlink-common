package services

import (
	"context"
	"math/big"

	"github.com/smartcontractkit/chainlink-relay/core/server/webhook"
	"github.com/smartcontractkit/chainlink/core/logger"
)

// DataSource struct for querying CL node job runs and waiting for response
type DataSource struct {
	id      string
	webhook *webhook.Trigger
	runData *chan *big.Int
	log     *logger.Logger
}

// Observe triggers a job run on CL node which is connected to EAs
// and return the job run result here as a "proxy" to the EAs
// this allows the CL node to track OCR rounds/job runs
func (ds *DataSource) Observe(ctx context.Context) (*big.Int, error) {
	ds.log.Infof("[%s] Observe triggered", ds.id)

	// send job trigger
	go ds.webhook.TriggerJob(ds.id)

	// wait for job run data to be returned
	data := <-*ds.runData
	ds.log.Infof("[%s] Observation received: %s", ds.id, data)

	return data, nil
}

// TODO: to be implemented (placeholder value)
type JuelsToEthDataSource struct {
}

func (ds *JuelsToEthDataSource) Observe(ctx context.Context) (*big.Int, error) {
	return big.NewInt(123456), nil
}
