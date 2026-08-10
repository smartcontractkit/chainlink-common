// Package balance provides a generic chain-agnostic balance monitoring service
// that tracks account balances across different blockchain networks.
package balance

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
)

// nodeBalance state for the NOP-facing Prometheus mirror.
var (
	nodeBalanceOnce sync.Once
	nodeBalanceVec  *prometheus.GaugeVec
	nodeBalanceErr  error
)

func nodeBalanceGauge() (*prometheus.GaugeVec, error) {
	nodeBalanceOnce.Do(func() {
		g := prometheus.NewGaugeVec(
			prometheus.GaugeOpts{Name: "node_balance", Help: "Account balances"},
			[]string{"account", "chainID", "chainFamily"},
		)
		err := prometheus.DefaultRegisterer.Register(g)
		if err == nil {
			nodeBalanceVec = g
			return
		}
		if already, ok := errors.AsType[prometheus.AlreadyRegisteredError](err); ok {
			if existing, ok := already.ExistingCollector.(*prometheus.GaugeVec); ok {
				nodeBalanceVec = existing
				return
			}
		}
		nodeBalanceErr = fmt.Errorf("failed to register node_balance gauge: %w", err)
	})
	return nodeBalanceVec, nodeBalanceErr
}

// GaugeAccBalance defines a new gauge metric for account balance
type GaugeAccBalance struct {
	// account_balance
	gauge metric.Float64Gauge
	// node_balance Prometheus mirror for NOPs
	promMirror *prometheus.GaugeVec
}

func NewGaugeAccBalance(unitStr string) (*GaugeAccBalance, error) {
	name := "account_balance"
	description := "Balance for configured WT account"
	gauge, err := beholder.GetMeter().Float64Gauge(name, metric.WithUnit(unitStr), metric.WithDescription(description))
	if err != nil {
		return nil, fmt.Errorf("failed to create new gauge %s: %+w", name, err)
	}
	promMirror, err := nodeBalanceGauge()
	if err != nil {
		return nil, err
	}
	return &GaugeAccBalance{gauge: gauge, promMirror: promMirror}, nil
}

func (g *GaugeAccBalance) Record(ctx context.Context, balance float64, account string, chainInfo ChainInfo) {
	oAttrs := metric.WithAttributeSet(g.GetAttributes(account, chainInfo))
	g.gauge.Record(ctx, balance, oAttrs)

	// Also record in Prom for availability to NOPs: node_balance is the
	// cross-chain standard gauge exposed on the node's /metrics endpoint
	// (the Beholder gauge above only reaches the internal telemetry pipeline).
	g.promMirror.
		WithLabelValues(account, chainInfo.ChainID, chainInfo.ChainFamilyName).
		Set(balance)
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
