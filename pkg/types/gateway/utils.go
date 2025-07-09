package gateway

import (
	"strings"

	"github.com/google/uuid"
)

// GetRequestID generates a unique request ID for a method call.
// The method name is prepended to the ID, followed by any identifiers such as
// workflow ID, execution ID, or request ID.
// The method name is used by the gateway to identify the message type
// (e.g. "http_action", "push_auth_metadata", "pull_auth_metadata").
func GetRequestID(methodName string, parts ...string) string {
	id := append([]string{methodName}, parts...)
	id = append(id, uuid.New().String())
	return strings.Join(id, "/")
}
