package pkg

import (
	_ "embed"
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/stubs/don/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
)

const TransferEvent = "Transfer"

func WorkflowLogTrigger(runner sdk.DonRunner) {
	logger := slog.Default()
	config := &Config{}
	var evmClient = evm.Client{}
	if err := json.Unmarshal(runner.Config(), config); err != nil {
		logger.Error("error unmarshalling config: %v", err)
		return
	}
	erc20, _ := abi.JSON(strings.NewReader(Erc20Abi))
	transferEvent := erc20.Events[TransferEvent]
	zeroAddr := common.Address{} // this is 0x000...000
	zeroAddressTopic := zeroAddr.Bytes()

	sdk.SubscribeToDonTrigger(
		runner,
		evmClient.LogTrigger(&evm.LogTriggerRequest{
			FilterQuery: &evm.FilterQuery{
				Address: []string{config.EvmTokenAddress},
				Topics:  []string{transferEvent.ID.String(), string(zeroAddressTopic)},
			},
		}),
		func(runtime sdk.DonRuntime, log *evm.Log) (struct{}, error) {
			return onTrigger(runtime, getBlockTime(log.BlockNumber), config)
		})
}

func getBlockTime(blockNumber uint64) int64 {
	//TODO: This would require getting the Block data and extract it's timestamp.
	panic("unimplemented")
}
