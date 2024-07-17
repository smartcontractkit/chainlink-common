package interfacetests

import (
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

func submitTransactionToCW[T TestingT[T]](t T, tester ChainReaderInterfaceTester[T], args any, contract types.BoundContract) {
	txID := uuid.New().String()
	cw := tester.GetChainWriter(t)
	err := cw.SubmitTransaction(tests.Context(t), contract.Name, "addTestStruct", args, txID, contract.Address, nil, big.NewInt(0))
	tester.IncNonce()
	require.NoError(t, err)

	waitForTransactionFinalization(t, tester, txID)
}

func waitForTransactionFinalization[T TestingT[T]](t T, tester ChainReaderInterfaceTester[T], txID string) error {
	ctx, cancel := context.WithTimeout(tests.Context(t), 5*time.Minute)
	defer cancel()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("transaction %s not finalized within timeout period", txID)
		case <-ticker.C:
			status, err := tester.GetChainWriter(t).GetTransactionStatus(ctx, txID)
			if err != nil {
				return fmt.Errorf("failed to get transaction status: %w", err)
			}

			switch status {
			case types.Finalized:
				fmt.Println("Found successful Transaction")
				return nil
			case types.Failed, types.Fatal:
				return fmt.Errorf("transaction %s has failed or is fatal", txID)
			case types.Unknown, types.Unconfirmed:
				fmt.Printf("Transaction %s is still %d\n", txID, status)
				// Continue polling for these statuses
			}
		}
	}
}

type ExpectedGetLatestValueArgs struct {
	ContractName, ReadName string
	ConfidenceLevel        primitives.ConfidenceLevel
	Params, ReturnVal      any
}

func (e ExpectedGetLatestValueArgs) String() string {
	return fmt.Sprintf("ContractName: %s, ReadName: %s, ConfidenceLevel: %s, Params: %v, ReturnVal: %v",
		e.ContractName, e.ReadName, e.ConfidenceLevel, e.Params, e.ReturnVal)
}

type InnerTestStruct struct {
	I int
	S string
}

type MidLevelTestStruct struct {
	FixedBytes [2]byte
	Inner      InnerTestStruct
}

type TestStruct struct {
	Field          *int32
	DifferentField string
	OracleID       commontypes.OracleID
	OracleIDs      [32]commontypes.OracleID
	Account        []byte
	Accounts       [][]byte
	BigField       *big.Int
	NestedStruct   MidLevelTestStruct
}

type TestStructWithExtraField struct {
	TestStruct
	ExtraField int
}

type TestStructMissingField struct {
	DifferentField string
	OracleID       commontypes.OracleID
	OracleIDs      [32]commontypes.OracleID
	Account        []byte
	Accounts       [][]byte
	BigField       *big.Int
	NestedStruct   MidLevelTestStruct
}

// compatibleTestStruct has fields in a different order
type compatibleTestStruct struct {
	Account        []byte
	Accounts       [][]byte
	BigField       *big.Int
	DifferentField string
	Field          int32
	NestedStruct   MidLevelTestStruct
	OracleID       commontypes.OracleID
	OracleIDs      [32]commontypes.OracleID
}

type LatestParams struct {
	// I should be > 0
	I int
}

type FilterEventParams struct {
	Field int32
}

type BatchCallEntry map[string]ContractBatchEntry
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
		DifferentField: s,
		OracleID:       commontypes.OracleID(i + 1),
		OracleIDs:      [32]commontypes.OracleID{commontypes.OracleID(i + 2), commontypes.OracleID(i + 3)},
		Account:        tester.GetAccountBytes(i + 3),
		Accounts:       [][]byte{tester.GetAccountBytes(i + 4), tester.GetAccountBytes(i + 5)},
		BigField:       big.NewInt(int64((i + 1) * (i + 2))),
		NestedStruct: MidLevelTestStruct{
			FixedBytes: [2]byte{uint8(i), uint8(i + 1)},
			Inner: InnerTestStruct{
				I: i,
				S: s,
			},
		},
	}
}
