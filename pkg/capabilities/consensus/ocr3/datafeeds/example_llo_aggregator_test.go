package datafeeds_test

import (
	"fmt"

	"github.com/shopspring/decimal"
	ocrcommon "github.com/smartcontractkit/libocr/commontypes"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/datafeeds"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/types"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/datastreams"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

// Example of using LLOAggregator.Aggregate with multiple oracles and price streams
// It constructs a LLOAggregator with two streams, simulates observations from three oracles,
// and demonstrates how to process the aggregation outcome.
// go test -run ExampleLLOAggregator_Aggregate
func ExampleLLOAggregator_Aggregate() {
	// Create a logger
	lggr, err := logger.New()
	if err != nil {
		panic(err)
	}
	// 1. Create aggregator with 2 stream configs
	configMap, _ := values.NewMap(map[string]interface{}{
		"streams": map[string]interface{}{
			"1": map[string]interface{}{
				"deviation":  "0.01", // 1% deviation threshold
				"heartbeat":  3600,   // 1 hour heartbeat
				"remappedID": "0x680084f7347baFfb5C323c2982dfC90e04F9F918",
			},
			"2": map[string]interface{}{
				"deviation":  "0.02", // 2% deviation threshold
				"heartbeat":  1800,   // 30 min heartbeat
				"remappedID": "0x00001237347baFfb5C323c1112dfC90e0789FFFF",
			},
		},
		"allowedPartialStaleness": "0.2", // 20% partial staleness
	})

	aggregator, err := datafeeds.NewLLOAggregator(*configMap)
	if err != nil {
		panic(err)
	}

	// 2. Create empty previous outcome (first round); empty previousOutcome will cause all streams to be updated
	var previousOutcome *types.AggregationOutcome

	// 3. Create observations from 3 oracles
	observations := make(map[ocrcommon.OracleID][]values.Value)
	timestamp := uint64(61116379204) //uint64(time.Now().UnixNano()) //nolint: gosec // G115

	// Setup price data for 2 streams
	prices := map[uint32]decimal.Decimal{
		1: decimal.NewFromFloat(1250.427975), // ETH/USD price
		2: decimal.NewFromFloat(39250.25),    // BTC/USD price
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
		val, err2 := values.Wrap(event)
		if err2 != nil {
			panic(err2)
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
			fmt.Printf("  Stream %d: ID=%d, Price=%s, Timestamp=%d, RemappedID=%x\n",
				i+1, report.StreamID, report.Price.String(), timestamp, report.RemappedID)
		}
	}

	// Output:
	// Should report: true
	// Updated streams: 2
	//   Stream 1: ID=1, Price=1250.427975, Timestamp=61116379204 RemappedID=680084f7347baFfb5C323c2982dfC90e04F9F918
	//   Stream 2: ID=2, Price=39250.25, Timestamp=61116379204 RemappedID=00001237347baFfb5C323c1112dfC90e0789FFFF
}
