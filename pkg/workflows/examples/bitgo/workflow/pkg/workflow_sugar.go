package pkg

import (
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
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
	// This version allows multiple fetches, but forces the workflow author to think about which one(s) to get consensus on.
	reserveInfo, err := client.ConsensusFetch(
		runtime,
		fetchPor,
		sdk.ConsensusAggregationFromTags[*ReserveInfo]()).
		Await()

	// The rest is the same
}

// Alternatively, instead of the fetcher having config, we can pass another object that gives config and secrets
func fetchPor(fetcher http.Fetcher) (*ReserveInfo, error) {
	config := &Config{}
	if err := json.Unmarshal(fetcher.Config(), config); err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %w", err)
	}

	request := &http.HttpFetchRequest{Url: config.Url}

	// This line is different, you don't have a runtime.
	response, err := fetcher.Fetch(request).Await()
	if err != nil {
		return nil, err
	}

	// parsing, verify signature, and the rest is the same
}
