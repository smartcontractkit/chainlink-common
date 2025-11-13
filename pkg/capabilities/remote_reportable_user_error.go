package capabilities

import (
	"strings"
)

const remoteReportableUserErrorIdentifier = remoteReportableErrorIdentifier + "UserError:"

// RemoteReportableUserError indicates the error is remote reportable and caused by user error.
type RemoteReportableUserError struct {
	err *RemoteReportableError
}

func NewRemoteReportableUserError(err error) *RemoteReportableUserError {
	return &RemoteReportableUserError{err: &RemoteReportableError{err}}
}

func (e *RemoteReportableUserError) Error() string {
	if e.err == nil {
		return ""
	}

	return e.err.Error()
}

func (e *RemoteReportableUserError) Unwrap() error {
	return e.err
}

func PrePendRemoteReportableUserErrorIdentifier(errorMessage string) string {
	return remoteReportableUserErrorIdentifier + errorMessage
}

func IsRemoteReportableUserErrorMessage(message string) bool {
	return strings.HasPrefix(message, remoteReportableUserErrorIdentifier)
}

func RemoveRemoteReportableUserErrorIdentifier(message string) string {
	if IsRemoteReportableUserErrorMessage(message) {
		return strings.TrimPrefix(message, remoteReportableUserErrorIdentifier)
	}
	return message
}
