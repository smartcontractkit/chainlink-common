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
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/shopspring/decimal"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/stubs/don/cron"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/stubs/don/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/stubs/node/http"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
)

func onCronTrigger(runtime sdk.DonRuntime, trigger *cron.CronTrigger) error {
	logger := slog.Default()
	config := &Config{}
	if err := json.Unmarshal(runtime.Config(), config); err != nil {
		logger.Error("error unmarshalling config", "err", err)
	}

	triggerTime := time.Unix(trigger.ScheduledExecutionTime, 0)
	reserveInfo, err := sdk.RunInNodeMode(
		runtime,
		func (nrt sdk.NodeRuntime) (*ReserveInfo, error) {
			return fetchPor(nrt, triggerTime)
		} ,
		sdk.ConsensusAggregationFromTags[*ReserveInfo]()).
		Await()

	if err != nil {
		return err
	}

	if time.UnixMilli(reserveInfo.LastUpdated).Before(triggerTime).Add(-time.Hour * 24)) {
		logger.Warn("reserve time is too old", "time", reserveInfo.LastUpdated)
		return errors.New("reserved time is too old")
	}

	// The rest is the same
}

func fetchPor(runtime sdk.NodeRuntime, triggerTime time.Time) (*ReserveInfo, error) {
	config := &Config{}
	if err := json.Unmarshal(runtime.Config(), config); err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %w", err)
	}

	nodeCache := kvstore.NodeStore{/* There's more to this, we'll need an affinity group, but that's not fully designed yet.*/ }
	latest, err := nodeCache.Get("PoR").Await()
	// Exact interface for the KV store is TBD.
	// Essentially, if we've serviced a request within the last 5 minutes, we can return the cached value.
	// This can't use the sugar because it has multiple different node calls to make.
	if err != nil && latest.Exists() {
		reserveInfo := &ReserveInfo{}
		if err = json.Unmarshal(latest.Value, reserveInfo); err == nil {
			if time.Unix(reserveInfo.LastUpdated, 0).Before(triggerTime.Add(-time.Minute*5)) {
				return reserveInfo, nil
			}
		}
	}

	request := &http.HttpFetchRequest{Url: config.Url}
	client := &http.Client{}
	response, err := client.Fetch(runtime, request).Await()
	if err != nil {
		return nil, err
	}

	// The rest is the same as before
}
