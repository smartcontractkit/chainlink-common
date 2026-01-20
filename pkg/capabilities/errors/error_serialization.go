package errors

import (
	"errors"
	"strings"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/errors/pb"
)

const errorMessageSeparator = ":"
const notPubliclyReportableErrorMsg = "error whilst executing capability - the error message is not publicly reportable"

func PrePendPrivateVisibilityIdentifier(errorMessage string) string {
	return VisibilityPrivate.String() + errorMessageSeparator + errorMessage
}

func DeserializeErrorFromString(errorMsg string) Error {
	parts := strings.SplitN(errorMsg, errorMessageSeparator, 4)

	if len(parts) < 4 {
		// To maintain backwards compatability with messages from remote nodes on an older code version, create an error
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

	return e.serializeToString(notPubliclyReportableErrorMsg)
}

// ToRemoteProto serializes the error to a protobuf message for sending to remote nodes.
// If the error is private, the actual error message is replaced with a generic message.
func (e capabilityError) ToRemoteProto() *pb.Error {
	msg := e.err.Error()
	if e.Visibility() == VisibilityPrivate {
		msg = notPubliclyReportableErrorMsg
	}

	return &pb.Error{
		Visibility: pb.Visibility(e.visibility),
		Origin:     pb.Origin(e.origin),
		Code:       uint32(e.Code()),
		Message:    msg,
	}
}

func FromProto(pbErr *pb.Error) Error {
	if pbErr == nil {
		return nil
	}

	visibility := Visibility(pbErr.Visibility)
	origin := Origin(pbErr.Origin)
	errorCode := ErrorCode(pbErr.Code)
	errorMsg := pbErr.Message

	return NewError(errors.New(errorMsg), visibility, origin, errorCode)
}
