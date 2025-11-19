package capabilities

import (
	"strings"
)

const reportableUserErrorIdentifier = remoteReportableErrorIdentifier + "UserError:"

// ReportableUserError indicates the error is a user error that is both locally and remotely reportable.
type ReportableUserError struct {
	err *RemoteReportableError
}

func NewReportableUserError(err error) *ReportableUserError {
	return &ReportableUserError{err: &RemoteReportableError{err}}
}

func (e *ReportableUserError) Error() string {
	if e.err == nil {
		return ""
	}

	return e.err.Error()
}

func (e *ReportableUserError) Unwrap() error {
	return e.err
}

func PrePendReportableUserErrorIdentifier(errorMessage string) string {
	return reportableUserErrorIdentifier + errorMessage
}

func IsReportableUserErrorMessage(message string) bool {
	return strings.HasPrefix(message, reportableUserErrorIdentifier)
}

func RemoveReportableUserErrorIdentifier(message string) string {
	if IsReportableUserErrorMessage(message) {
		return strings.TrimPrefix(message, reportableUserErrorIdentifier)
	}
	return message
}
