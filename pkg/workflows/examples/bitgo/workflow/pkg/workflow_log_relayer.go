package pkg

import (
	_ "embed"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/stubs/don/cron"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/stubs/don/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
)

const (
	FillRequested = "FillRequested"
	FillCompleted = "FillCompleted"
)

var FastFillABI string
var erc20 abi.ABI

type LogRelayerConfig struct {
	// Note this could be the tokenAdminRegistry instead
	// for multi-token.
	SourceTokenPoolAddress string
	DestTokenPoolAddress   string
	LookbackMinutes        int
	Schedule               string
	// Specific pair of chains for this workflow,
	// again could be generalized.
	SourceChainSelector, DestChainSelector uint64
}

// FillRequestedEvent represents the structure of the FillRequested event data
type FillRequestedEvent struct {
	RequestID string `json:"requestId"`
	Amount    string `json:"amount"`
	Recipient string `json:"recipient"`
}

// FillCompletedEvent represents the structure of the FillCompleted event data
type FillCompletedEvent struct {
	RequestID string `json:"requestId"`
}

// LogRelayerWorkflow relays logs from source to destination
// for a given pair of chains and token pools.
// Relatively straightforward to extend to multi-chain/multi-token
// where token pools are read from the tokenAdminRegistry
// (a singleton contract per chain).
func LogRelayerWorkflow(runner sdk.DonRunner) {
	logger := slog.Default()
	config := &LogRelayerConfig{}
	if err := json.Unmarshal(runner.Config(), config); err != nil {
		logger.Error("error unmarshalling config", "err", err)
		return
	}

	// TODO: register logs
	evmClient := evm.Client{}
	evmClientDest := evm.Client{}
	erc20, _ := abi.JSON(strings.NewReader(FastFillABI))

	evmClient.RegisterLogTracking(nil /* add runner */, &evm.RegisterLogTrackingRequest{
		Filter: &evm.LPFilter{
			Addresses: []string{config.SourceTokenPoolAddress},
			EventSigs: []string{erc20.Events[FillRequested].Sig},
		},
	})
	evmClientDest.RegisterLogTracking(nil /* add runner */, &evm.RegisterLogTrackingRequest{
		Filter: &evm.LPFilter{
			Addresses: []string{config.DestTokenPoolAddress},
			EventSigs: []string{erc20.Events[FillCompleted].Sig},
		},
	})

	sdk.SubscribeToDonTrigger(
		runner,
		cron.Cron{}.Trigger(&cron.Config{Schedule: config.Schedule}),
		func(runtime sdk.DonRuntime, trigger *cron.CronTrigger) (struct{}, error) {
			return doBatchFill(runtime, trigger.ScheduledExecutionTime, config)
		})

	/*
		Optional enhancement to minimize unnecessary reads
		during a steady stead with no load, we could reduce
		the cron trigger to merely a fallback and instead
		trigger based on new log arrivals. To do batching
		with the log trigger, we can add an additional cron trigger
		which just flushes the batch after a short timeout.

		erc20, _ := abi.JSON(strings.NewReader(FastFillABI))
		fillRequest := erc20.Events[FillRequested]
		zeroAddr := common.Address{} // this is 0x000...000
		zeroAddressTopic := zeroAddr.Bytes()
		sdk.SubscribeToDonTrigger(
				runner,
				evmClient.LogTrigger(&evm.LogTriggerRequest{
					FilterQuery: &evm.FilterQuery{
						Address: []string{config.SourceTokenPoolAddress},
						Topics:  []string{fillRequest.ID.String(), string(zeroAddressTopic)},
					},
				}),
				func(runtime sdk.DonRuntime, log *evm.Log) (struct{}, error) {
					return doBatchFill(runtime, getBlockTime(log.BlockNumber), config)
				})
	*/
}

func blockByTime(client evm.Client, ts time.Time) uint64 {
	return 0
}

func doBatchFill(runtime sdk.DonRuntime, executionTime int64, config *LogRelayerConfig) (struct{}, error) {
	logger := slog.Default()
	// TODO use config chain selectors to load clients
	evmClientSource, evmClientDest := &evm.Client{}, &evm.Client{}

	// Calculate time window for log queries (last X minutes)
	fromTime := time.Unix(executionTime, 0).Add(-time.Minute * time.Duration(config.LookbackMinutes))

	// QueryLogs itself does consensus to ensure that
	// f+1 nodes have seen identical logs.
	// For example the individual nodes n1,n2,n3,n4 and f=1 see the following (rpcs delayed etc.):
	// n1=[1,2,3], n2=[1], n3=[1,2] n3=[2,3]
	// take all logs with at least 2 votes, we get [1,2,3]
	fillRequestedLogs, err := evmClientSource.QueryLogs(runtime, &evm.QueryLogsRequest{
		Filter: &evm.FilterQuery{
			FromBlock: &evm.FilterQuery_FromBlockNumber{
				FromBlockNumber: blockByTime(*evmClientSource, fromTime),
			},
			// Latest meaning no-confirmations
			ToBlock: &evm.FilterQuery_ToBlockTag{
				ToBlockTag: "latest",
			},
			Address: []string{config.SourceTokenPoolAddress},
			Topics:  []string{FillRequested},
		},
	}).Await()

	if err != nil {
		return struct{}{}, err
	}

	// Query FillCompleted logs from chain B
	fillCompletedLogs, err := evmClientDest.QueryLogs(runtime, &evm.QueryLogsRequest{
		Filter: &evm.FilterQuery{
			FromBlock: &evm.FilterQuery_FromBlockNumber{
				FromBlockNumber: uint64(fromTime.Unix()),
			},
			// If a fill re-org happens, we'll automatically retry
			// on the next cron tick.
			ToBlock: &evm.FilterQuery_ToBlockTag{
				ToBlockTag: "latest",
			},
			Address: []string{config.DestTokenPoolAddress},
			Topics:  []string{FillCompleted},
		},
	}).Await()

	if err != nil {
		return struct{}{}, err
	}

	// At this point, all nodes have the same set of completed/pending requests.

	// Create a map of completed request IDs for quick lookup
	completedRequests := make(map[string]bool)
	for _, log := range fillCompletedLogs.Logs {
		var completedEvent FillCompletedEvent
		if err := json.Unmarshal(log.Data, &completedEvent); err != nil {
			logger.Error("failed to parse FillCompleted event data", "err", err)
			continue
		}
		completedRequests[completedEvent.RequestID] = true
	}

	// Find FillRequested events that don't have corresponding FillCompleted events
	var pendingRequests []FillRequestedEvent
	for _, log := range fillRequestedLogs.Logs {
		var requestedEvent FillRequestedEvent
		if err := json.Unmarshal(log.Data, &requestedEvent); err != nil {
			logger.Error("failed to parse FillRequested event data", "err", err)
			continue
		}

		if !completedRequests[requestedEvent.RequestID] {
			pendingRequests = append(pendingRequests, requestedEvent)
		}
	}

	if len(pendingRequests) == 0 {
		logger.Info("No pending fill requests found")
		return struct{}{}, nil
	}

	// TODO: here we have the option to fine tune batching,
	//

	// Prepare batch transaction data
	erc20, err := abi.JSON(strings.NewReader(FastFillABI))
	if err != nil {
		return struct{}{}, err
	}

	// Pack the batch fill transaction data
	batchFillData, err := erc20.Pack("batchFill", pendingRequests)
	if err != nil {
		return struct{}{}, err
	}

	// Submit the batch transaction to chain B
	// TODO: update to writeReport.
	writeReportResultPromise := evmClientDest.WriteAsReport(runtime, &evm.WriteAsReportRequest{
		Receiver: config.DestTokenPoolAddress,
		Data:     batchFillData,
	})

	// Note this blocks until confirmed.
	// This prevents requests in quick succession from the log trigger
	// from creating duplicate fills.
	writeReportResult, err := writeReportResultPromise.Await()
	if err != nil {
		return struct{}{}, err
	}

	logger.Info("Submitted batch fill transaction",
		"txID", writeReportResult.TxHash,
		"numRequests", len(pendingRequests),
	)

	return struct{}{}, nil
}
