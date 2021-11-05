package services

import (
	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/libocr/commontypes"
)

type monitor struct{}

func (m monitor) SendLog(log []byte) {
	logger.Debugf("[Monitoring] %s", string(log))
}

func Monitoring() commontypes.MonitoringEndpoint {
	return monitor{}
}
