package engine

import "fmt"

type EngineType string

const (
	EngineWasmtime EngineType = "wasmtime"
	EngineWasm3    EngineType = "wasm3"
)

type MemoryAccessor interface {
	Memory() []byte
}

type LoadConfig struct {
	InitialFuel uint64
}

type Engine interface {
	Load(binary []byte, cfg LoadConfig) (Runtime, error)
}

type Runtime interface {
	V2ImportName() string
	NewStore() Store
	IncrementEpoch()
	Close()
}

type Store interface {
	SetWasi(argv []string)
	SetFuel(fuel uint64) error
	SetLimiter(memoryBytes, tableElements, instances, tables, memories int64)
	SetEpochDeadline(deadline uint64)
	LinkNoDAG(exec *Execution) (Instance, error)
	LinkLegacyDAG(exec *LegacyExecution) (Instance, error)
	Close()
}

type Instance interface {
	CallStart() error
}

type Execution struct {
	V2ImportName      string
	SendResponse      func(caller MemoryAccessor, ptr, ptrlen int32) int32
	CallCapability    func(caller MemoryAccessor, ptr, ptrlen int32) int64
	AwaitCapabilities func(caller MemoryAccessor, awaitReq, awaitReqLen, respBuf, maxRespLen int32) int64
	GetSecrets        func(caller MemoryAccessor, req, reqLen, respBuf, maxRespLen int32) int64
	AwaitSecrets      func(caller MemoryAccessor, awaitReq, awaitReqLen, respBuf, maxRespLen int32) int64
	Log               func(caller MemoryAccessor, ptr, ptrlen int32)
	SwitchModes       func(caller MemoryAccessor, mode int32)
	RandomSeed        func(mode int32) int64
	Now               func(caller MemoryAccessor, resultTimestamp int32) int32
	PollOneoff        func(caller MemoryAccessor, subPtr, evPtr, nSubs, resultNEvents int32) int32
	ClockTimeGet      func(caller MemoryAccessor, id int32, precision int64, resultTs int32) int32
}

type LegacyExecution struct {
	SendResponse func(caller MemoryAccessor, ptr, ptrlen int32) int32
	Fetch        func(caller MemoryAccessor, respPtr, respLenPtr, reqPtr, reqPtrLen int32) int32
	Emit         func(caller MemoryAccessor, respPtr, respLenPtr, msgPtr, msgLen int32) int32
	Log          func(caller MemoryAccessor, ptr, ptrlen int32)
	PollOneoff   func(caller MemoryAccessor, subPtr, evPtr, nSubs, resultNEvents int32) int32
	ClockTimeGet func(caller MemoryAccessor, id int32, precision int64, resultTs int32) int32
	RandomGet    func(caller MemoryAccessor, buf, bufLen int32) int32 // nil when determinism is disabled
}

func NewEngine(engineType EngineType) (Engine, error) {
	switch engineType {
	case EngineWasmtime:
		return &wasmtimeEngine{}, nil
	case EngineWasm3:
		return newWasm3Engine()
	default:
		return nil, fmt.Errorf("unsupported engine type: %q", engineType)
	}
}
