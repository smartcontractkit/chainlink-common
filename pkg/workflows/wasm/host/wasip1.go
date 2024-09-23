package host

import (
	"encoding/binary"
	"time"

	"github.com/bytecodealliance/wasmtime-go/v23"
	"github.com/jonboulle/clockwork"
)

var (
	nanoBase = time.Now()
	clock    = clockwork.NewFakeClockAt(nanoBase)
	tick     = 100 * time.Millisecond
)

func newWasiLinker(engine *wasmtime.Engine) (*wasmtime.Linker, error) {
	linker := wasmtime.NewLinker(engine)
	linker.AllowShadowing(true)

	err := linker.DefineWasi()
	if err != nil {
		return nil, err
	}

	err = linker.FuncWrap(
		"wasi_snapshot_preview1",
		"poll_oneoff",
		pollOneoff,
	)
	if err != nil {
		return nil, err
	}

	err = linker.FuncWrap(
		"wasi_snapshot_preview1",
		"clock_time_get",
		clockTimeGet,
	)
	if err != nil {
		return nil, err
	}

	return linker, nil
}

const (
	clockIDRealtime = iota
	clockIDMonotonic
)

// Loosely based off the implementation here:
// https://github.com/tetratelabs/wazero/blob/main/imports/wasi_snapshot_preview1/clock.go#L42
// Each call to clockTimeGet increments our fake clock by `tick`.
func clockTimeGet(caller *wasmtime.Caller, id int32, precision int64, resultTimestamp int32) int32 {
	var val int64
	switch id {
	case clockIDMonotonic:
		clock.Advance(tick)
		val = clock.Since(nanoBase).Nanoseconds()
	case clockIDRealtime:
		clock.Advance(tick)
		val = clock.Now().UnixNano()
	default:
		return ErrnoInval
	}

	uint64Size := int32(8)
	trg := make([]byte, uint64Size)
	binary.LittleEndian.PutUint64(trg, uint64(val))
	copyBuffer(caller, trg, resultTimestamp, uint64Size)
	return ErrnoSuccess
}

const (
	subscriptionLen = 48
	eventsLen       = 32

	eventTypeClock = iota
	eventTypeFDRead
	eventTypeFDWrite
)

// Loosely based off the implementation here:
// https://github.com/tetratelabs/wazero/blob/main/imports/wasi_snapshot_preview1/poll.go#L52
// For an overview of the spec, including the datatypes being referred to, see:
// https://github.com/WebAssembly/WASI/blob/snapshot-01/phases/snapshot/docs.md
// This implementation only responds to clock events, not to file descriptor notifications.
// It doesn't actually sleep though, and will instead advance our fake clock by the sleep duration.
func pollOneoff(caller *wasmtime.Caller, subscriptionptr int32, eventsptr int32, nsubscriptions int32, resultNevents int32) int32 {
	if nsubscriptions == 0 {
		return ErrnoInval
	}

	subs, err := safeMem(caller, subscriptionptr, nsubscriptions*subscriptionLen)
	if err != nil {
		return ErrnoFault
	}

	// Each subscription should have an event
	events := make([]byte, nsubscriptions*eventsLen)

	timeout := time.Duration(0)
	for i := int32(0); i < nsubscriptions; i++ {
		// First, let's read the subscription
		inOffset := i * subscriptionLen

		userData := subs[inOffset : inOffset+8]
		eventType := subs[inOffset+8]
		argBuf := subs[inOffset+8+8:]

		outOffset := events[i*eventsLen]

		slot := events[outOffset:]
		switch eventType {
		case eventTypeClock:
			// We want to stub out clock events,
			// so let's just return success, and
			// we'll advance the clock by the timeout duration
			// below.

			// Structure of event, per:
			// https://github.com/WebAssembly/WASI/blob/snapshot-01/phases/snapshot/docs.md#-subscription_clock-struct
			// - 0-8: clock id
			// - 8-16: timeout
			// - 16-24: precision
			// - 24-32: flag
			newTimeout := binary.LittleEndian.Uint16(argBuf[8:16])
			flag := binary.LittleEndian.Uint16(argBuf[24:32])

			var errno Errno
			switch flag {
			case 0: // relative time
				errno = ErrnoSuccess
				if timeout < time.Duration(newTimeout) {
					timeout = time.Duration(newTimeout)
				}
			default:
				errno = ErrnoNotsup
			}
			writeEvent(slot, userData, errno, eventTypeClock)
		case eventTypeFDRead:
			// Our sandbox doesn't allow access to the filesystem,
			// so let's just error these events
			writeEvent(slot, userData, ErrnoBadf, eventTypeFDRead)
		case eventTypeFDWrite:
			// Our sandbox doesn't allow access to the filesystem,
			// so let's just error these events
			writeEvent(slot, userData, ErrnoBadf, eventTypeFDWrite)
		default:
			writeEvent(slot, userData, ErrnoInval, int(eventType))
		}
	}

	// Advance the clock by timeout.
	// This will make it seem like we've slept by timeout.
	if timeout > 0 {
		clock.Advance(timeout)
	}

	uint32Size := int32(4)
	rne := make([]byte, uint32Size)
	binary.LittleEndian.PutUint32(rne, uint32(nsubscriptions))

	// Write the number of events to `resultNevents`
	size := copyBuffer(caller, rne, resultNevents, uint32Size)
	if size == -1 {
		return ErrnoFault
	}

	// Write the events to `events`
	size = copyBuffer(caller, events, eventsptr, nsubscriptions*eventsLen)
	if size == -1 {
		return ErrnoFault
	}

	return ErrnoSuccess
}

func writeEvent(slot []byte, userData []byte, errno Errno, eventType int) {
	// the event structure is described here:
	// https://github.com/WebAssembly/WASI/blob/snapshot-01/phases/snapshot/docs.md#-event-struct
	copy(slot, userData)
	slot[8] = byte(errno)
	slot[9] = 0
	binary.LittleEndian.PutUint32(slot[10:], uint32(eventType))
}
