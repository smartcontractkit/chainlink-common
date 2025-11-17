package solana

import (
	"github.com/smartcontractkit/chainlink-common/pkg/types/chains/solana"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
)

type Visitor interface {
	Address(primitive *Address)
	EventSig(primitive *EventSig)
	EventBySubkey(primitive *EventBySubkey)
}

type Address struct {
	PubKey solana.PublicKey
}

func (a *Address) Accept(visitor primitives.Visitor) {
	if v, ok := visitor.(Visitor); ok {
		v.Address(a)
	}
}

func NewAddressFilter(address solana.PublicKey) query.Expression {
	return query.Expression{
		Primitive: &Address{PubKey: address},
	}
}

type EventSig struct {
	Sig solana.EventSignature
}

func (e *EventSig) Accept(visitor primitives.Visitor) {
	if v, ok := visitor.(Visitor); ok {
		v.EventSig(e)
	}
}

func NewEventSigFilter(e solana.EventSignature) query.Expression {
	return query.Expression{
		Primitive: &EventSig{
			Sig: e,
		},
	}
}

type EventBySubkey struct {
	SubKeyIndex    uint64
	ValueComparers []IndexedValueComparator
}

type IndexedValueComparator struct {
	Value    []solana.IndexedValue
	Operator primitives.ComparisonOperator
}

func (esk *EventBySubkey) Accept(visitor primitives.Visitor) {
	if v, ok := visitor.(Visitor); ok {
		v.EventBySubkey(esk)
	}
}

func NewEventBySubkeyFilter(subkeyIndex uint64, comparers []IndexedValueComparator) query.Expression {
	return query.Expression{
		Primitive: &EventBySubkey{
			SubKeyIndex:    subkeyIndex,
			ValueComparers: comparers,
		},
	}
}
