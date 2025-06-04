package pkg

import (
	_ "embed"
	"encoding/json"
	"errors"
	"log/slog"
	"math/big"
	"time"

	"github.com/shopspring/decimal"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/stubs/don/cron"
	evm "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/chain-capabilities/evm/chain-service"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/examples/bitgo/workflow/pkg/bindings"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
)

func WorkflowWithBindings(runner sdk.Runner) {
	logger := slog.Default()
	config := &Config{}
	if err := json.Unmarshal(runner.Config(), config); err != nil {
		logger.Error("error unmarshalling config", "err", err)
		return
	}

	runner.Run(&sdk.WorkflowArgs[sdk.DonRuntime]{
		Handlers: []sdk.Handler[sdk.DonRuntime]{
			sdk.NewEmptyDonHandler(
				cron.Cron{}.Trigger(&cron.Config{Schedule: config.Schedule}),
				onCronTrigger,
			),
		},
	})
}

func onCronTriggerWithBindings(runtime sdk.DonRuntime, trigger *cron.CronTrigger) error {
	logger := slog.Default()
	config := &Config{}
	if err := json.Unmarshal(runtime.Config(), config); err != nil {
		logger.Error("error unmarshalling config", "err", err)
	}

	reserveInfo, err := sdk.RunInNodeMode(
		runtime,
		fetchPor,
		sdk.ConsensusAggregationFromTags[*ReserveInfo]()).
		Await()

	if err != nil {
		return err
	}

	if time.UnixMilli(reserveInfo.LastUpdated).Before(time.Unix(trigger.ScheduledExecutionTime, 0).Add(-time.Hour * 24)) {
		logger.Warn("reserve time is too old", "time", reserveInfo.LastUpdated)
		return errors.New("reserved time is too old")
	}

	totalSupply := big.NewInt(0)

	token := bindings.NewIERC20(config.EvmChainSelector, hexToBytes(config.EvmTokenAddress), nil)
	reserveManager := bindings.NewIReserveManager(config.EvmChainSelector, hexToBytes(config.EvmPorAddress), evm.GasConfig{
		GasLimit: config.GasLimit,
	})

	evmTotalSupplyPromise := token.TotalSupplyAccessor().TotalSupply()
	evmSupply, err := evmTotalSupplyPromise.Await()
	if err != nil {
		// TODO specify which EVM
		logger.Error("Could not read from evm", "err", err.Error())
		return err
	}

	totalSupply = totalSupply.Add(totalSupply, evmSupply)
	// TODO add other chains

	totalReserveScaled := reserveInfo.TotalReserve.Mul(decimal.NewFromUint64(10e18)).BigInt()

	writeReportReplyPromise := reserveManager.UpdateReserveAccessor().WriteReport(bindings.UpdateReserveData{
		TotalMinted:  *totalSupply,
		TotalReserve: *totalReserveScaled,
	})

	writeReportReply, err := writeReportReplyPromise.Await()

	var writeErrors []error
	if err == nil {
		txHash := writeReportReply.TxHash
		logger.Debug("Submitted transaction", "tx hash", txHash)
	} else {
		logger.Error("failed to submit transaction", "err", err)
		writeErrors = append(writeErrors, err)
	}

	return errors.Join(writeErrors...)
}
