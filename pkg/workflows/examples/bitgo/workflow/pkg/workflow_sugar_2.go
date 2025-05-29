package pkg

import (
	_ "embed"
	"encoding/json"
	"log/slog"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/stubs/don/cron"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/stubs/node/http"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
)

func onCronTrigger(runtime sdk.DonRuntime, trigger *cron.CronTrigger) error {
	logger := slog.Default()
	config := &Config{}
	if err := json.Unmarshal(runtime.Config(), config); err != nil {
		logger.Error("error unmarshalling config", "err", err)
	}

	client := &http.Client{}

	// In TypeScript, we can call this fetch also, we'll need to decide what goes before or after fetch.
	// This version allows a single fetch, but the workflow author doesn't need to think about when to get consensus.
	request := &http.HttpFetchRequest{Url: config.Url}
	reserveInfo, err := client.ConsensusFetch(
		runtime,
		request,
		fetchPor,
		sdk.ConsensusAggregationFromTags[*ReserveInfo]()).
		Await()

	// The rest is the same
}

// Result's name and result.Value's name are TBD.
// Alternatively, we can pass another struct that gives config and secrets
func extractPor(result Result[*http.HttpFetchResponse]) (*ReserveInfo, error) {
	porResponse := &PorResponse{}
	if err := json.Unmarshal(result.Value.Body, porResponse); err != nil {
		return nil, err
	}

	// parsing, verify signature, and the rest is the same
}
