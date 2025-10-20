package testdata

import (
	"context"

	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
)

type Nested interface {
	GetInt(ctx context.Context) (uint64, error)
	GetIntSlice(ctx context.Context) ([]int, error)
}

type TestFace interface {
	Nested
	GetIntArr(ctx context.Context) ([10]int, error)
	GetBytes(ctx context.Context, b []byte) ([]byte, error)
	SendStruct(ctx context.Context, ms MyStruct) (MyStruct, error)
	SendNested(ctx context.Context, ns NestedStruct) (NestedStruct, error)
}

type SolanaService interface {
	// SendTx(ctx context.Context, tx *solanago.Transaction) (solanago.Signature, error)
	// SimulateTx(ctx context.Context, tx *solanago.Transaction, opts *rpc.SimulateTransactionOpts) (*rpc.SimulateTransactionResult, error)
}

type MyStruct struct {
	B    []byte
	Prim primitives.ConfidenceLevel
	Expr query.Expression
}

type NestedStruct struct {
	F1 MyStruct
	F2 MyStruct
}
