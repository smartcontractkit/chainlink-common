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

type LogRelayerConfig struct {
	SourceTokenPoolAddress string
	DestTokenPoolAddress   string
	LookbackMinutes        int
	Schedule               string
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

// BuildWorkflow creates a workflow that listens for FillRequested events
func LogRelayerWorkflow(runner sdk.DonRunner) {
	logger := slog.Default()
	config := &Config{}
	if err := json.Unmarshal(runner.Config(), config); err != nil {
		logger.Error("error unmarshalling config", "err", err)
		return
	}

	sdk.SubscribeToDonTrigger(
		runner,
		cron.Cron{}.Trigger(&cron.Config{Schedule: config.Schedule}),
		func(runtime sdk.DonRuntime, trigger *cron.CronTrigger) (struct{}, error) {
			return onTrigger(runtime, trigger.ScheduledExecutionTime, config)
		})
}

func blockByTime(client evm.Client, ts time.Time) uint64 {
	return 0
}

func onFillBatchTick(runtime sdk.DonRuntime, executionTime int64, config *LogRelayerConfig) (struct{}, error) {
	logger := slog.Default()
	evmClientSource, evmClientDest := &evm.Client{}, &evm.Client{}

	// Calculate time window for log queries (last X minutes)
	fromTime := time.Unix(executionTime, 0).Add(-time.Minute * time.Duration(config.LookbackMinutes))

	// Query FillRequested logs from chain A
	fillRequestedLogs, err := evmClientSource.QueryLogs(runtime, &evm.QueryLogsRequest{
		Filter: &evm.FilterQuery{
			FromBlock: &evm.FilterQuery_FromBlockNumber{
				FromBlockNumber: blockByTime(*evmClientSource, fromTime),
			},
			ToBlock: &evm.FilterQuery_ToBlockTag{
				ToBlockTag: "latest",
			},
			Address: []string{config.SourceTokenPoolAddress},
			Topics:  []string{FillRequested},
		},
		// Low finality
		ConfidenceLevel: evm.ConfidenceLevel_LOW,
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
			ToBlock: &evm.FilterQuery_ToBlockTag{
				ToBlockTag: "latest",
			},
			Address: []string{config.DestTokenPoolAddress},
			Topics:  []string{FillCompleted},
		},
		// If a fill re-org happens, we'll automatically retry.
		ConfidenceLevel: evm.ConfidenceLevel_LOW,
	}).Await()

	if err != nil {
		return struct{}{}, err
	}

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
	tx := evmClientDest.SubmitTransaction(runtime, &evm.SubmitTransactionRequest{
		ToAddress: config.DestTokenPoolAddress,
		Calldata:  batchFillData,
	})

	txID, err := tx.Await()
	if err != nil {
		return struct{}{}, err
	}

	logger.Info("Submitted batch fill transaction",
		"txID", txID,
		"numRequests", len(pendingRequests),
	)

	return struct{}{}, nil
}
