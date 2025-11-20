package capabilities

import (
	"errors"
	"strconv"
	"strings"
)

const remoteReportableErrorIdentifier = "RemoteReportableError:"

const reportableUserErrorIdentifier = remoteReportableErrorIdentifier + "UserError:"

const localReportableErrorIdentifier = "LocalReportableError:"

const errorCodeIdentifier = "ErrorCode="

func PrePendLocalReportableErrorIdentifier(errorMessage string) string {
	return localReportableErrorIdentifier + errorMessage
}

func IsReportableUserErrorMessage(message string) bool {
	return strings.HasPrefix(message, reportableUserErrorIdentifier)
}

// GetErrorCode Returns the error code and removes it from the message if present.
func GetErrorCode(message string) (ErrorCode, string) {
	if strings.HasPrefix(message, errorCodeIdentifier) {
		rest := message[len(errorCodeIdentifier):]
		colonIdx := strings.Index(rest, ":")
		if colonIdx != -1 {
			codeStr := rest[:colonIdx]
			code, err := strconv.Atoi(codeStr)
			if err == nil {
				return ErrorCodeFromInt(code), rest[colonIdx+1:]
			}
		}
	}
	return Uncategorized, message
}

func DeserializeErrorFromString(errorMsg string) Error {
	// Order is important here as reportable user errors also have the remote reportable error identifier.
	if strings.HasPrefix(errorMsg, reportableUserErrorIdentifier) {
		errorMsg = strings.TrimPrefix(errorMsg, reportableUserErrorIdentifier)
		errorCode, msg := GetErrorCode(errorMsg)
		return NewReportableUserError(errors.New(msg), errorCode)
	}

	if strings.HasPrefix(errorMsg, remoteReportableErrorIdentifier) {
		msg := strings.TrimPrefix(errorMsg, remoteReportableErrorIdentifier)
		errorCode, msg := GetErrorCode(msg)
		return NewRemoteReportableError(errors.New(msg), errorCode)
	}

	if strings.HasPrefix(errorMsg, localReportableErrorIdentifier) {
		msg := strings.TrimPrefix(errorMsg, localReportableErrorIdentifier)
		errorCode, msg := GetErrorCode(msg)
		return NewLocalReportableError(errors.New(msg), errorCode)
	}

	// Default to local reportable error if no identifier is found.
	errorCode, errorMsg := GetErrorCode(errorMsg)
	return NewLocalReportableError(errors.New(errorMsg), errorCode)
}

func (e *capabilityError) SerializeToString() string {
	var prefix string
	switch e.ReportType() {
	case ErrorReportTypeRemote:
		prefix = remoteReportableErrorIdentifier
	case ErrorReportTypeUser:
		prefix = reportableUserErrorIdentifier
	case ErrorReportTypeLocal:
		prefix = localReportableErrorIdentifier
	}

	return prefix + errorCodeIdentifier + strconv.Itoa(int(e.Code())) + ":" + e.err.Error()
}
