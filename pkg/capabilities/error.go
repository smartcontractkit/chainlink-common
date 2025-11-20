package capabilities

import "fmt"

type ErrorReportType int

const (
	// ErrorReportTypeLocal The message in the error may contain sensitive node local information and should not be reported remotely.
	// In addition by marking the error as local it prevents the error from being accidentally or maliciously marked as remote reportable
	// by injecting the remote reportable error identifier in front of the error message.
	ErrorReportTypeLocal ErrorReportType = 0

	// ErrorReportTypeRemote The message in the error is safe to report remotely between nodes.
	ErrorReportTypeRemote ErrorReportType = 1

	// ErrorReportTypeUser The error is due to user error and is safe to report remotely between nodes.
	ErrorReportTypeUser ErrorReportType = 2
)

type Error interface {
	error

	ReportType() ErrorReportType
	Code() ErrorCode
}

type capabilityError struct {
	err        error
	reportType ErrorReportType
	errorCode  ErrorCode
}

func newError(err error, reportType ErrorReportType, errorCode ErrorCode) Error {
	return &capabilityError{
		err:        err,
		reportType: reportType,
		errorCode:  errorCode,
	}
}

func NewRemoteReportableError(err error, errorCode ErrorCode) Error {
	return newError(err, ErrorReportTypeRemote, errorCode)
}

func NewUserError(err error, errorCode ErrorCode) Error {
	return newError(err, ErrorReportTypeUser, errorCode)
}

func NewLocalReportableError(err error, errorCode ErrorCode) Error {
	return newError(err, ErrorReportTypeLocal, errorCode)
}

func (e *capabilityError) Error() string {
	return fmt.Sprintf("[%d]%s:", e.errorCode, e.errorCode.String()) + " " + e.err.Error()
}

func (e *capabilityError) ReportType() ErrorReportType {
	return e.reportType
}

func (e *capabilityError) Code() ErrorCode {
	return e.errorCode
}
