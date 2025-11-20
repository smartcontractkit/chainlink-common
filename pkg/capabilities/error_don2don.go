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

func PrePendRemoteUnreportableErrorIdentifier(errorMessage string) string {
	return localReportableErrorIdentifier + errorMessage
}

func RemoveRemoteUnreportableErrorIdentifier(message string) string {
	return strings.TrimPrefix(message, localReportableErrorIdentifier)
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

// Returns the error code and removes it from the message if present.
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

func ToCapabilityError(errorMsg string) Error {
	// Order is important here as reportable user errors also have the remote reportable error identifier.
	if strings.HasPrefix(errorMsg, reportableUserErrorIdentifier) {
		errorMsg = RemoveReportableUserErrorIdentifier(errorMsg)
		errorCode, msg := GetErrorCode(errorMsg)
		return NewUserError(errors.New(msg), errorCode)
	}

	if IsRemoteReportableErrorMessage(errorMsg) {
		msg := RemoveRemoteReportableErrorIdentifier(errorMsg)
		errorCode, msg := GetErrorCode(msg)
		return NewRemoteReportableError(errors.New(msg), errorCode)
	}

	if strings.HasPrefix(errorMsg, localReportableErrorIdentifier) {
		msg := RemoveRemoteUnreportableErrorIdentifier(errorMsg)
		errorCode, msg := GetErrorCode(msg)
		return NewLocalReportableError(errors.New(msg), errorCode)
	}

	// Default to remote reportable error if no identifier is found.
	errorCode, errorMsg := GetErrorCode(errorMsg)
	return NewLocalReportableError(errors.New(errorMsg), errorCode)
}
