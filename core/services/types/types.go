package types

import (
	"github.com/smartcontractkit/libocr/offchainreporting2/reportingplugin/median"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
)

type ContractTracker interface {
	Start()
	Close() error
	types.ContractTransmitter
	types.ContractConfigTracker
	median.ReportCodec
	median.MedianContract
	types.OffchainConfigDigester
}

type Blockchain interface {
	NewContractTracker(address, jobID string) (ContractTracker, error)
	OCR() types.OnchainKeyring
}

type OCRJobRunData struct {
	Result   string `json:"result"`
	JuelsToX string `json:"juelsToX"`
}
