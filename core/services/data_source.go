package services

import (
	"context"
	"errors"
	"math/big"

	"github.com/smartcontractkit/chainlink-relay/core/server/webhook"
	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/libocr/offchainreporting2/reportingplugin/median"
	"go.uber.org/atomic"
)

var (
	priceFeedParam      = "PriceFeed"
	juelsToXParam       = "JuelsToX"
	errAlreadyTriggered = errors.New("Observe (job run) has already been triggered")
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
	locked  *atomic.Bool
	Done    chan struct{}
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
		locked:  atomic.NewBool(false), // initialize to unlocked
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
	// lock
	if !dss.locked.CAS(false, true) {
		return errAlreadyTriggered
	}
	defer dss.locked.Store(false)

	// set observation request started
	dss.Log.Infof("[%s] Observe (job run) triggered", dss.ID)
	dss.Done = make(chan struct{}) // create channel for signaling finish

	// send job trigger
	go dss.Webhook.TriggerJob(dss.ID)

	// wait for job run data to be returned
	dss.Prices[priceFeedParam] = <-*dss.RunData // use first value
	dss.Prices[juelsToXParam] = <-*dss.RunData  // use second value
	dss.Log.Infof("[%s] Observation (job run) received: %+v", dss.ID, dss.Prices)

	close(dss.Done) // close channel to indicate done
	return nil
}

type dataSource struct {
	Key string
	dss *dataSourceState
}

func (ds dataSource) Observe(ctx context.Context) (*big.Int, error) {
	ds.dss.Log.Infof("[%s - %s] Observe triggered", ds.dss.ID, ds.Key)

	// try triggering observe
	err := ds.dss.Observe(ctx)

	switch err {
	case nil: // occurs when new observation is triggered successfully
		// do nothing
	case errAlreadyTriggered: // occrus when observe has already been triggered
		ds.dss.Log.Infof("[%s - %s] Waiting for observe job run to complete", ds.dss.ID, ds.Key)
		// once observe is complete (channel closure)
		<-ds.dss.Done
	default: // return if unknown error
		return big.NewInt(0), err
	}

	// return specific price data from state
	ds.dss.Log.Infof("[%s - %s] Observe complete: %s", ds.dss.ID, ds.Key, ds.dss.Prices[ds.Key])
	return ds.dss.Prices[ds.Key], nil
}
