//go:build !wasip1

package wasm

import "unsafe"

var logs [][]byte

func log(message unsafe.Pointer, messageLen int32) {
	currentLog := unsafe.Slice((*byte)(message), messageLen)
	logs = append(logs, currentLog)
}
