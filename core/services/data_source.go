package services

import (
	"context"
	"errors"
	"math/big"

	"github.com/smartcontractkit/chainlink-relay/core/server/webhook"
	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/libocr/offchainreporting2/reportingplugin/median"
)

var (
	priceFeedParam = "PriceFeed"
	juelsToXParam  = "JuelsToX"
)

// DataSources struct for both data source interfaces
type DataSources struct {
	Price    median.DataSource
	JuelsToX median.DataSource
}

type dataSourceState struct {
	ID      string
	Webhook *webhook.Trigger
	RunData *chan *big.Int
	Log     *logger.Logger
	Prices  map[string]*big.Int
	Started bool
}

// NewDataSources creates the desired interface for the price feed and juelsToEth data source
// it maintains a single state where both prices are returned together, and attempst to make sure that the job run calls are not duplicated
// (especially if the observations are called in rapid succession)
func NewDataSources(id string, trigger *webhook.Trigger, runChannel *chan *big.Int, log *logger.Logger) DataSources {
	// initialize the state + various required connections
	dss := dataSourceState{
		ID:      id,
		Webhook: trigger,
		RunData: runChannel,
		Log:     log,
		Prices:  map[string]*big.Int{},
	}

	return DataSources{
		Price:    &dataSource{Key: priceFeedParam, dss: &dss},
		JuelsToX: &dataSource{Key: juelsToXParam, dss: &dss},
	}

}

// Observe triggers a job run on CL node which is connected to EAs
// and return the job run result here as a "proxy" to the EAs
// this allows the CL node to track OCR rounds/job runs
func (dss *dataSourceState) Observe(ctx context.Context) error {
	if dss.Started == true {
		return errors.New("Observe (job run) has already been triggered")
	}

	// set observation request started
	dss.Started = true
	dss.Log.Infof("[%s] Observe (job run) triggered", dss.ID)

	// send job trigger
	go dss.Webhook.TriggerJob(dss.ID)

	// wait for job run data to be returned
	dss.Prices[priceFeedParam] = <-*dss.RunData
	dss.Prices[juelsToXParam] = big.NewInt(0)
	dss.Log.Infof("[%s] Observation (job run) received: %+v", dss.ID, dss.Prices)

	// set observation request completed
	dss.Started = false

	return nil
}

type dataSource struct {
	Key string
	dss *dataSourceState
}

func (ds dataSource) Observe(ctx context.Context) (*big.Int, error) {
	ds.dss.Log.Infof("[%s - %s] Observe triggered", ds.dss.ID, ds.Key)
	// if no observation has been triggered, trigger new observation round
	if !ds.dss.Started {
		ds.dss.Log.Infof("[%s - %s] Triggering new observe job run", ds.dss.ID, ds.Key)
		if err := ds.dss.Observe(ctx); err != nil {
			return big.NewInt(0), err
		}
	} else {
		// if observation has already been triggered
		ds.dss.Log.Infof("[%s - %s] Waiting for observe job run to complete", ds.dss.ID, ds.Key)
		for ds.dss.Started == true {
			// wait for Done to be called then return prices
		}
	}

	// once observe is complete, return specific price data from state
	ds.dss.Log.Infof("[%s - %s] Observe complete: %s", ds.dss.ID, ds.Key, ds.dss.Prices[ds.Key])
	return ds.dss.Prices[ds.Key], nil
}
