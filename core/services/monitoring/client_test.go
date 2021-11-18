package monitoring

import (
	"testing"
	"time"

	corelogger "github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/chainlink/core/services/ocrcommon"
	"github.com/stretchr/testify/mock"
)

func TestClient(t *testing.T) {
	defaultLogger := corelogger.Default.Named("OCR2")
	logger := ocrcommon.NewLogger(defaultLogger, true, func(string) {})

	t.Run("published telemetry gets sent to the backend", func(t *testing.T) {
		mockOTIConn := new(MockWSRPCConnection)
		mockOTIConn.
			On("Invoke", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(nil)
		mockOTIConn.On("Close").Once()

		client := NewClient(mockOTIConn, "oracle#1", 100, logger)
		go client.Start()
		client.SendLog([]byte{'1', '2', '3'})
		<-time.After(10 * time.Millisecond)
		client.Close()
		mockOTIConn.AssertExpectations(t)
	})
}
