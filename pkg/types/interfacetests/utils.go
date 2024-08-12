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

func batchChainWrite[T TestingT[T]](t T, tester ChainReaderInterfaceTester[T], batchCallEntry BatchCallEntry) {
	nameToAddress := make(map[string]string)
	boundContracts := tester.GetBindings(t)
	for _, bc := range boundContracts {
		nameToAddress[bc.Name] = bc.Address
	}

	for contractName, contractBatch := range batchCallEntry {
		require.Contains(t, nameToAddress, contractName)
		for _, readEntry := range contractBatch {
			val, isOk := readEntry.ReturnValue.(*TestStruct)
			if !isOk {
				require.Fail(t, "expected *TestStruct for contract: %s read: %s, but received %T", contractName, readEntry.Name, readEntry.ReturnValue)
			}
			// it.sendTxWithTestStruct(t, nameToAddress[contractName], val, (*chain_reader_tester.ChainReaderTesterTransactor).AddTestStruct)
			SubmitTransactionToCW(t, tester, "addTestStruct", val, types.BoundContract{Name: contractName, Address: nameToAddress[contractName]}, types.Unconfirmed)
		}
	}
}

func SubmitTransactionToCW[T TestingT[T]](t T, tester ChainReaderInterfaceTester[T], method string, args any, contract types.BoundContract, status types.TransactionStatus) {
	txID := uuid.New().String()
	cw := tester.GetChainWriter(t)
	err := cw.SubmitTransaction(tests.Context(t), contract.Name, method, args, txID, contract.Address, nil, big.NewInt(0))
	require.NoError(t, err)

	waitForTransactionStatus(t, tester, txID, status)
}

func waitForTransactionStatus[T TestingT[T]](t T, tester ChainReaderInterfaceTester[T], txID string, status types.TransactionStatus) error {
	ctx, cancel := context.WithTimeout(tests.Context(t), 5*time.Minute)
	defer cancel()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("transaction %s not finalized within timeout period", txID)
		case <-ticker.C:
			current, err := tester.GetChainWriter(t).GetTransactionStatus(ctx, txID)
			if err != nil {
				return fmt.Errorf("failed to get transaction status: %w", err)
			}

			if current == types.Failed || current == types.Fatal {
				return fmt.Errorf("transaction %s has failed or is fatal", txID)
			} else if current >= status {
				fmt.Printf("Transaction %s reached status: %d\n", txID, current)
				return nil
			} else {
				fmt.Printf("Transaction %s is still %d\n", txID, current)
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
