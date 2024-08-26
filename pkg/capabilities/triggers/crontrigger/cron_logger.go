package crontrigger

import (
	"github.com/go-co-op/gocron/v2"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

var _ gocron.Logger = (*CronLogger)(nil)

type CronLogger struct {
	lggr logger.Logger
}

func NewCronLogger(lggr logger.Logger) gocron.Logger {
	return &CronLogger{lggr: lggr}
}

func (cl CronLogger) Debug(msg string, args ...any) {
	cl.lggr.Debugw(msg, "args", args)
}
func (cl CronLogger) Error(msg string, args ...any) {
	cl.lggr.Errorw(msg, "args", args)
}
func (cl CronLogger) Info(msg string, args ...any) {
	cl.lggr.Infow(msg, "args", args)
}
func (cl CronLogger) Warn(msg string, args ...any) {
	cl.lggr.Warnw(msg, "args", args)
}
