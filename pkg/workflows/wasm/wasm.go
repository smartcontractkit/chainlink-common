package wasm

import (
	"context"
	_ "embed"
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

	"github.com/smartcontractkit/chainlink-common/pkg/workflows"
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

type module struct {
	engine *wasmtime.Engine
	module *wasmtime.Module
	linker *wasmtime.Linker

	wg     sync.WaitGroup
	stopCh chan struct{}

	response *wasmpb.Response
	err      error
}

func newModule(binary []byte) (*module, error) {
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
	linker.FuncWrap(
		"env",
		"sendResponse",
		func(caller *wasmtime.Caller, ptr int32, ptrlen int32) {
			b, err := safeMem(caller, unsafe.Pointer(uintptr(ptr)), ptrlen)
			if err != nil {
				outerErr = err
				return
			}

			err = proto.Unmarshal(b, &resp)
			if err != nil {
				outerErr = err
				return
			}
		},
	)

	m := &module{
		engine: engine,
		module: mod,
		linker: linker,

		stopCh: make(chan struct{}),

		response: &resp,
		err:      outerErr,
	}

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()

		ticker := time.NewTicker(1 * time.Second)
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

func (m *module) close() {
	close(m.stopCh)
	m.wg.Wait()

	m.linker.Close()
	m.engine.Close()
	m.module.Close()
}

func (m *module) run(ctx context.Context, request *wasmpb.Request) (*wasmpb.Response, error) {
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

	// set the deadline to be 5 relative to the current epoch
	// this roughly translates to a timeout of 5 seconds, since the
	// epoch goroutine ticks every second.
	store.SetEpochDeadline(5)

	instance, err := m.linker.Instantiate(store, m.module)
	if err != nil {
		return nil, err
	}

	start := instance.GetFunc(store, "_start")
	if start == nil {
		return nil, errors.New("could not get start function")
	}

	_, err = start.Call(store)
	if err != nil {
		if !strings.Contains(err.Error(), "exit status 0") {
			return nil, err

		}
	}

	return m.response, m.err
}

func GetWorkflowSpec(ctx context.Context, binary []byte, config []byte) (*workflows.WorkflowSpec, error) {
	m, err := newModule(binary)
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
	resp, err := m.run(ctx, req)
	if err != nil {
		return nil, err
	}

	sr := resp.GetSpecResponse()
	if sr == nil {
		return nil, errors.New("unexpected response from WASM binary: got nil spec response")
	}

	m.close()

	return wasmpb.ProtoToWorkflowSpec(sr)
}
