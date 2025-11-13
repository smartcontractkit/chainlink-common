package capabilities

import (
	"strings"
)

const remoteUnreportableErrorIdentifier = "RemoteUnreportableError:"

func PrePendRemoteUnreportableErrorIdentifier(errorMessage string) string {
	return remoteUnreportableErrorIdentifier + errorMessage
}

func RemoveRemoteUnreportableErrorIdentifier(message string) string {
	return strings.TrimPrefix(message, remoteUnreportableErrorIdentifier)
}
