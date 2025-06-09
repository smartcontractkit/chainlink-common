//go:build !wasip1

package wasm

import (
	"testing"
	"unsafe"
)

type runnerInternalsTestHook struct {
	testTb       testing.TB
	execId       string
	arguments    []string
	sentResponse []byte
	modeSwitched bool
	mode         int32
}

func (r *runnerInternalsTestHook) args() []string {
	return r.arguments
}

func (r *runnerInternalsTestHook) sendResponse(response unsafe.Pointer, responseLen int32) int32 {
	r.sentResponse = unsafe.Slice((*byte)(response), responseLen)
	return 0
}

func (r *runnerInternalsTestHook) versionV2() {}

func (r *runnerInternalsTestHook) switchModes(mode int32) {
	r.mode = mode
	r.modeSwitched = true
}

var _ runnerInternals = (*runnerInternalsTestHook)(nil)
