package standalone

import (
	"fmt"
	"net"
	"strconv"
)

// OffsetPort returns hostPort with its port raised by delta, keeping the host part as it is:
// OffsetPort(":50051", 2) is ":50053" and OffsetPort("127.0.0.1:1234", 1) is "127.0.0.1:1235".
//
// This is what lets several embedded instances share one configured address: the address a single
// instance would take literally becomes instance i's address by adding i to it, so nothing has to
// be configured per instance. Port 0 is returned unchanged, since every instance asking the OS for
// an ephemeral port already gets a distinct one.
func OffsetPort(hostPort string, delta int) (string, error) {
	host, portStr, err := net.SplitHostPort(hostPort)
	if err != nil {
		return "", fmt.Errorf("failed to parse address %q: %w", hostPort, err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse port of address %q: %w", hostPort, err)
	}
	if port == 0 {
		return hostPort, nil
	}
	return net.JoinHostPort(host, strconv.Itoa(port+delta)), nil
}

// OffsetPorts applies OffsetPort to every address, returning a new slice. A nil or empty slice
// returns nil, so an unset setting stays unset.
func OffsetPorts(hostPorts []string, delta int) ([]string, error) {
	if len(hostPorts) == 0 {
		return nil, nil
	}
	out := make([]string, len(hostPorts))
	for i, hostPort := range hostPorts {
		offset, err := OffsetPort(hostPort, delta)
		if err != nil {
			return nil, err
		}
		out[i] = offset
	}
	return out, nil
}
