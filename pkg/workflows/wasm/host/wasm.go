package host

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/bytecodealliance/wasmtime-go/v23"
	"github.com/google/uuid"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
	wasmpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
)

func safeMem(caller *wasmtime.Caller, ptr unsafe.Pointer, size int32) ([]byte, error) {
	mem := caller.GetExport("memory").Memory()
	data := mem.UnsafeData(caller)
	iptr := int32(uintptr(ptr))
	if iptr+size > int32(len(data)) {
		return nil, errors.New("out of bounds memory access")
	}

	cd := make([]byte, size)
	copy(cd, data[iptr:iptr+size])
	return cd, nil
}

var (
	defaultTickInterval = time.Duration(100 * time.Millisecond)
	defaultTimeout      = time.Duration(300 * time.Millisecond)
)

type ModuleConfig struct {
	TickInterval time.Duration
	Timeout      *time.Duration
}

type Module struct {
	engine *wasmtime.Engine
	module *wasmtime.Module
	linker *wasmtime.Linker

	cfg ModuleConfig

	wg     sync.WaitGroup
	stopCh chan struct{}

	response *wasmpb.Response
	err      error
}

func NewModule(modCfg ModuleConfig, binary []byte) (*Module, error) {
	if modCfg.TickInterval == 0 {
		modCfg.TickInterval = defaultTickInterval
	}

	if modCfg.Timeout == nil {
		modCfg.Timeout = &defaultTimeout
	}

	cfg := wasmtime.NewConfig()
	cfg.SetEpochInterruption(true)

	engine := wasmtime.NewEngineWithConfig(cfg)

	mod, err := wasmtime.NewModule(engine, binary)
	if err != nil {
		return nil, fmt.Errorf("error creating wasmtime module: %w", err)
	}

	linker := wasmtime.NewLinker(engine)

	err = linker.DefineWasi()
	if err != nil {
		return nil, err
	}

	var (
		resp     wasmpb.Response
		outerErr error
	)
	err = linker.FuncWrap(
		"env",
		"sendResponse",
		func(caller *wasmtime.Caller, ptr int32, ptrlen int32) {
			b, innerErr := safeMem(caller, unsafe.Pointer(uintptr(ptr)), ptrlen)
			if err != nil {
				outerErr = innerErr
				return
			}

			innerErr = proto.Unmarshal(b, &resp)
			if err != nil {
				outerErr = innerErr
				return
			}
		},
	)
	if err != nil {
		return nil, err
	}

	m := &Module{
		engine: engine,
		module: mod,
		linker: linker,

		cfg: modCfg,

		stopCh: make(chan struct{}),

		response: &resp,
		err:      outerErr,
	}

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()

		ticker := time.NewTicker(modCfg.TickInterval)
		for {
			select {
			case <-m.stopCh:
				return
			case <-ticker.C:
				engine.IncrementEpoch()
			}
		}
	}()

	return m, nil
}

func (m *Module) Close() {
	close(m.stopCh)
	m.wg.Wait()

	m.linker.Close()
	m.engine.Close()
	m.module.Close()
}

func (m *Module) Run(request *wasmpb.Request) (*wasmpb.Response, error) {
	store := wasmtime.NewStore(m.engine)
	defer store.Close()

	reqpb, err := proto.Marshal(request)
	if err != nil {
		return nil, err
	}

	reqstr := base64.StdEncoding.EncodeToString(reqpb)

	wasi := wasmtime.NewWasiConfig()
	wasi.InheritStdout()
	wasi.InheritStderr()
	wasi.SetArgv([]string{"wasi", reqstr})

	store.SetWasi(wasi)

	// Limit memory to 64 megabytes per instance.
	store.Limiter(
		1000*1000*64,
		-1,
		1,
		1,
		1,
	)

	deadline := *m.cfg.Timeout / m.cfg.TickInterval
	store.SetEpochDeadline(uint64(deadline))

	instance, err := m.linker.Instantiate(store, m.module)
	if err != nil {
		return nil, err
	}

	start := instance.GetFunc(store, "_start")
	if start == nil {
		return nil, errors.New("could not get start function")
	}

	_, err = start.Call(store)
	switch {
	case containsCode(err, 0):
		return m.response, m.err
	case containsCode(err, 110):
		return nil, fmt.Errorf("error marshaling response: %s: %w", m.response.ErrMsg, err)
	case containsCode(err, 111):
		return nil, fmt.Errorf("error executing runner: %s: %w", m.response.ErrMsg, err)
	default:
		return nil, err
	}
}

func containsCode(err error, code int) bool {
	return strings.Contains(err.Error(), fmt.Sprintf("exit status %d", code))
}

func GetWorkflowSpec(ctx context.Context, modCfg ModuleConfig, binary []byte, config []byte) (*sdk.WorkflowSpec, error) {
	m, err := NewModule(modCfg, binary)
	if err != nil {
		return nil, fmt.Errorf("could not instantiate module: %w", err)
	}

	rid := uuid.New().String()
	req := &wasmpb.Request{
		Id:     rid,
		Config: config,
		Message: &wasmpb.Request_SpecRequest{
			SpecRequest: &emptypb.Empty{},
		},
	}
	resp, err := m.Run(req)
	if err != nil {
		return nil, err
	}

	sr := resp.GetSpecResponse()
	if sr == nil {
		return nil, errors.New("unexpected response from WASM binary: got nil spec response")
	}

	m.Close()

	return wasmpb.ProtoToWorkflowSpec(sr)
}
