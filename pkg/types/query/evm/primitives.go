package evm

import (
	"github.com/smartcontractkit/chainlink-common/pkg/types/chains/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
)

type Address = evm.Address
type Hash = evm.Hash

type EVMVisitor interface {
	VisitAddressFilter(f *AddressFilter)
	VisitEventSigFilter(f *EventSig)
	VisitEventTopicsByValueFilter(f *EventByTopic)
	VisitEventByWordFilter(f *EventByWord)
}

type AddressFilter struct {
	Address Address
}

func NewAddressFilter(address Address) query.Expression {
	return query.Expression{
		Primitive: &AddressFilter{Address: address},
	}
}

func (a *AddressFilter) Accept(visitor primitives.Visitor) {
	if v, ok := visitor.(EVMVisitor); ok {
		v.VisitAddressFilter(a)
	}
}

type EventSig struct {
	EventSig Hash
}

func NewEventSigFilter(eventSig Hash) query.Expression {
	return query.Expression{
		Primitive: &EventSig{EventSig: eventSig},
	}
}

func (es *EventSig) Accept(visitor primitives.Visitor) {
	if v, ok := visitor.(EVMVisitor); ok {
		v.VisitEventSigFilter(es)
	}
}

type HashedValueComparator struct {
	Values   []Hash
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
	if v, ok := visitor.(EVMVisitor); ok {
		v.VisitEventByWordFilter(ew)
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
	if v, ok := visitor.(EVMVisitor); ok {
		v.VisitEventTopicsByValueFilter(et)
	}
}
