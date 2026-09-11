package host

import (
	"encoding/binary"
	"fmt"

	"github.com/bytecodealliance/wasmtime-go/v47"
)

const (
	clockIDRealtime = iota
	clockIDMonotonic
)

const (
	subscriptionLen = 48
	eventsLen       = 32
)

const (
	eventTypeClock = iota
	eventTypeFDRead
	eventTypeFDWrite
)

func newWasiLinker[T any](exec *execution[T], engine *wasmtime.Engine) (*wasmtime.Linker, error) {
	linker := wasmtime.NewLinker(engine)
	cleanupLinker := true
	defer func() {
		if cleanupLinker {
			linker.Close()
		}
	}()
	linker.AllowShadowing(true)

	err := linker.DefineWasi()
	if err != nil {
		return nil, err
	}

	err = linker.FuncWrap(
		"wasi_snapshot_preview1",
		"poll_oneoff",
		exec.pollOneoff,
	)
	if err != nil {
		return nil, err
	}

	exec.timeFetcher = newTimeFetcher(exec.ctx, exec.executor)
	exec.timeFetcher.Start()

	err = linker.FuncWrap(
		"wasi_snapshot_preview1",
		"clock_time_get",
		exec.clockTimeGet,
	)
	if err != nil {
		return nil, err
	}

	cleanupLinker = false
	return linker, nil
}

func writeEvent(slot []byte, userData []byte, errno Errno, eventType int) {
	// the event structure is described here:
	// https://github.com/WebAssembly/WASI/blob/snapshot-01/phases/snapshot/docs.md#-event-struct
	copy(slot, userData)
	slot[8] = byte(errno)
	slot[9] = 0
	binary.LittleEndian.PutUint32(slot[10:], uint32(eventType))
}

func getSlot(events []byte, i int32) ([]byte, error) {
	offset := i * eventsLen

	if offset+eventsLen > int32(len(events)) {
		return nil, fmt.Errorf("slot %d out of bounds", i)
	}

	slot := events[offset : offset+eventsLen]
	return slot, nil
}
