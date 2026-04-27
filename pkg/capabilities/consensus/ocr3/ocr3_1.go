package ocr3

import (
	"context"
	"errors"
	"time"

	"github.com/jonboulle/clockwork"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/types"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/requests"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/reportingplugins"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

// CapabilityOCR3_1 is the OCR3_1 entry point, parallel to Capability in
// ocr3.go. It is intentionally its own type so the OCR3 path remains
// untouched during the staged rollout (plan §3.8).
//
// Unlike the OCR3 Capability, this one does not implement the LOOP
// ProviderServer interface in v1 — following the Vault precedent where the
// OCR3_1 plugin is instantiated directly in-process rather than over LOOP's
// gRPC boundary. Adding a LOOP sibling is a separate follow-up (plan §3.12).
type CapabilityOCR3_1 struct {
	loop.Plugin
	reportingplugins.PluginProviderServer
	config             Config
	capabilityRegistry core.CapabilitiesRegistry
}

// NewOCR3_1 constructs the OCR3_1 capability using the same Config shape as
// NewOCR3. Defaults mirror the OCR3 path so migration does not require
// reconfiguring the caller-supplied fields.
func NewOCR3_1(config Config) *CapabilityOCR3_1 {
	if config.RequestTimeout == nil {
		dre := defaultRequestExpiry
		config.RequestTimeout = &dre
	}
	if config.SendBufferSize == 0 {
		config.SendBufferSize = defaultSendBufferSize
	}
	if config.clock == nil {
		config.clock = clockwork.NewRealClock()
	}
	if config.store == nil {
		config.store = requests.NewStore[*ReportRequest]()
	}
	if config.capability == nil {
		ci := NewCapability(
			config.store,
			config.clock,
			*config.RequestTimeout,
			config.AggregatorFactory,
			config.EncoderFactory,
			config.Logger,
			config.SendBufferSize,
		)
		config.capability = ci
	}
	cp := &CapabilityOCR3_1{
		Plugin:               loop.Plugin{Logger: config.Logger},
		PluginProviderServer: reportingplugins.PluginProviderServer{},
		config:               config,
	}
	cp.SubService(config.capability)
	return cp
}

// NewReportingPluginFactoryOCR3_1 returns the OCR3_1 factory directly
// (*factoryOCR3_1 implements ocr3_1types.ReportingPluginFactory[[]byte]).
//
// Callers that drive libocr's OCR3_1 oracle harness should use this entry
// point. The integration-test framework in chainlink wires through here.
func (o *CapabilityOCR3_1) NewReportingPluginFactoryOCR3_1(
	ctx context.Context,
	_ core.ReportingPluginServiceConfig,
	capabilityRegistry core.CapabilitiesRegistry,
) (*factoryOCR3_1, error) {
	f, err := newFactoryOCR3_1(o.config.store, o.config.capability, o.config.Logger)
	if err != nil {
		return nil, err
	}
	if err := capabilityRegistry.Add(ctx, o.config.capability); err != nil {
		return nil, err
	}
	o.capabilityRegistry = capabilityRegistry
	return f, nil
}

// NewValidationServiceOCR3_1 mirrors the OCR3 validation-service entry.
// No behavioral difference — validation is over offchain config bytes which
// share a schema across OCR3 / OCR3_1 (with the new OCR3_1 fields additive).
func (o *CapabilityOCR3_1) NewValidationServiceOCR3_1(ctx context.Context) (core.ValidationService, error) {
	s := &validationService{lggr: o.Logger}
	o.SubService(s)
	return s, nil
}

func (o *CapabilityOCR3_1) Close() error {
	err := o.Plugin.Close()
	if o.capabilityRegistry != nil {
		err = errors.Join(err, o.capabilityRegistry.Remove(context.TODO(), o.config.capability.ID))
	}
	return err
}

// ensure unused imports are retained against future additions
var (
	_ = time.Second
	_ *types.ReportingPluginConfig
	_ = logger.Nop
)
