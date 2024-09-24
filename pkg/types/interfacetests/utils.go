package interfacetests

import (
	"cmp"
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/smartcontractkit/libocr/commontypes"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

type BasicTester[T any] interface {
	Setup(t T)
	Name() string
	GetAccountBytes(i int) []byte
}

type testcase[T any] struct {
	name string
	test func(t T)
}

type TestingT[T any] interface {
	tests.TestingT
	Failed() bool
	Run(name string, f func(t T)) bool
}

func runTests[T TestingT[T]](t T, tester BasicTester[T], tests []testcase[T]) {
	for _, test := range tests {
		t.Run(test.name+" for "+tester.Name(), func(t T) {
			tester.Setup(t)
			test.test(t)
		})
	}
}

// Batch chain write takes a batch call entry and writes it to the chain using the ChainWriter.
func batchChainWrite[T TestingT[T]](t T, tester ChainComponentsInterfaceTester[T], batchCallEntry BatchCallEntry, mockRun bool) {
	// This is necessary because the mock helper function requires the entire batchCallEntry rather than an individual testStruct
	if mockRun {
		cw := tester.GetChainWriter(t)
		err := cw.SubmitTransaction(tests.Context(t), AnyContractName, "batchChainWrite", batchCallEntry, "", "", nil, big.NewInt(0))
		require.NoError(t, err)
		return
	}
	nameToAddress := make(map[string]string)
	boundContracts := tester.GetBindings(t)
	for _, bc := range boundContracts {
		nameToAddress[bc.Name] = bc.Address
	}

	// For each contract in the batch call entry, submit the read entries to the chain
	for contract, contractBatch := range batchCallEntry {
		require.Contains(t, nameToAddress, contract.Name)
		for _, readEntry := range contractBatch {
			val, isOk := readEntry.ReturnValue.(*TestStruct)
			if !isOk {
				require.Fail(t, "expected *TestStruct for contract: %s read: %s, but received %T", contract.Name, readEntry.Name, readEntry.ReturnValue)
			}
			SubmitTransactionToCW(t, tester, MethodSettingStruct, val, types.BoundContract{Name: contract.Name, Address: nameToAddress[contract.Name]}, types.Unconfirmed)
		}
	}
}

// SubmitTransactionToCW submits a transaction to the ChainWriter and waits for it to reach the given status.
func SubmitTransactionToCW[T TestingT[T]](t T, tester ChainComponentsInterfaceTester[T], method string, args any, contract types.BoundContract, status types.TransactionStatus) string {
	tester.DirtyContracts()
	txID := uuid.New().String()
	cw := tester.GetChainWriter(t)
	err := cw.SubmitTransaction(tests.Context(t), contract.Name, method, args, txID, contract.Address, nil, big.NewInt(0))
	require.NoError(t, err)

	err = WaitForTransactionStatus(t, tester, txID, status, false)
	require.NoError(t, err)

	return txID
}

// WaitForTransactionStatus waits for a transaction to reach the given status.
func WaitForTransactionStatus[T TestingT[T]](t T, tester ChainComponentsInterfaceTester[T], txID string, status types.TransactionStatus, mockRun bool) error {
	ctx, cancel := context.WithTimeout(tests.Context(t), 15*time.Minute)
	defer cancel()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("transaction %s not finalized within timeout period", txID)
		case <-ticker.C:
			if mockRun {
				tester.GenerateBlocksTillConfidenceLevel(t, "", "", primitives.Finalized)
				return nil
			}
			current, err := tester.GetChainWriter(t).GetTransactionStatus(ctx, txID)
			if err != nil {
				return fmt.Errorf("failed to get transaction status: %w", err)
			}

			if current == types.Failed || current == types.Fatal {
				return fmt.Errorf("transaction %s has failed or is fatal", txID)
			} else if current >= status {
				return nil
			} else {
				continue
			}
		}
	}
}

type ExpectedGetLatestValueArgs struct {
	ContractName, ReadName string
	ConfidenceLevel        primitives.ConfidenceLevel
	Params, ReturnVal      any
}

type PrimitiveArgs struct {
	Value uint64
}

func (e ExpectedGetLatestValueArgs) String() string {
	return fmt.Sprintf("ContractName: %s, ReadName: %s, ConfidenceLevel: %s, Params: %v, ReturnVal: %v",
		e.ContractName, e.ReadName, e.ConfidenceLevel, e.Params, e.ReturnVal)
}

type InnerDynamicTestStruct struct {
	I int
	S string
}

type InnerStaticTestStruct struct {
	I int
	A  []byte
}

type MidLevelDynamicTestStruct struct {
	FixedBytes [2]byte
	Inner      InnerDynamicTestStruct
}

type MidLevelStaticTestStruct struct {
	FixedBytes [2]byte
	Inner      InnerStaticTestStruct
}

type TestStruct struct {
	Field               *int32
	OracleID            commontypes.OracleID
	OracleIDs           [32]commontypes.OracleID
	Account             []byte
	Accounts            [][]byte
	DifferentField      string
	BigField            *big.Int
	NestedDynamicStruct MidLevelDynamicTestStruct
	NestedStaticStruct  MidLevelStaticTestStruct
}

type TestStructWithExtraField struct {
	TestStruct
	ExtraField int
}

type TestStructMissingField struct {
	DifferentField      string
	OracleID            commontypes.OracleID
	OracleIDs           [32]commontypes.OracleID
	Account             []byte
	Accounts            [][]byte
	BigField            *big.Int
	NestedDynamicStruct MidLevelDynamicTestStruct
	NestedStaticStruct  MidLevelStaticTestStruct
}

// compatibleTestStruct has fields in a different order
type compatibleTestStruct struct {
	Account             []byte
	Accounts            [][]byte
	BigField            *big.Int
	DifferentField      string
	Field               int32
	NestedDynamicStruct MidLevelDynamicTestStruct
	NestedStaticStruct  MidLevelStaticTestStruct
	OracleID            commontypes.OracleID
	OracleIDs           [32]commontypes.OracleID
}

type LatestParams struct {
	// I should be > 0
	I int
}

type FilterEventParams struct {
	Field int32
}

type BatchCallEntry map[types.BoundContract]ContractBatchEntry
type ContractBatchEntry []ReadEntry
type ReadEntry struct {
	Name        string
	ReturnValue any
}

func CreateTestStruct[T any](i int, tester BasicTester[T]) TestStruct {
	s := fmt.Sprintf("field%v", i)
	fv := int32(i)
	return TestStruct{
		Field:          &fv,
		OracleID:       commontypes.OracleID(i + 1),
		OracleIDs:      [32]commontypes.OracleID{commontypes.OracleID(i + 2), commontypes.OracleID(i + 3)},
		Account:        tester.GetAccountBytes(i + 3),
		Accounts:       [][]byte{tester.GetAccountBytes(i + 4), tester.GetAccountBytes(i + 5)},
		DifferentField: s,
		BigField:       big.NewInt(int64((i + 1) * (i + 2))),
		NestedDynamicStruct: MidLevelDynamicTestStruct{
			FixedBytes: [2]byte{uint8(i), uint8(i + 1)},
			Inner: InnerDynamicTestStruct{
				I: i,
				S: s,
			},
		},
		NestedStaticStruct: MidLevelStaticTestStruct{
			FixedBytes: [2]byte{uint8(i), uint8(i + 1)},
			Inner: InnerStaticTestStruct{
				I: i,
				A:  tester.GetAccountBytes(i + 6),
			},
		},
	}
}

func Compare[T cmp.Ordered](a, b T, op primitives.ComparisonOperator) bool {
	switch op {
	case primitives.Eq:
		return a == b
	case primitives.Neq:
		return a != b
	case primitives.Gt:
		return a > b
	case primitives.Lt:
		return a < b
	case primitives.Gte:
		return a >= b
	case primitives.Lte:
		return a <= b
	}
	return false
}
