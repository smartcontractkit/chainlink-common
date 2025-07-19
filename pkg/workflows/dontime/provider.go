package dontime

import (
	"context"

	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
)

// Provider wraps an existing OCR3 plugin provider (from the relayer)
// and overrides the plugin factory and contract transmitter for DonTime.
type Provider struct {
	Factory     ocr3types.ReportingPluginFactory[[]byte]
	Transmitter ocr3types.ContractTransmitter[struct{}]
}

var _ services.Service = &Provider{}

func (p *Provider) Start(_ context.Context) error {
	return nil
}

func (p *Provider) Close() error {
	return nil
}

func (p *Provider) Name() string {
	return "DonTimeOCR3Provider"
}

func (p *Provider) HealthReport() map[string]error {
	return map[string]error{p.Name(): nil}
}

func (p *Provider) Ready() error {
	return nil
}

func (p *Provider) ReportingPluginFactory() ocr3types.ReportingPluginFactory[[]byte] {
	return p.Factory
}

func (p *Provider) ContractTransmitter() ocr3types.ContractTransmitter[struct{}] {
	return p.Transmitter
}
