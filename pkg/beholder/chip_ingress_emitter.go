package beholder

import (
	"context"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
)

type chipIngressEmitter struct {
	client chipingress.ChipIngressClient
}

func NewChipIngressEmitter(client chipingress.ChipIngressClient) Emitter {
	return &chipIngressEmitter{client: client,}
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
