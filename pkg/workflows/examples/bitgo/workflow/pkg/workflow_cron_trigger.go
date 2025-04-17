package pkg

import (
	_ "embed"
	"encoding/json"
	"log/slog"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/stubs/don/cron"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
)

func Workflow(runner sdk.DonRunner) {
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
