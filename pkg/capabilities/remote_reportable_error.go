package capabilities

import (
	"fmt"
	"strings"
)

const remoteReportableErrorIdentifier = "RemoteReportableError:"

// RemoteReportableError wraps an error to indicate that the error does contain any node specific
// information and is safe to report remotely between nodes.
type RemoteReportableError struct {
	err error
}

func NewRemoteReportableError(err error) *RemoteReportableError {
	return &RemoteReportableError{err: err}
}

func (e *RemoteReportableError) Error() string {
	if e.err != nil {
		return fmt.Sprintf("%v", e.err)
	}
	return ""
}

// Unwrap allows errors.Is and errors.As to work with CustomError.
func (e *RemoteReportableError) Unwrap() error {
	return e.err
}

func PrePendRemoteReportableErrorIdentifier(errorMessage string) string {
	return remoteReportableErrorIdentifier + errorMessage
}

func IsRemoteReportableErrorMessage(message string) bool {
	return strings.HasPrefix(message, remoteReportableErrorIdentifier)
}

func RemoveRemoteReportableErrorIdentifier(message string) string {
	if IsRemoteReportableErrorMessage(message) {
		return strings.TrimPrefix(message, remoteReportableErrorIdentifier)
	}
	return message
}
