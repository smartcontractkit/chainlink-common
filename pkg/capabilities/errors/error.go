package errors

import "fmt"

type ReportType int

const (
	// LocalOnly The message in the error may contain sensitive node local information and should not be reported remotely.
	// In addition the serialised string representation is prefixed with an identify that prevents the error
	// from being accidentally or maliciously marked as remote reportable by manipulating the error string.
	LocalOnly ReportType = 0

	// RemoteReportable The message in the error is safe to report remotely between nodes.
	RemoteReportable ReportType = 1

	// ReportableUser The error is due to user error and is safe to report remotely between nodes.
	ReportableUser ReportType = 2
)

type Error interface {
	error

	ReportType() ReportType
	Code() ErrorCode
	SerializeToString() string
	SerializeToRemoteReportableString() string
}

type capabilityError struct {
	err        error
	reportType ReportType
	errorCode  ErrorCode
}

func newError(err error, reportType ReportType, errorCode ErrorCode) Error {
	return &capabilityError{
		err:        err,
		reportType: reportType,
		errorCode:  errorCode,
	}
}

// NewRemoteReportableError indicates that the wrapped error does not contain any node local confidential information
// and is safe to report to other nodes in the network.
func NewRemoteReportableError(err error, errorCode ErrorCode) Error {
	return newError(err, RemoteReportable, errorCode)
}

// NewReportableUserError indicates that the wrapped error is due to user error and does not contain any node local confidential information
// and is safe to report to other nodes in the network.
func NewReportableUserError(err error, errorCode ErrorCode) Error {
	return newError(err, ReportableUser, errorCode)
}

// NewLocalReportableError indicates that the wrapped error may contain node local confidential information
// that should not be reported to other nodes in the network.  Only the error code and generic message will be reported remotely.
func NewLocalReportableError(err error, errorCode ErrorCode) Error {
	return newError(err, LocalOnly, errorCode)
}

func (e *capabilityError) Error() string {
	return fmt.Sprintf("[%d]%s:", e.errorCode, e.errorCode.String()) + " " + e.err.Error()
}

func (e *capabilityError) ReportType() ReportType {
	return e.reportType
}

func (e *capabilityError) Code() ErrorCode {
	return e.errorCode
}
