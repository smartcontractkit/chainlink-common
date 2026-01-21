// Package balance provides a generic chain-agnostic balance monitoring service
// that tracks account balances across different blockchain networks.
package balance

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
)

// GaugeAccBalance defines a new gauge metric for account balance
type GaugeAccBalance struct {
	// account_balance
	gauge metric.Float64Gauge
}

func NewGaugeAccBalance(unitStr string) (*GaugeAccBalance, error) {
	name := "account_balance"
	description := "Balance for configured WT account"
	gauge, err := beholder.GetMeter().Float64Gauge(name, metric.WithUnit(unitStr), metric.WithDescription(description))
	if err != nil {
		return nil, fmt.Errorf("failed to create new gauge %s: %+w", name, err)
	}
	return &GaugeAccBalance{gauge}, nil
}

func (g *GaugeAccBalance) Record(ctx context.Context, balance float64, account string, chainInfo ChainInfo) {
	oAttrs := metric.WithAttributeSet(g.GetAttributes(account, chainInfo))
	g.gauge.Record(ctx, balance, oAttrs)

	// TODO: consider also recording record in Prom for availability to NOPs
}

func (g *GaugeAccBalance) GetAttributes(account string, chainInfo ChainInfo) attribute.Set {
	return attribute.NewSet(
		attribute.String("account", account),

		// Execution Context - Source
		attribute.String("source_id", ValOrUnknown(account)), // reusing account as source_id
		// Execution Context - Chain
		attribute.String("chain_family_name", ValOrUnknown(chainInfo.ChainFamilyName)),
		attribute.String("chain_id", ValOrUnknown(chainInfo.ChainID)),
		attribute.String("network_name", ValOrUnknown(chainInfo.NetworkName)),
		attribute.String("network_name_full", ValOrUnknown(chainInfo.NetworkNameFull)),
	)
}
