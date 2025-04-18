package capabilities

import (
	"errors"
	"testing"
)

// TestReportableError_As checks if errors.As works with RemoteReportableError.
func TestReportableError_As(t *testing.T) {
	underlyingErr := errors.New("underlying error")
	reportableErr := NewRemoteReportableError(underlyingErr)

	var target *RemoteReportableError
	if !errors.As(reportableErr, &target) {
		t.Fatalf("expected errors.As to identify RemoteReportableError")
	}
}

// TestReportableError_Message checks if the underlying error's message is correctly output.
func TestReportableError_Message(t *testing.T) {
	underlyingErr := errors.New("underlying error message")
	reportableErr := NewRemoteReportableError(underlyingErr)

	if reportableErr.Error() != "underlying error message" {
		t.Fatalf("expected error message to be %q, got %q", "underlying error message", reportableErr.Error())
	}
}
