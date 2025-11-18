package capabilities

import (
	"errors"
	"testing"
)

// TestReportableUserError_As checks if errors.As works with ReportableUserError.
func TestReportableUserError_As(t *testing.T) {
	underlyingErr := NewRemoteReportableError(errors.New("underlying error"))
	userErr := NewReportableUserError(underlyingErr)

	var target *ReportableUserError
	if !errors.As(userErr, &target) {
		t.Fatalf("expected errors.As to identify ReportableUserError")
	}
}

// TestReportableUserError_Message checks if the underlying error's message is correctly output.
func TestReportableUserError_Message(t *testing.T) {
	underlyingErr := errors.New("underlying user error message")
	userErr := NewReportableUserError(underlyingErr)

	if userErr.Error() != "underlying user error message" {
		t.Fatalf("expected error message to be %q, got %q", "underlying user error message", userErr.Error())
	}
}

func TestAsRemoteReportableErrorForRemoteReportableUserError(t *testing.T) {
	underlyingErr := errors.New("underlying user error message")
	userErr := NewReportableUserError(underlyingErr)

	var target *ReportableUserError
	if !errors.As(userErr, &target) {
		t.Fatalf("expected errors.As to identify ReportableUserError")
	}

	var targetReportableError *RemoteReportableError
	if !errors.As(userErr, &targetReportableError) {
		t.Fatalf("expected errors.As to identify RemoteReportableError from ReportableUserError")
	}
}
