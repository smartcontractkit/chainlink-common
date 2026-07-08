package host

import (
	"encoding/binary"
	"fmt"
)

// The host now relies on wazero's standard wasi_snapshot_preview1
// implementation for clock_time_get, poll_oneoff and random_get. The custom
// deterministic-clock / fake-sleep / seeded-random overrides that previously
// lived here have been removed. The pure helpers below are retained because the
// (currently unused) execution.pollOneoff / execution.clockTimeGet methods and
// the unit tests still reference them; they can be rewired into a custom WASI
// host module if deterministic behaviour is reintroduced.

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
