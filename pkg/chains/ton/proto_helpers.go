package ton

import (
	"time"

	types "github.com/smartcontractkit/chainlink-common/pkg/types/chains/ton"
	valuespb "github.com/smartcontractkit/chainlink-common/pkg/values/pb"
)

func ConvertBlockIDExtToProto(block *types.BlockIDExt) *BlockIDExt {
	if block == nil {
		return nil
	}
	return &BlockIDExt{
		Workchain: block.Workchain,
		Shard:     block.Shard,
		Seqno:     block.Seqno,
	}
}

func ConvertBlockIDExtFromProto(pb *BlockIDExt) *types.BlockIDExt {
	if pb == nil {
		return nil
	}
	return &types.BlockIDExt{
		Workchain: pb.Workchain,
		Shard:     pb.Shard,
		Seqno:     pb.Seqno,
	}
}

func ConvertBlockToProto(block *types.Block) *Block {
	if block == nil {
		return nil
	}
	return &Block{
		GlobalId: block.GlobalID,
	}
}

func ConvertBlockFromProto(pb *Block) *types.Block {
	if pb == nil {
		return nil
	}
	return &types.Block{
		GlobalID: pb.GlobalId,
	}
}

func ConvertBalanceToProto(balance *types.Balance) *Balance {
	if balance == nil {
		return nil
	}
	return &Balance{
		Balance: valuespb.NewBigIntFromInt(balance.Balance),
	}
}

func ConvertBalanceFromProto(pb *Balance) *types.Balance {
	if pb == nil {
		return nil
	}
	return &types.Balance{
		Balance: valuespb.NewIntFromBigInt(pb.Balance),
	}
}

func ConvertMessageToProto(msg *types.Message) *Message {
	if msg == nil {
		return nil
	}
	return &Message{
		Mode:       uint32(msg.Mode),
		ToAddress:  msg.ToAddress,
		AmountNano: msg.AmountNano,
		Bounce:     msg.Bounce,
		Body:       msg.Body,
		StateInit:  msg.StateInit,
	}
}

func ConvertMessageFromProto(pb *Message) *types.Message {
	if pb == nil {
		return nil
	}
	return &types.Message{
		Mode:       uint8(pb.Mode),
		ToAddress:  pb.ToAddress,
		AmountNano: pb.AmountNano,
		Bounce:     pb.Bounce,
		Body:       pb.Body,
		StateInit:  pb.StateInit,
	}
}

func ConvertLPFilterToProto(f types.LPFilterQuery) *LPFilterQuery {
	return &LPFilterQuery{
		Id:            f.ID,
		Name:          f.Name,
		Address:       f.Address,
		EventName:     f.EventName,
		EventTopic:    f.EventTopic,
		StartingSeq:   f.StartingSeq,
		RetentionSecs: int64(f.Retention.Seconds()),
	}
}

func ConvertLPFilterFromProto(pb *LPFilterQuery) types.LPFilterQuery {
	return types.LPFilterQuery{
		ID:          pb.Id,
		Name:        pb.Name,
		Address:     pb.Address,
		EventName:   pb.EventName,
		EventTopic:  pb.EventTopic,
		StartingSeq: pb.StartingSeq,
		Retention:   time.Duration(pb.RetentionSecs) * time.Second,
	}
}
