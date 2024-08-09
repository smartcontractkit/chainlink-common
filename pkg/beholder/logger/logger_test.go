package logger_test

import (
	"errors"
	"io/fs"
	"testing"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder/logger"
)

func TestLoggerExample(t *testing.T) {
	log := logger.New()
	log.Info("This is an info log")
	log.Error("This is an error log")
	log.Debug("This is a debug log")
	log.Warn("This is a warning log")
	log.Infof("This is a formatted info log with %s", "arguments")
	log.Warnf("This is a formatted warning log with %s", "arguments")
	log.Errorf("This is a formatted error log with %s", "arguments")
	log.Debugf("This is a formatted debug log with %s", "arguments")
	log.Infow("This is a structured info log", "key", "value")
	log.Warnw("This is a structured warning log", "key", "value")
	log.Errorw("This is a structured error log", "key", "value")
	log.Debugw("This is a structured debug log", "key", "value")
	var pathErr *fs.PathError
	if err := log.Sync(); err != nil && !errors.As(err, &pathErr) {
		t.Fatal(err)
	}
}
