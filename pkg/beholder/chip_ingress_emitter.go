package beholder

import (
	"context"
	"fmt"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
	"google.golang.org/grpc/credentials/insecure"
)

type chipIngressEmitter struct {
	client chipingress.ChipIngressClient
}

func NewChipIngressEmitter(client chipingress.ChipIngressClient) Emitter {
	return &chipIngressEmitter{client: client,}
}

func NewChipIngressClient(cfg Config) (chipingress.ChipIngressClient, error) {

	if cfg.ChipIngressEmitterGRPCEndpoint == "" {
		return nil, fmt.Errorf("missing chip ingress emitter gRPC endpoint")
	}

	// TODO: add support for csa auth signing interceptor
	// We should add csa signed headers, that will be authenticated on the server-side
	client, err := chipingress.NewChipIngressClient(
		cfg.ChipIngressEmitterGRPCEndpoint,
		chipingress.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (c *chipIngressEmitter) Emit(ctx context.Context, body []byte, attrKVs ...any) error {

	baseMsg := &pb.BaseMessage{
		Msg: "test",
	}

	event, err := chipingress.NewEvent("test", "test", baseMsg)
	if err != nil {
		return err
	}

	_, err = c.client.Publish(ctx, event)
	if err != nil {
		return err
	}

	return nil
}
