package errors

import (
	"errors"
	"strings"
)

const errorMessageSeparator = ":"

func PrePendPrivateVisibilityIdentifier(errorMessage string) string {
	return VisibilityPrivate.String() + errorMessageSeparator + errorMessage
}

func DeserializeErrorFromString(errorMsg string) Error {
	parts := strings.SplitN(errorMsg, errorMessageSeparator, 4)

	if len(parts) < 4 {
		// To maintain backwards compatibility with messages from remote nodes on an older code version, create an error
		// with the full message and default to private system error with an unknown error code.
		return NewError(errors.New(errorMsg), VisibilityPrivate, OriginSystem, Unknown)
	}

	visibility := FromVisibilityString(parts[0])
	origin := FromOriginString(parts[1])
	errorCode := FromErrorCodeString(parts[2])
	errorMsg = parts[3]

	return NewError(errors.New(errorMsg), visibility, origin, errorCode)
}

func (e capabilityError) SerializeToString() string {
	return e.serializeToString(e.err.Error())
}

func (e capabilityError) serializeToString(errMsg string) string {
	return e.visibility.String() + errorMessageSeparator + e.origin.String() + errorMessageSeparator + e.Code().String() + errorMessageSeparator + errMsg
}

// SerializeToRemoteString serializes the error for sending to remote nodes.
// If the error is private, the actual error message is replaced with a generic message.
func (e capabilityError) SerializeToRemoteString() string {
	if e.Visibility() == VisibilityPublic {
		return e.serializeToString(e.err.Error())
	}

	return e.serializeToString("error whilst executing capability - the error message is not publicly reportable")
}
