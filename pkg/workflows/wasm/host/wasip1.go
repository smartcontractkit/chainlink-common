package host

import (
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"time"

	"github.com/jonboulle/clockwork"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/engine"
)

var (
	nanoBase = time.Now()
	clock    = clockwork.NewFakeClockAt(nanoBase)
	tick     = 100 * time.Millisecond
)

const (
	clockIDRealtime = iota
	clockIDMonotonic
)

const (
	subscriptionLen = 48
	eventsLen       = 32

	eventTypeClock = iota
	eventTypeFDRead
	eventTypeFDWrite
)

// legacyClockTimeGet is the fake-clock implementation used by legacy DAG workflows.
func legacyClockTimeGet(caller engine.MemoryAccessor, id int32, precision int64, resultTimestamp int32) int32 {
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
	wasmWrite(caller, trg, resultTimestamp, uint64Size)
	return ErrnoSuccess
}

// legacyPollOneoff is the fake-clock implementation used by legacy DAG workflows.
// It advances the fake clock by the sleep duration rather than actually sleeping.
func legacyPollOneoff(caller engine.MemoryAccessor, subscriptionptr int32, eventsptr int32, nsubscriptions int32, resultNevents int32) int32 {
	if nsubscriptions == 0 {
		return ErrnoInval
	}

	subs, err := wasmRead(caller, subscriptionptr, nsubscriptions*subscriptionLen)
	if err != nil {
		return ErrnoFault
	}

	events := make([]byte, nsubscriptions*eventsLen)

	timeout := time.Duration(0)
	for i := range nsubscriptions {
		inOffset := i * subscriptionLen
		userData := subs[inOffset : inOffset+8]
		eventType := subs[inOffset+8]
		argBuf := subs[inOffset+8+8:]

		slot, serr := getSlot(events, i)
		if serr != nil {
			return ErrnoFault
		}

		switch eventType {
		case eventTypeClock:
			newTimeout := binary.LittleEndian.Uint64(argBuf[8:16])
			flag := binary.LittleEndian.Uint16(argBuf[24:32])

			var errno Errno
			switch flag {
			case 0:
				errno = ErrnoSuccess
				if timeout < time.Duration(newTimeout) {
					timeout = time.Duration(newTimeout)
				}
			default:
				errno = ErrnoNotsup
			}
			writeEvent(slot, userData, errno, eventTypeClock)
		case eventTypeFDRead:
			writeEvent(slot, userData, ErrnoBadf, eventTypeFDRead)
		case eventTypeFDWrite:
			writeEvent(slot, userData, ErrnoBadf, eventTypeFDWrite)
		default:
			writeEvent(slot, userData, ErrnoInval, int(eventType))
		}
	}

	if timeout > 0 {
		clock.Advance(timeout)
	}

	uint32Size := int32(4)
	rne := make([]byte, uint32Size)
	binary.LittleEndian.PutUint32(rne, uint32(nsubscriptions))

	if wasmWrite(caller, rne, resultNevents, uint32Size) == -1 {
		return ErrnoFault
	}
	if wasmWrite(caller, events, eventsptr, nsubscriptions*eventsLen) == -1 {
		return ErrnoFault
	}

	return ErrnoSuccess
}

func writeEvent(slot []byte, userData []byte, errno Errno, eventType int) {
	copy(slot, userData)
	slot[8] = byte(errno)
	slot[9] = 0
	binary.LittleEndian.PutUint32(slot[10:], uint32(eventType))
}

func createRandomGet(cfg *ModuleConfig) func(caller engine.MemoryAccessor, buf, bufLen int32) int32 {
	return func(caller engine.MemoryAccessor, buf, bufLen int32) int32 {
		if cfg == nil || cfg.Determinism == nil {
			return ErrnoInval
		}

		var (
			seed       = cfg.Determinism.Seed
			randSource = rand.New(rand.NewSource(seed)) //nolint:gosec
			randOutput = make([]byte, bufLen)
		)

		if _, err := io.ReadAtLeast(randSource, randOutput, int(bufLen)); err != nil {
			return ErrnoFault
		}

		if n := wasmWrite(caller, randOutput, buf, bufLen); n != int64(len(randOutput)) {
			return ErrnoFault
		}

		return ErrnoSuccess
	}
}

func getSlot(events []byte, i int32) ([]byte, error) {
	offset := i * eventsLen

	if offset+eventsLen > int32(len(events)) {
		return nil, fmt.Errorf("slot %d out of bounds", i)
	}

	slot := events[offset : offset+eventsLen]
	return slot, nil
}
