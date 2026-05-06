package relayer

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	loopnet "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
)

func TestReplaceActiveRelayerClosesPreviousResource(t *testing.T) {
	server := newPluginRelayerServer(nil, loopnet.BrokerConfig{Logger: logger.Test(t)}, nil)
	first := &countingCloser{}
	second := &countingCloser{}

	server.replaceActiveRelayer(loopnet.Resource{Closer: first, Name: "first"})
	require.Equal(t, 0, first.closed)

	server.replaceActiveRelayer(loopnet.Resource{Closer: second, Name: "second"})
	require.Equal(t, 1, first.closed)
	require.Equal(t, 0, second.closed)
}

type countingCloser struct {
	closed int
}

func (c *countingCloser) Close() error {
	c.closed++
	return nil
}
