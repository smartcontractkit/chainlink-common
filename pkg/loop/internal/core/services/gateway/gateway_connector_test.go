package gateway_test

import (
	"testing"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/gateway"
	loopnet "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	loopnettest "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net/test"
	pb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/gatewayconnector"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"

	"github.com/stretchr/testify/require"
)

func Test_GatewayConnectorServer(t *testing.T) {
	t.Run("calling AddHandler with a nil connector client does not panic", func(t *testing.T) {
		var (
			lggr      = logger.Test(t)
			ctx       = t.Context()
			handlerID = uint32(0)
			broker    = &loopnettest.Broker{T: t}
			brokerCfg = loopnet.BrokerConfig{Logger: lggr, StopCh: make(chan struct{})}
			brokerExt = &loopnet.BrokerExt{
				Broker:       broker,
				BrokerConfig: brokerCfg,
			}
		)

		// allocate a listener for the handler
		_, err := broker.Accept(handlerID)
		require.NoError(t, err)

		// create instance of connector server with empty GatewayConnector
		var gc core.GatewayConnector
		gcs := gateway.NewGatewayConnectorServer(brokerExt, gc)

		// assert that call does not panic, yet errors
		res, err := gcs.AddHandler(ctx, &pb.AddHandlerRequest{
			HandlerId: handlerID,
		})
		require.Error(t, err)
		require.ErrorContains(t, err, "not implemented")
		require.Nil(t, res)
	})
}
