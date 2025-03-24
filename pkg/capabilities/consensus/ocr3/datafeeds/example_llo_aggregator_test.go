package datafeeds_test

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	ocrcommon "github.com/smartcontractkit/libocr/commontypes"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/datafeeds"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/types"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/datastreams"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

// Example of using LLOAggregator.Aggregate with multiple oracles and price streams
func Example_lloAggregator_Aggregate() {
	// Create a logger
	lggr, err := logger.New()
	if err != nil {
		panic(err)
	}
	// 1. Create aggregator with 2 stream configs
	configMap, _ := values.NewMap(map[string]interface{}{
		"streams": map[string]interface{}{
			"1": map[string]interface{}{
				"deviation": "0.01", // 1% deviation threshold
				"heartbeat": 3600,   // 1 hour heartbeat
			},
			"2": map[string]interface{}{
				"deviation": "0.02", // 2% deviation threshold
				"heartbeat": 1800,   // 30 min heartbeat
			},
		},
		"allowedPartialStaleness": "0.2", // 20% partial staleness
	})

	aggregator, _ := datafeeds.NewLLOAggregator(*configMap)

	// 2. Create empty previous outcome (first round)
	var previousOutcome *types.AggregationOutcome = nil

	// 3. Create observations from 3 oracles
	observations := make(map[ocrcommon.OracleID][]values.Value)
	timestamp := uint64(time.Now().UnixNano())

	// Setup price data for 2 streams
	prices := map[uint32]decimal.Decimal{
		1: decimal.NewFromFloat(1250.75),  // ETH/USD price
		2: decimal.NewFromFloat(39250.25), // BTC/USD price
	}

	// Create the same observation for each oracle to ensure f+1 consensus
	for i := ocrcommon.OracleID(1); i <= 3; i++ {
		// Create LLO event with price payload
		event := &datastreams.LLOStreamsTriggerEvent{
			ObservationTimestampNanoseconds: timestamp,
			Payload:                         make([]*datastreams.LLOStreamDecimal, 0, len(prices)),
		}

		// Add each price to the payload
		for streamID, price := range prices {
			// Convert decimal to binary representation
			priceBinary, _ := price.MarshalBinary()

			event.Payload = append(event.Payload, &datastreams.LLOStreamDecimal{
				StreamID: streamID,
				Decimal:  priceBinary,
			})
		}

		// Wrap the event in a values.Value
		val, err := values.Wrap(event)
		if err != nil {
			panic(err)
		}
		observations[i] = []values.Value{val}
	}

	// 4. Call Aggregate with f=1
	outcome, err := aggregator.Aggregate(lggr, previousOutcome, observations, 1)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// 5. Print results
	fmt.Printf("Should report: %v\n", outcome.ShouldReport)

	// Decode the results to view updated streams
	if outcome.ShouldReport {
		streamIDs, reports, err := processOutcome(outcome)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Updated streams: %d\n", len(streamIDs))

		// Print details of each updated stream
		for i, report := range reports {

			fmt.Printf("  Stream %d: ID=%d, Price=%s, Timestamp=%d\n",
				i+1, report.StreamID, report.Price.String(), timestamp)
		}
	}
}

// Output:
// Should report: true
// Updated streams: 2
//   Stream 1: ID=1, Price=1250.75, Timestamp=1616744307328492000
//   Stream 2: ID=2, Price=39250.25, Timestamp=1616744307328492000
