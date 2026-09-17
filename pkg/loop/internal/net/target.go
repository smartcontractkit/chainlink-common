package net

import (
	"fmt"
	"strconv"
	"strings"
)

// BrokerScheme is the gRPC target scheme for connections brokered by the go-plugin
// host. A target "broker://<id>" names a connection served on this process's
// [Broker]; it is meaningful only to the loop packages, which own the broker.
const BrokerScheme = "broker"

const brokerTargetPrefix = BrokerScheme + "://"

// BrokerTarget returns a gRPC target naming the brokered connection id.
func BrokerTarget(id uint32) string {
	return brokerTargetPrefix + strconv.FormatUint(uint64(id), 10)
}

// ParseBrokerTarget returns the brokered connection id named by target.
func ParseBrokerTarget(target string) (uint32, error) {
	rest, ok := strings.CutPrefix(target, brokerTargetPrefix)
	if !ok {
		return 0, fmt.Errorf("not a %s target: %q", BrokerScheme, target)
	}
	id, err := strconv.ParseUint(rest, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid %s target %q: %w", BrokerScheme, target, err)
	}
	return uint32(id), nil
}
