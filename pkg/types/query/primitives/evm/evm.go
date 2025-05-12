package evm

import (
	"github.com/smartcontractkit/chainlink-common/pkg/types/chains/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
)

type Visitor interface {
	Address(primitive *AddressFilter)
	EventSig(primitive *EventSig)
	EventTopicsByValue(primitive *EventByTopic)
	EventByWord(primitive *EventByWord)
}

type AddressFilter struct {
	Address evm.Address
}

func NewAddressFilter(address evm.Address) query.Expression {
	return query.Expression{
		Primitive: &AddressFilter{Address: address},
	}
}

func (a *AddressFilter) Accept(visitor primitives.Visitor) {
	if v, ok := visitor.(Visitor); ok {
		v.Address(a)
	}
}

type EventSig struct {
	EventSig evm.Hash
}

func NewEventSigFilter(eventSig evm.Hash) query.Expression {
	return query.Expression{
		Primitive: &EventSig{EventSig: eventSig},
	}
}

func (es *EventSig) Accept(visitor primitives.Visitor) {
	if v, ok := visitor.(Visitor); ok {
		v.EventSig(es)
	}
}

type HashedValueComparator struct {
	Values   []evm.Hash
	Operator primitives.ComparisonOperator
}

type EventByWord struct {
	WordIndex            int
	HashedValueComparers []HashedValueComparator
}

func NewEventByWordFilter(wordIndex int, valueComparers []HashedValueComparator) query.Expression {
	return query.Expression{
		Primitive: &EventByWord{
			WordIndex:            wordIndex,
			HashedValueComparers: valueComparers,
		},
	}
}

func (ew *EventByWord) Accept(visitor primitives.Visitor) {
	if v, ok := visitor.(Visitor); ok {
		v.EventByWord(ew)
	}
}

type EventByTopic struct {
	Topic                 uint64
	HashedValueComprarers []HashedValueComparator
}

func NewEventByTopicFilter(topic uint64, valueComprarer []HashedValueComparator) query.Expression {
	return query.Expression{
		Primitive: &EventByTopic{
			Topic:                 topic,
			HashedValueComprarers: valueComprarer,
		},
	}
}

func (et *EventByTopic) Accept(visitor primitives.Visitor) {
	if v, ok := visitor.(Visitor); ok {
		v.EventTopicsByValue(et)
	}
}
