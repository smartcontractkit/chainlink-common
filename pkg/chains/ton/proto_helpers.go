package ton

import (
	"time"

	types "github.com/smartcontractkit/chainlink-common/pkg/types/chains/ton"
	valuespb "github.com/smartcontractkit/chainlink-protos/cre/go/values/pb"
)

func NewBlockIDExt(block *types.BlockIDExt) *BlockIDExt {
	if block == nil {
		return nil
	}
	return &BlockIDExt{
		Workchain: block.Workchain,
		Shard:     block.Shard,
		SeqNo:     block.SeqNo,
	}
}

func (pb *BlockIDExt) AsBlockIDExt() *types.BlockIDExt {
	if pb == nil {
		return nil
	}
	return &types.BlockIDExt{
		Workchain: pb.Workchain,
		Shard:     pb.Shard,
		SeqNo:     pb.SeqNo,
	}
}

func NewBlock(block *types.Block) *Block {
	if block == nil {
		return nil
	}
	return &Block{
		GlobalId: block.GlobalID,
	}
}

func (pb *Block) AsBlock() *types.Block {
	if pb == nil {
		return nil
	}
	return &types.Block{
		GlobalID: pb.GlobalId,
	}
}

func NewBalance(balance *types.Balance) *Balance {
	if balance == nil {
		return nil
	}
	return &Balance{
		Balance: valuespb.NewBigIntFromInt(balance.Balance),
	}
}

func (pb *Balance) AsBalance() *types.Balance {
	if pb == nil {
		return nil
	}
	return &types.Balance{
		Balance: valuespb.NewIntFromBigInt(pb.Balance),
	}
}

func NewMessage(msg *types.Message) *Message {
	if msg == nil {
		return nil
	}
	return &Message{
		Mode:      uint32(msg.Mode),
		ToAddress: msg.ToAddress,
		Amount:    msg.Amount,
		Bounce:    msg.Bounce,
		Body:      msg.Body,
		StateInit: msg.StateInit,
	}
}

func (pb *Message) AsMessage() *types.Message {
	if pb == nil {
		return nil
	}
	return &types.Message{
		Mode:      uint8(pb.Mode),
		ToAddress: pb.ToAddress,
		Amount:    pb.Amount,
		Bounce:    pb.Bounce,
		Body:      pb.Body,
		StateInit: pb.StateInit,
	}
}

func NewLPFilter(f types.LPFilterQuery) *LPFilterQuery {
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

func (pb *LPFilterQuery) AsLPFilter() types.LPFilterQuery {
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
